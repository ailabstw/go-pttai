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

package content

import (
	"path/filepath"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

// config
var (
	DefaultConfig = Config{
		DataDir:     filepath.Join(node.DefaultDataDir(), "content"),
		KeystoreDir: filepath.Join(node.DefaultDataDir(), ".keyContent"),
	}
)

// protocol
const (
	_ pkgservice.OpType = iota + pkgservice.NMsg
	// friend-oplog
	AddBoardOplogMsg //30
	AddBoardOplogsMsg

	AddPendingBoardOplogMsg
	AddPendingBoardOplogsMsg

	SyncBoardOplogMsg
	SyncBoardOplogAckMsg
	SyncBoardOplogNewOplogsMsg
	SyncBoardOplogNewOplogsAckMsg

	SyncPendingBoardOplogMsg
	SyncPendingBoardOplogAckMsg

	SyncCreateMessageMsg
	SyncCreateMessageAckMsg

	SyncCreateMessageBlockMsg
	SyncCreateMessageBlockAckMsg

	// init board info
	InitBoardInfoMsg
	InitBoardInfoAckMsg
)

// db
var (
	dbKey *pttdb.LDBDatabase = nil

	dbBoardCore *pttdb.LDBDatabase = nil
	dbBoard     *pttdb.LDBBatch    = nil

	dbMeta *pttdb.LDBDatabase = nil

	DBBoardIdxOplogPrefix    = []byte(".bdig")
	DBBoardOplogPrefix       = []byte(".bdlg")
	DBBoardMerkleOplogPrefix = []byte(".bdmk")

	DBBoardPrefix                  = []byte(".bddb")
	DBBoardIdxPrefix               = []byte(".bdix")
	DBBoardIdx2Prefix              = []byte(".bdi2")
	DBBoardLastSeenPrefix          = []byte(".bdls")
	DBBoardArticleCreateTSPrefix   = []byte(".bdac")
	DBBoardCommentCreateTSPrefix   = []byte(".bdcc")
	DBArticlePrefix                = []byte(".aldb")
	DBArticleIdxPrefix             = []byte(".alix")
	DBArticleLastSeenPrefix        = []byte(".alls")
	DBArticleCommentCreateTSPrefix = []byte(".alcc")
	DBPushPrefix                   = []byte(".alps")
	DBBooPrefix                    = []byte(".albo")
	DBCommentPrefix                = []byte(".ctdb")
	DBCommentIdxPrefix             = []byte(".ctix")
	DBReplyPrefix                  = []byte(".rpdb")
	DBReplyIdxPrefix               = []byte(".rpix")
	DBImagePrefix                  = []byte(".imdb")
	DBImageIdxPrefix               = []byte(".imix")
	DBMediaPrefix                  = []byte(".madb")
	DBMediaIdxPrefix               = []byte(".maix")
	DBTitlePrefix                  = []byte(".tldb")
	DBTitleIdxPrefix               = []byte(".tlix")
)

// max-masters
const (
	MaxMasters = 1
)

// sync
const (
	MaxSyncRandomSeconds = 30
	MinSyncRandomSeconds = 15
)

// op-key
var (
	RenewOpKeySeconds  int64 = 86400
	ExpireOpKeySeconds int64 = 259200
)

// article
const (
	NFirstLineInBlock = 1
)

func InitContent(dataDir string, keystoreDir string) error {
	var err error

	// db
	dbBoardCore, err = pttdb.NewLDBDatabase("board", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbBoard, err = pttdb.NewLDBBatch(dbBoardCore)
	if err != nil {
		return err
	}

	dbKey, err = pttdb.NewLDBDatabase("key", keystoreDir, 0, 0)
	if err != nil {
		return err
	}

	dbMeta, err = pttdb.NewLDBDatabase("contentmeta", dataDir, 0, 0)
	if err != nil {
		return err
	}

	InitLocaleInfo()

	return nil
}

// default-title
func DefaultTitle(myID *types.PttID, creatorID *types.PttID, myName string) []byte {
	log.Debug("DefaultTitle: start", "myID", myID, "creatorID", creatorID, "myName", myName, "currentLocale", pkgservice.CurrentLocale)
	return localeInfos[pkgservice.CurrentLocale].DefaultTitle(myID, creatorID, myName)
}

func TeardownContent() {
	if dbBoard != nil {
		dbBoard = nil
	}

	if dbBoardCore != nil {
		dbBoardCore.Close()
		dbBoardCore = nil
	}

	if dbKey != nil {
		dbKey.Close()
		dbKey = nil
	}

	if dbMeta != nil {
		dbMeta.Close()
		dbMeta = nil
	}
}
