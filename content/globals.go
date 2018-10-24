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
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
)

// config
var (
	DefaultConfig = Config{
		DataDir:     filepath.Join(node.DefaultDataDir(), "content"),
		KeystoreDir: filepath.Join(node.DefaultDataDir(), ".keyContent"),
	}
)

// MyInfo
var (
	MyID     *types.PttID     = nil
	MyNodeID *discover.NodeID = nil
)

// db
var (
	dbKey *pttdb.LDBDatabase = nil

	dbBoardCore *pttdb.LDBDatabase = nil
	dbBoard     *pttdb.LDBBatch    = nil

	dbCommentCore *pttdb.LDBDatabase = nil
	dbComment     *pttdb.LDBBatch    = nil

	dbMeta *pttdb.LDBDatabase = nil

	DBNodeIdxOplogPrefix    = []byte(".ndig")
	DBNodeOplogPrefix       = []byte(".ndlg")
	DBNodeMerkleOplogPrefix = []byte(".ndmk")

	DBBoardIdxOplogPrefix    = []byte(".bdig")
	DBBoardOplogPrefix       = []byte(".bdlg")
	DBBoardMerkleOplogPrefix = []byte(".bdmk")

	DBCommentIdxOplogPrefix    = []byte(".ctig")
	DBCommentOplogPrefix       = []byte(".ctlg")
	DBCommentMerkleOplogPrefix = []byte(".ctmk")

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

	// squeeze article and comment and reply to have better reference
	DBContentBlockPrefix = []byte(".cont")

	DBMasterPrefix            = []byte(".mrdb")
	DBMasterIdxPrefix         = []byte(".mrix")
	DBMasterIdxOplogPrefix    = []byte(".mrig")
	DBMasterOplogPrefix       = []byte(".mrlg")
	DBMasterMerkleOplogPrefix = []byte(".mrmk")

	DBMemberPrefix            = []byte(".mbdb")
	DBMemberIdxPrefix         = []byte(".mbix")
	DBMemberIdxOplogPrefix    = []byte(".mbig")
	DBMemberOplogPrefix       = []byte(".mblg")
	DBMemberMerkleOplogPrefix = []byte(".mbmk")
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

	dbCommentCore, err = pttdb.NewLDBDatabase("comment", dataDir, 0, 0)
	if err != nil {
		return err
	}

	dbComment, err = pttdb.NewLDBBatch(dbCommentCore)
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

	return nil
}

func initMyInfo(id *types.PttID, nodeID *discover.NodeID) error {
	MyID = id
	MyNodeID = nodeID

	return nil
}

func TeardownContent() {
	if dbBoard != nil {
		dbBoard = nil
	}

	if dbBoardCore != nil {
		dbBoardCore.Close()
		dbBoardCore = nil
	}

	if dbComment != nil {
		dbComment = nil
	}

	if dbCommentCore != nil {
		dbCommentCore.Close()
		dbCommentCore = nil
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
