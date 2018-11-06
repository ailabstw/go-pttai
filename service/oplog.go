// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

type Oplog interface {
	GetBaseOplog() *BaseOplog
	Save(isLocked bool) error
	Delete(isLocked bool) error

	SetPreLogID(logID *types.PttID)
	ToStatus() types.Status
}

type BaseOplog struct {
	V         types.Version
	ID        *types.PttID
	CreatorID *types.PttID    `json:"CID"`
	CreateTS  types.Timestamp `json:"CT"`

	ObjID *types.PttID `json:"OID"`

	Op OpType `json:"O"`

	PreLogID *types.PttID `json:"p,omitempty"`

	Data OpData `json:"D,omitempty"`

	db               *pttdb.LDBBatch
	dbPrefixID       *types.PttID
	dbPrefix         []byte
	dbIdxPrefix      []byte
	dbMerklePrefix   []byte
	dbPrefixInternal []byte
	dbPrefixMaster   []byte

	CreatorHash []byte        `json:"cH,omitempty"`
	Salt        types.Salt    `json:"s,omitempty"`
	Sig         []byte        `json:"S,omitempty"`
	Pubkey      []byte        `json:"K,omitempty"`
	KeyExtra    *KeyExtraInfo `json:"k,omitempty"`

	// to remove when doing sign
	UpdateTS types.Timestamp `json:"UT"`
	Hash     []byte          `json:"H,omitempty"`

	MasterLogID   *types.PttID `json:"mID,omitempty"`
	MasterSigns   []*SignInfo  `json:"m,omitempty"`
	InternalSigns []*SignInfo  `json:"i,omitempty"`

	dbLock *types.LockMap

	IsSync  types.Bool  `json:"y"`           // not distribute
	IsNewer types.Bool  `json:"n,omitempty"` // for p2p, should be empty in save / sign
	Extra   interface{} `json:"e,omitempty"`
}

func NewOplogForLoadData(data interface{}, db *pttdb.LDBBatch) *BaseOplog {
	return &BaseOplog{Data: data, db: db}
}

func NewOplog(id *types.PttID, ts types.Timestamp, doerID *types.PttID, op OpType, data interface{}, db *pttdb.LDBBatch, dbPrefixID *types.PttID, dbPrefix []byte, dbIdxPrefix []byte, dbMerklePrefix []byte, dbLock *types.LockMap) (*BaseOplog, error) {

	opID, err := types.NewPttID()
	if err != nil {
		return nil, err
	}

	dbPrefixInternal := dbPrefixToDBPrefixInternal(dbPrefix)
	dbPrefixMaster := dbPrefixToDBPrefixMaster(dbPrefix)

	oplog := &BaseOplog{
		V:         types.CurrentVersion,
		ID:        opID,
		CreatorID: doerID,
		CreateTS:  ts,
		UpdateTS:  ts,

		ObjID: id,

		Op: op,

		Data: data,

		db:               db,
		dbPrefixID:       dbPrefixID,
		dbPrefix:         dbPrefix,
		dbPrefixInternal: dbPrefixInternal,
		dbPrefixMaster:   dbPrefixMaster,
		dbIdxPrefix:      dbIdxPrefix,
		dbMerklePrefix:   dbMerklePrefix,
		dbLock:           dbLock,

		IsSync: true,
	}

	return oplog, nil
}

func (o *BaseOplog) SetDB(db *pttdb.LDBBatch, id *types.PttID, prefix []byte, idxPrefix []byte, merklePrefix []byte, dbLock *types.LockMap) {
	dbPrefixInternal := dbPrefixToDBPrefixInternal(prefix)
	dbPrefixMaster := dbPrefixToDBPrefixMaster(prefix)

	o.db = db
	o.dbPrefixID = id
	o.dbPrefix = prefix
	o.dbPrefixInternal = dbPrefixInternal
	o.dbPrefixMaster = dbPrefixMaster
	o.dbIdxPrefix = idxPrefix
	o.dbMerklePrefix = merklePrefix
	o.dbLock = dbLock
}

func (o *BaseOplog) GetDB() *pttdb.LDBBatch {
	return o.db
}

func (o *BaseOplog) GetDBPrefxiID() *types.PttID {
	return o.dbPrefixID
}

func (o *BaseOplog) GetDBPrefix() []byte {
	return o.dbPrefix
}

func (o *BaseOplog) GetDBPrefixInternal() []byte {
	return o.dbPrefixInternal
}

func (o *BaseOplog) GetDBPrefixMaster() []byte {
	return o.dbPrefixMaster
}

func (o *BaseOplog) GetDBIdxPrefix() []byte {
	return o.dbIdxPrefix
}

func (o *BaseOplog) GetDBMerklePrefix() []byte {
	return o.dbMerklePrefix
}

func (o *BaseOplog) GetDBLock() *types.LockMap {
	return o.dbLock
}

func (o *BaseOplog) SaveWithIsSync(isLocked bool) error {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	idxKey, idx, kvs, err := o.SaveCore()
	if err != nil {
		return err
	}

	origO := &BaseOplog{}
	origO.SetDB(o.db, o.dbPrefixID, o.dbPrefix, o.dbIdxPrefix, o.dbMerklePrefix, o.dbLock)
	err = origO.Load(kvs[0].K)
	if err == nil && reflect.DeepEqual(o.Hash, origO.Hash) && bool(origO.IsSync) {
		o.IsSync = true
		return nil
	}

	_, err = o.db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (o *BaseOplog) Save(isLocked bool) error {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	idxKey, idx, kvs, err := o.SaveCore()
	if err != nil {
		return err
	}

	_, err = o.db.TryPutAll(idxKey, idx, kvs, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (o *BaseOplog) ForceSave(isLocked bool) error {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	idxKey, idx, kvs, err := o.SaveCore()
	if err != nil {
		return err
	}

	// XXX need to do verify before save
	_, err = o.db.ForcePutAll(idxKey, idx, kvs)
	if err != nil {
		return err
	}

	return nil
}

func (o *BaseOplog) SaveCore() ([]byte, *pttdb.Index, []*pttdb.KeyVal, error) {
	o.IsNewer = false
	if o.MasterLogID == nil {
		return o.SaveCorePending()
	}

	key, err := o.MarshalKey(o.dbPrefix)
	if err != nil {
		return nil, nil, nil, err
	}

	marshaled, err := o.Marshal()
	if err != nil {
		return nil, nil, nil, err
	}

	idxKey, err := o.IdxKey()
	if err != nil {
		return nil, nil, nil, err
	}

	keys := make([][]byte, 1, 2)
	keys[0] = key
	kvs := make([]*pttdb.KeyVal, 1, 2)
	kvs[0] = &pttdb.KeyVal{K: key, V: marshaled}

	if o.IsSync && o.dbMerklePrefix != nil {
		addr := types.HashToAddr(o.Hash)
		merkleNode := &MerkleNode{
			Level:     MerkleTreeLevelNow,
			Addr:      addr,
			UpdateTS:  o.CreateTS,
			NChildren: 0,
			Key:       key,
		}
		marshaledNode, err := merkleNode.Marshal()
		if err != nil {
			return nil, nil, nil, err
		}
		marshaledMerkleKey, err := o.MarshalMerkleKey()
		if err != nil {
			return nil, nil, nil, err
		}

		keys = append(keys, marshaledMerkleKey)
		kvs = append(kvs, &pttdb.KeyVal{K: marshaledMerkleKey, V: marshaledNode})
	}

	idx := &pttdb.Index{Keys: keys, UpdateTS: o.UpdateTS}

	// dealing with oplog with master-log-id.
	// delete orig-log if the status of the orig-log is not with master-log-id
	origKey, err := o.db.GetKeyByIdxKey(idxKey, 0)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, nil, nil, err
	}
	if err == nil {
		origStatus := bytesToStatus(origKey)
		if origStatus != types.StatusAlive {
			o.db.DeleteAll(idxKey)
		}
	}

	return idxKey, idx, kvs, nil
}

func (o *BaseOplog) SaveCorePending() ([]byte, *pttdb.Index, []*pttdb.KeyVal, error) {
	idxKey, err := o.IdxKey()
	if err != nil {
		return nil, nil, nil, err
	}

	origKey, err := o.db.GetKeyByIdxKey(idxKey, 0)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, nil, nil, err
	}

	currentStatus := o.ToStatus()
	if err == nil {
		origStatus := bytesToStatus(origKey)
		if currentStatus < origStatus {
			return nil, nil, nil, ErrInvalidOplog
		} else if currentStatus > origStatus {
			o.db.DeleteAll(idxKey)
		}
	}

	var key []byte
	switch currentStatus {
	case types.StatusInternalPending:
		key, err = o.MarshalKey(o.dbPrefixInternal)
	case types.StatusPending:
		key, err = o.MarshalKey(o.dbPrefixMaster)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	marshaled, err := o.Marshal()
	if err != nil {
		return nil, nil, nil, err
	}

	idx := &pttdb.Index{Keys: [][]byte{key}, UpdateTS: o.UpdateTS}

	kvs := []*pttdb.KeyVal{
		&pttdb.KeyVal{
			K: key,
			V: marshaled,
		},
	}

	return idxKey, idx, kvs, nil
}

func (o *BaseOplog) Get(id *types.PttID, isLocked bool) error {
	if !isLocked {
		err := o.dbLock.RLock(id)
		if err != nil {
			return err
		}
		defer o.dbLock.RUnlock(id)
	}

	key, err := o.GetKey(id, true)
	if err != nil {
		return err
	}

	return o.Load(key)
}

func (o *BaseOplog) GetKey(id *types.PttID, isLocked bool) ([]byte, error) {
	if !isLocked {
		err := o.dbLock.RLock(id)
		if err != nil {
			return nil, err
		}
		defer o.dbLock.RUnlock(id)
	}

	o.ID = id
	idxKey, err := o.IdxKey()
	if err != nil {
		return nil, err
	}

	return o.db.GetKeyByIdxKey(idxKey, 0)
}

func (o *BaseOplog) Delete(isLocked bool) error {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	idxKey, err := o.IdxKey()
	if err != nil {
		return err
	}

	return o.db.DeleteAll(idxKey)
}

func (o *BaseOplog) Load(key []byte) error {
	/*
		if !isLocked {
			log.Debug("Load", "dbLock", o.dbLock, "key", key, "o", o.ID)
			err := o.dbLock.RLock(o.ID)
			if err != nil {
				return err
			}
			defer o.dbLock.RUnlock(o.ID)
		}
	*/

	marshaled, err := o.db.DBGet(key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshaled, o)
	if err != nil {
		log.Error("Oplog.Load: unable to Unmarshal", "marshaled", marshaled, "o", o)
		return err
	}

	return nil
}

/*
IdxPrefix
*/
func (o *BaseOplog) IdxPrefix() []byte {
	return append(o.dbIdxPrefix, o.dbPrefixID[:]...)
}

/*
IdxKey: idxPrefix:OplogID
*/
func (o *BaseOplog) IdxKey() ([]byte, error) {
	return common.Concat([][]byte{o.dbIdxPrefix, o.dbPrefixID[:], o.ID[:]})
}

/*
Prefix
*/
func (o *BaseOplog) DBPrefix() []byte {
	return append(o.dbPrefix, o.dbPrefixID[:]...)
}

/*
PrefixInternal
*/
func (o *BaseOplog) DBPrefixInternal() []byte {
	return append(o.dbPrefixInternal, o.dbPrefixID[:]...)
}

/*
PrefixInternal
*/
func (o *BaseOplog) DBPrefixMaster() []byte {
	return append(o.dbPrefixMaster, o.dbPrefixID[:]...)
}

/*
MarshalKey: prefixID:TS:OplogID:Op
*/
func (o *BaseOplog) MarshalMerkleKey() ([]byte, error) {
	marshaledTS, err := o.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}

	opBytes, err := MarshalOp(o.Op)
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{o.dbMerklePrefix, o.dbPrefixID[:], []byte{uint8(MerkleTreeLevelNow)}, marshaledTS, o.ID[:], opBytes})
}

/*
MarshalKey: prefixID:TS:OplogID:Op
*/
func (o *BaseOplog) MarshalKey(prefix []byte) ([]byte, error) {
	marshaledTS, err := o.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}

	opBytes, err := MarshalOp(o.Op)
	if err != nil {
		return nil, err
	}

	return common.Concat([][]byte{prefix, o.dbPrefixID[:], marshaledTS[:], o.ID[:], opBytes})
}

func (o *BaseOplog) Marshal() ([]byte, error) {
	return json.Marshal(o)
}

func (o *BaseOplog) Unmarshal(data []byte) error {
	return json.Unmarshal(data, o)
}

func (o *BaseOplog) GetData(data interface{}) error {
	marshaled, err := json.Marshal(o.Data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(marshaled, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *BaseOplog) Sign(keyInfo *KeyInfo) error {
	origSync := o.IsSync
	origExtra := o.Extra
	defer func() {
		o.IsSync = origSync
		o.Extra = origExtra
	}()

	o.UpdateTS = types.ZeroTimestamp
	o.Hash = nil

	o.CreatorHash = nil
	o.Sig = nil
	o.Salt = types.Salt{}
	o.Pubkey = nil

	o.MasterLogID = nil
	o.MasterSigns = nil
	o.InternalSigns = nil

	o.IsSync = false
	o.IsNewer = false
	o.Extra = nil

	marshaled, err := o.Marshal()
	if err != nil {
		return err
	}

	bytesWithSalt, hash, sig, pubBytes, err := SignData(marshaled, keyInfo)
	if err != nil {
		return err
	}

	o.CreatorHash = hash
	copy(o.Salt[:], bytesWithSalt[len(marshaled):])
	o.Sig = sig
	o.Pubkey = pubBytes
	o.KeyExtra = keyInfo.Extra

	o.UpdateTS = o.CreateTS
	o.Hash, _ = o.SignsHash()

	return nil
}

func (o *BaseOplog) SignsHash() ([]byte, error) {

	// master-logid + doer
	lenData := 1 + (len(o.MasterSigns)+len(o.InternalSigns)+1)*4

	theBytes := make([][]byte, lenData)

	// master-logid
	if o.MasterLogID != nil {
		theBytes[0] = o.MasterLogID[:]
	} else {
		theBytes[0] = []byte{}
	}

	offset := 1

	// doer
	theBytes[offset], theBytes[offset+1], theBytes[offset+2], theBytes[offset+3] = o.CreatorHash, o.Salt[:], o.Sig, o.Pubkey
	offset += 4

	// already sorted by id
	// masters
	for _, eachSign := range o.MasterSigns {
		theBytes[offset], theBytes[offset+1], theBytes[offset+2], theBytes[offset+3] = eachSign.Hash, eachSign.Salt[:], eachSign.Sig, eachSign.Pubkey
		offset += 4
	}

	// already sorted by id
	// internals
	for _, eachSign := range o.InternalSigns {
		theBytes[offset], theBytes[offset+1], theBytes[offset+2], theBytes[offset+3] = eachSign.Hash, eachSign.Salt[:], eachSign.Sig, eachSign.Pubkey
		offset += 4
	}

	concatBytes, err := common.Concat(theBytes)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash(concatBytes)

	return hash[:], nil
}

func (o *BaseOplog) MasterSign(id *types.PttID, keyInfo *KeyInfo) error {
	// ts
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	// check expire
	expireTS := ts
	expireTS.Ts -= ExpireOplogSeconds
	if o.CreateTS.IsLess(expireTS) {
		return ErrInvalidOplog
	}

	// init
	origMasterLogID := o.MasterLogID
	origSync := o.IsSync
	origMasterSigns := o.MasterSigns
	origExtra := o.Extra
	defer func(o *BaseOplog) {
		o.MasterLogID = origMasterLogID
		o.IsSync = origSync
		o.Extra = origExtra
	}(o)

	o.UpdateTS = types.ZeroTimestamp
	o.Hash = nil

	o.MasterLogID = nil
	o.MasterSigns = nil
	o.InternalSigns = nil
	o.IsSync = false
	o.IsNewer = false
	o.Extra = nil

	// sign
	marshaled, err := o.Marshal()
	if err != nil {
		return err
	}

	bytesWithSalt, hash, sig, pubBytes, err := SignData(marshaled, keyInfo)
	if err != nil {
		return err
	}

	masterSign := &SignInfo{
		ID:       id,
		CreateTS: ts,

		Hash:   hash,
		Sig:    sig,
		Pubkey: pubBytes,
		Extra:  keyInfo.Extra,
	}

	copy(masterSign.Salt[:], bytesWithSalt[len(marshaled):])

	// post-sign
	// XXX master-signs in order
	idx := sort.Search(len(origMasterSigns), func(i int) bool {
		return bytes.Compare(origMasterSigns[i].ID[:], masterSign.ID[:]) >= 0
	})

	// insert-into-slice
	origMasterSigns = append(origMasterSigns, nil)
	copy(origMasterSigns[(idx+1):], origMasterSigns[idx:])
	origMasterSigns[idx] = masterSign
	o.MasterSigns = origMasterSigns

	o.UpdateTS = ts
	o.Hash, _ = o.SignsHash()

	return nil
}

func (o *BaseOplog) InternalSign(id *types.PttID, keyInfo *KeyInfo) error {
	// ts
	ts, err := types.GetTimestamp()
	if err != nil {
		return err
	}

	// check expire
	expireTS := ts
	expireTS.Ts -= ExpireOplogSeconds
	if o.CreateTS.IsLess(expireTS) {
		return ErrInvalidOplog
	}

	// init
	origMasterLogID, origMasterSigns := o.MasterLogID, o.MasterSigns
	origSync := o.IsSync
	origInternalSigns := o.InternalSigns
	origExtra := o.Extra
	defer func(o *BaseOplog) {
		o.MasterLogID, o.MasterSigns = origMasterLogID, origMasterSigns
		o.IsSync = origSync
		o.Extra = origExtra
	}(o)

	o.UpdateTS = types.ZeroTimestamp
	o.Hash = nil

	o.MasterLogID = nil
	o.MasterSigns = nil
	o.InternalSigns = nil
	o.IsSync = false
	o.IsNewer = false
	o.Extra = nil

	// sign
	marshaled, err := o.Marshal()
	if err != nil {
		return err
	}

	bytesWithSalt, hash, sig, pubBytes, err := SignData(marshaled, keyInfo)
	if err != nil {
		return err
	}

	internalSign := &SignInfo{
		ID:       id,
		CreateTS: ts,

		Hash:   hash,
		Sig:    sig,
		Pubkey: pubBytes,
		Extra:  keyInfo.Extra,
	}

	copy(internalSign.Salt[:], bytesWithSalt[len(marshaled):])

	// post-sign
	// XXX internal-signs in order
	idx := sort.Search(len(origInternalSigns), func(i int) bool {
		return bytes.Compare(origInternalSigns[i].ID[:], internalSign.ID[:]) >= 0
	})

	// insert-into-slice
	origInternalSigns = append(origInternalSigns, nil)
	copy(origInternalSigns[(idx+1):], origInternalSigns[idx:])
	origInternalSigns[idx] = internalSign
	o.InternalSigns = origInternalSigns

	o.UpdateTS = ts
	o.Hash, _ = o.SignsHash()

	return nil
}

func (o *BaseOplog) SetMasterLogID(oplogID *types.PttID, weight uint32) error {
	o.MasterLogID = oplogID
	o.InternalSigns = nil
	o.Hash, _ = o.SignsHash()

	return nil
}

func (o *BaseOplog) Verify() error {
	hash, err := o.SignsHash()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(o.Hash, hash) {
		log.Warn("Verify: hash not equal")
		return ErrInvalidData
	}

	origUpdateTS := o.UpdateTS
	origHash, origCreatorHash, origSalt, origSig, origPubBytes, origKeyExtra := o.Hash, o.CreatorHash, o.Salt, o.Sig, o.Pubkey, o.KeyExtra
	origMasterLogID, origMasterSigns, origInternalSigns := o.MasterLogID, o.MasterSigns, o.InternalSigns
	origIsSync := o.IsSync
	origIsNewer := o.IsNewer
	origExtra := o.Extra
	defer func(o *BaseOplog) {
		o.UpdateTS = origUpdateTS
		o.Hash, o.CreatorHash, o.Salt, o.Sig, o.Pubkey = origHash, origCreatorHash, origSalt, origSig, origPubBytes
		o.MasterLogID, o.MasterSigns, o.InternalSigns = origMasterLogID, origMasterSigns, origInternalSigns
		o.IsSync = origIsSync
		o.IsNewer = origIsNewer
		o.Extra = origExtra
	}(o)

	o.UpdateTS = types.ZeroTimestamp
	o.Hash = nil

	o.CreatorHash = nil
	o.Sig = nil
	o.Salt = types.Salt{}
	o.Pubkey = nil
	o.KeyExtra = nil

	o.MasterLogID = nil
	o.MasterSigns = nil
	o.InternalSigns = nil
	o.IsSync = false
	o.IsNewer = false
	o.Extra = nil

	marshaled, err := o.Marshal()
	if err != nil {
		return err
	}
	bytesWithSalt := append(marshaled, origSalt[:]...)

	err = VerifyData(bytesWithSalt, origSig, origPubBytes, o.CreatorID, origKeyExtra)
	if err != nil {
		log.Warn("Verify (sign)", "bytesWithSalt", bytesWithSalt, "origSig", origSig, "origPubBytes", origPubBytes)
		return err
	}

	o.Hash = nil
	o.CreatorHash = origCreatorHash
	o.Sig = origSig
	o.Salt = origSalt
	o.Pubkey = origPubBytes
	o.KeyExtra = origKeyExtra

	// master signs
	if origMasterSigns != nil {
		for _, masterSign := range origMasterSigns {
			marshaled, err = o.Marshal()
			if err != nil {
				return err
			}
			bytesWithSalt = append(marshaled, masterSign.Salt[:]...)

			err = VerifyData(bytesWithSalt, masterSign.Sig, masterSign.Pubkey, masterSign.ID, masterSign.Extra)
			if err != nil {
				log.Warn("Verify (master-sign)", "masterSign", masterSign)
				return err
			}
		}
	}

	// internal signs
	if origInternalSigns != nil {
		for _, internalSign := range origInternalSigns {
			marshaled, err = o.Marshal()
			if err != nil {
				return err
			}
			bytesWithSalt = append(marshaled, internalSign.Salt[:]...)

			err = VerifyData(bytesWithSalt, internalSign.Sig, internalSign.Pubkey, internalSign.ID, internalSign.Extra)
			if err != nil {
				log.Warn("Verify (internal-sign)", "internalSign", internalSign)
				return err
			}
		}
	}

	return nil
}

func (o *BaseOplog) CheckAlreadyExists() error {
	idxKey, err := o.IdxKey()
	if err != nil {
		return err
	}

	val, err := o.db.DBGet(idxKey)
	if err == leveldb.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	idx := &pttdb.Index{}
	err = idx.Unmarshal(val)
	if err != nil {
		return err
	}

	if o.UpdateTS.IsLessEqual(idx.UpdateTS) {
		return ErrOplogAlreadyExists
	}

	return nil
}

func (o *BaseOplog) ToStatus() types.Status {
	switch {
	case o.MasterLogID != nil:
		return types.StatusAlive
	case len(o.InternalSigns) > 0:
		return types.StatusInternalPending
	case len(o.MasterSigns) > 0:
		return types.StatusPending
	}

	return types.StatusInvalid
}

func (o *BaseOplog) SetPreLogID(id *types.PttID) {
	o.PreLogID = id
}

func (o *BaseOplog) SelectExisting(isLocked bool) error {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	// newest-log
	origIsNewer := o.IsNewer
	defer func() {
		o.IsNewer = origIsNewer
	}()
	o.IsNewer = false

	// get orig
	orig := &BaseOplog{}
	orig.SetDB(o.db, o.dbPrefixID, o.dbPrefix, o.dbIdxPrefix, o.dbMerklePrefix, o.dbLock)

	err := orig.Get(o.ID, true)
	if err == leveldb.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	// is-sync
	status := o.ToStatus()
	origStatus := orig.ToStatus()
	if status <= origStatus {
		o.IsSync = orig.IsSync
	}
	o.Extra = orig.Extra

	// same
	cmp := bytes.Compare(o.Hash, orig.Hash)
	if cmp == 0 {
		return nil
	}

	// require o is valid
	if o.MasterLogID == nil {
		return ErrInvalidOplog
	}

	// orig is not valid
	if orig.MasterLogID == nil {
		o.ForceSave(true)
		return nil
	}

	// both are valid
	// 1. cmp update-ts
	// 2. cmp hash
	switch {
	case o.UpdateTS.IsLess(orig.UpdateTS):
		err = o.ForceSave(true)
	case orig.UpdateTS.IsLess(o.UpdateTS):
		o.UpdateTS = orig.UpdateTS
		o.Hash = orig.Hash
		o.MasterLogID = orig.MasterLogID
		o.MasterSigns = orig.MasterSigns
		o.InternalSigns = orig.InternalSigns
		err = nil
	case cmp < 0:
		err = o.ForceSave(true)
	default:
		o.UpdateTS = orig.UpdateTS
		o.Hash = orig.Hash
		o.MasterLogID = orig.MasterLogID
		o.MasterSigns = orig.MasterSigns
		o.InternalSigns = orig.InternalSigns
		err = nil
	}

	return err
}

// IntegrateExisting integrates with existing oplog.
// Return: is-to-re-sign, error
func (o *BaseOplog) IntegrateExisting(isLocked bool) (bool, error) {
	if !isLocked {
		err := o.dbLock.Lock(o.ID)
		if err != nil {
			return false, err
		}
		defer o.dbLock.Unlock(o.ID)
	}

	// newest-log
	origIsNewer := o.IsNewer
	defer func() {
		o.IsNewer = origIsNewer
	}()
	o.IsNewer = false

	orig := &BaseOplog{}
	orig.SetDB(o.db, o.dbPrefixID, o.dbPrefix, o.dbIdxPrefix, o.dbMerklePrefix, o.dbLock)
	// no orig-log
	err := orig.Get(o.ID, true)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// is-sync
	status := o.ToStatus()
	origStatus := orig.ToStatus()
	if status <= origStatus {
		o.IsSync = orig.IsSync
	}
	o.Extra = orig.Extra

	// same
	if reflect.DeepEqual(o.Hash, orig.Hash) {
		return false, nil
	}

	// o valid
	if o.MasterLogID != nil {
		err = o.SelectExisting(true)
		return false, err
	}

	// orig valid
	if orig.MasterLogID != nil {
		o.UpdateTS = orig.UpdateTS
		o.Hash = orig.Hash
		o.MasterLogID = orig.MasterLogID
		o.MasterSigns = orig.MasterSigns
		o.InternalSigns = orig.InternalSigns
		return false, nil
	}

	// both not valid
	newMasterSigns, isAllMasters, isAllOrigMasters, err := integrateSignInfos(o.MasterSigns, orig.MasterSigns)
	if err != nil {
		return false, err
	}
	o.MasterSigns = newMasterSigns

	newInternalSigns, isAllInternals, isAllOrigInternals, err := integrateSignInfos(o.InternalSigns, orig.InternalSigns)
	if err != nil {
		return false, err
	}
	o.InternalSigns = newInternalSigns

	if isAllOrigMasters && isAllOrigInternals {
		o.UpdateTS = orig.UpdateTS
		o.Hash = orig.Hash

		return false, nil
	}

	if isAllMasters && isAllInternals {
		err = o.ForceSave(true)
		return false, err
	}

	// new-sign
	o.Hash, err = o.SignsHash()
	if err != nil {
		return false, err
	}

	o.UpdateTS, err = types.GetTimestamp()
	if err != nil {
		return false, err
	}

	return true, nil
}

func integrateSignInfos(signInfos []*SignInfo, origSignInfos []*SignInfo) ([]*SignInfo, bool, bool, error) {
	lenSignInfos := len(signInfos)
	lenOrigSignInfos := len(origSignInfos)

	if lenSignInfos == 0 && lenOrigSignInfos == 0 {
		return nil, true, true, nil
	}

	if lenOrigSignInfos == 0 {
		return signInfos, true, false, nil
	}

	if lenSignInfos == 0 {
		return origSignInfos, false, true, nil
	}

	newSignInfos := make([]*SignInfo, 0, lenSignInfos+lenOrigSignInfos)
	idx := 0
	idxOrig := 0
	signInfoID := signInfos[idx].ID
	origSignInfoID := origSignInfos[idxOrig].ID
	var cmp int

loop:
	for {
		cmp = bytes.Compare(signInfoID[:], origSignInfoID[:])
		switch {
		case cmp < 0:
			newSignInfos = append(newSignInfos, signInfos[idx])
			idx++
			if idx >= lenSignInfos {
				break loop
			}
			signInfoID = signInfos[idx].ID
		case cmp > 0:
			newSignInfos = append(newSignInfos, origSignInfos[idxOrig])
			idxOrig++
			if idxOrig >= lenOrigSignInfos {
				break loop
			}
			origSignInfoID = origSignInfos[idxOrig].ID
		case cmp == 0:
			newSignInfos = append(newSignInfos, signInfos[idx])
			idx++
			idxOrig++
			if idx >= lenSignInfos || idxOrig >= lenOrigSignInfos {
				break loop
			}
			signInfoID = signInfos[idx].ID
			origSignInfoID = origSignInfos[idxOrig].ID
		}
	}

	return newSignInfos, len(newSignInfos) == lenSignInfos, len(newSignInfos) == lenOrigSignInfos, nil
}

func (o *BaseOplog) Lock() error {
	return o.dbLock.Lock(o.ID)
}

func (o *BaseOplog) Unlock() error {
	return o.dbLock.Unlock(o.ID)
}

func dbPrefixToDBPrefixInternal(prefix []byte) []byte {
	dbPrefixInternal := common.CloneBytes(prefix)
	dbPrefixInternal[pttdb.SizeDBKeyPrefix-1] = 'i'

	return dbPrefixInternal
}

func dbPrefixToDBPrefixMaster(prefix []byte) []byte {
	dbPrefixMaster := common.CloneBytes(prefix)
	dbPrefixMaster[pttdb.SizeDBKeyPrefix-1] = 'm'

	return dbPrefixMaster
}

func bytesToStatus(theBytes []byte) types.Status {
	switch theBytes[pttdb.SizeDBKeyPrefix-1] {
	case 'i':
		return types.StatusInternalPending
	case 'm':
		return types.StatusPending
	}

	return types.StatusAlive
}

func OplogKeyToIDBytes(key []byte) ([]byte, error) {
	offset := pttdb.SizeDBKeyPrefix + types.SizePttID + types.SizeTimestamp
	next := offset + types.SizePttID

	if len(key) < next {
		return nil, ErrInvalidKey
	}

	return key[offset:next], nil
}

func OplogKeyToIdxKey(key []byte, dbIdxPrefix []byte) []byte {
	offset := pttdb.SizeDBKeyPrefix
	nextOffset := offset + types.SizePttID
	prefixID := &types.PttID{}
	copy(prefixID[:], key[offset:nextOffset])

	offset = nextOffset + types.SizeTimestamp
	nextOffset = offset + types.SizePttID
	theID := &types.PttID{}
	copy(theID[:], key[offset:nextOffset])

	o := &BaseOplog{
		dbIdxPrefix: dbIdxPrefix,
		dbPrefixID:  prefixID,
		ID:          theID,
	}

	idxKey, _ := o.IdxKey()
	return idxKey
}
