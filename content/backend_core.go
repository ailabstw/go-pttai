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
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

func (b *Backend) CreateBoard(title []byte, isPublic bool) (*Board, error) {
	return nil, types.ErrNotImplemented
}

func (b *Backend) CreateArticle(entityIDBytes []byte, title []byte, article [][]byte, mediaIDStrs []string) (*BackendCreateArticle, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	lenMediaIDs := len(mediaIDStrs)
	var mediaIDs []*types.PttID = nil
	if len(mediaIDStrs) != 0 {
		mediaIDs = make([]*types.PttID, lenMediaIDs)
		for i, mediaIDStr := range mediaIDStrs {
			mediaIDs[i] = &types.PttID{}
			err := mediaIDs[i].UnmarshalText([]byte(mediaIDStr))
			if err != nil {
				return nil, err
			}
		}
	}

	theArticle, err := pm.CreateArticle(title, article, mediaIDs)
	if err != nil {
		return nil, err
	}

	backendArticle := articleToBackendCreateArticle(theArticle)

	return backendArticle, nil
}

func (b *Backend) CreateComment(entityIDBytes []byte, articleIDBytes []byte, commentType CommentType, commentBytes []byte, mediaIDBytes []byte) (*BackendCreateComment, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	articleID, err := types.UnmarshalTextPttID(articleIDBytes)
	if err != nil {
		return nil, err
	}
	if articleID == nil {
		return nil, types.ErrInvalidID
	}

	mediaID, err := types.UnmarshalTextPttID(mediaIDBytes)
	if err != nil {
		return nil, err
	}

	theComment, err := pm.CreateComment(articleID, commentType, commentBytes, mediaID)
	if err != nil {
		return nil, err
	}

	backendComment := commentToBackendCreateComment(theComment)

	return backendComment, nil
}

func (b *Backend) CreateReply(entityIDBytes []byte, articleIDBytes []byte, commentIDBytes []byte, reply [][]byte, mediaIDBytes []byte) (*BackendCreateReply, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) UpdateArticle(entityIDBytes []byte, articleIDBytes []byte, article [][]byte, mediaIDStrs []string) (*BackendUpdateArticle, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) UpdateReply(entityIDBytes []byte, articleIDBytes []byte, commentIDBytes []byte, reply [][]byte, mediaIDBytes []byte) (*BackendUpdateReply, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) DeleteArticle(entityIDBytes []byte, articleIDBytes []byte) (*BackendDeleteArticle, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) DeleteComment(entityIDBytes []byte, articleIDBytes []byte, commentIDBytes []byte) (*BackendDeleteComment, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) DeleteReply(entityIDBytes []byte, articleIDBytes []byte, commentIDBytes []byte) (*BackendDeleteReply, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) DeleteBoard(entityIDBytes []byte) (*BackendDeleteBoard, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) LeaveBoard(entityIDBytes []byte) (*BackendLeaveBoard, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) InviteMaster(boardID []byte, userID []byte, nodeURL []byte) (*BackendInviteMaster, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) RevokeMaster(boardID []byte, userID []byte) (*BackendRevokeMaster, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) TransferMaster(boardID []byte, userID []byte) (*BackendRevokeMaster, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) GetBoard(entityIDBytes []byte) (*BackendGetBoard, error) {
	board, err := b.GetRawBoard(entityIDBytes)
	if err != nil {
		return nil, err
	}

	myID := b.Ptt().GetMyEntity().GetID()
	accountBackend := b.accountBackend
	userName, err := accountBackend.GetRawUserNameByID(board.CreatorID)
	if err != nil {
		userName = account.NewEmptyUserName()
	}
	theTitle, err := b.GetRawTitleByID(board.ID)

	backendBoard := boardToBackendGetBoard(board, string(userName.Name), theTitle, myID)

	return backendBoard, nil
}

func (b *Backend) GetRawBoard(entityIDBytes []byte) (*Board, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	log.Debug("GetRawBoard: after unmarshal", "entityIDBytes", entityIDBytes, "entityID", entityID, "e", err)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	board := entity.(*Board)

	return board, nil
}

func (b *Backend) GetBoardList(startingIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetBoard, error) {

	startID, err := types.UnmarshalTextPttID(startingIDBytes)
	if err != nil {
		return nil, err
	}

	boardList, err := b.SPM().(*ServiceProtocolManager).GetBoardList(startID, limit, listOrder)
	if err != nil {
		return nil, err
	}

	accountBackend := b.accountBackend
	backendBoardList := make([]*BackendGetBoard, len(boardList))
	var userName *account.UserName
	var title *Title
	myID := b.Ptt().GetMyEntity().GetID()
	for i, f := range boardList {
		userName, err = accountBackend.GetRawUserNameByID(f.CreatorID)
		if err != nil {
			userName = account.NewEmptyUserName()
		}
		title, err = b.GetRawTitleByID(f.ID)
		backendBoardList[i] = boardToBackendGetBoard(f, string(userName.Name), title, myID)
	}

	return backendBoardList, nil
}

func (b *Backend) GetArticle(entityIDBytes []byte, articleIDBytes []byte) (*BackendGetArticle, error) {

	article, err := b.GetRawArticle(entityIDBytes, articleIDBytes)
	if err != nil {
		return nil, err
	}

	return articleToBackendGetArticle(article), nil
}

func (b *Backend) GetRawArticle(entityIDBytes []byte, articleIDBytes []byte) (*Article, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	articleID, err := types.UnmarshalTextPttID(articleIDBytes)
	if err != nil {
		return nil, err
	}
	if articleID == nil {
		return nil, types.ErrInvalidID
	}

	return pm.GetArticle(articleID)
}

func (b *Backend) GetRawComment(entityIDBytes []byte, commentIDBytes []byte) (*Comment, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) GetRawReply(entityIDBytes []byte, articleIDBytes []byte, commentIDBytes []byte) (*Reply, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) GetArticleBlockList(entityIDBytes []byte, articleIDBytes []byte, subContentIDBytes []byte, contentType ContentType, blockID uint32, limit int, listOrder pttdb.ListOrder) ([]*ArticleBlock, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	articleID, err := types.UnmarshalTextPttID(articleIDBytes)
	if err != nil {
		return nil, err
	}
	if articleID == nil {
		return nil, types.ErrInvalidID
	}

	blockInfoID, err := types.UnmarshalTextPttID(subContentIDBytes)
	if err != nil {
		return nil, err
	}

	articleBlockList, err := pm.GetArticleBlockList(articleID, blockInfoID, contentType, blockID, limit, listOrder)
	if err != nil {
		return nil, err
	}

	return articleBlockList, nil
}

func (b *Backend) GetArticleList(entityIDBytes []byte, startingArticleIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BackendGetArticle, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	startID, err := types.UnmarshalTextPttID(startingArticleIDBytes)
	if err != nil {
		return nil, err
	}

	articleList, err := pm.GetArticleList(startID, limit, listOrder, false)
	if err != nil {
		return nil, err
	}
	theList := make([]*BackendGetArticle, len(articleList))
	for i, article := range articleList {
		theList[i] = articleToBackendGetArticle(article)
	}

	return theList, nil
}

func (b *Backend) GetPokedArticleList(boardID []byte) ([]*BackendGetArticle, error) {

	return nil, types.ErrNotImplemented
}

func (b *Backend) ShowBoardURL(entityIDBytes []byte) (*pkgservice.BackendJoinURL, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	board, ok := entity.(*Board)
	if !ok {
		return nil, pkgservice.ErrInvalidEntity
	}

	nodeID := b.Ptt().MyNodeID()

	myID := b.Ptt().GetMyEntity().GetID()
	pm := entity.PM().(*ProtocolManager)

	if !pm.IsMaster(myID, false) {
		return nil, types.ErrInvalidID
	}

	keyInfo, err := pm.GetJoinKey()
	if err != nil {
		return nil, err
	}

	theTitle, err := pm.GetTitle()
	if err != nil {
		return nil, err
	}
	title := board.Title
	if theTitle != nil {
		title = theTitle.Title
	}

	return pkgservice.MarshalBackendJoinURL(board.CreatorID, nodeID, keyInfo, title, pkgservice.PathJoinBoard)
}

/**********
 * BoardOplog
 **********/

func (b *Backend) GetBoardOplogList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetBoardOplogList(logID, limit, listOrder, types.StatusAlive)
}

func (b *Backend) GetPendingBoardOplogMasterList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetBoardOplogList(logID, limit, listOrder, types.StatusPending)
}

func (b *Backend) GetPendingBoardOplogInternalList(entityIDBytes []byte, logIDBytes []byte, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	logID, err := types.UnmarshalTextPttID(logIDBytes)
	if err != nil {
		return nil, err
	}

	return pm.GetBoardOplogList(logID, limit, listOrder, types.StatusInternalPending)
}

func (b *Backend) GetBoardOplogMerkleNodeList(entityIDBytes []byte, level pkgservice.MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	merkleNodeList, err := pm.GetBoardOplogMerkleNodeList(level, startKey, limit, listOrder)
	if err != nil {
		return nil, err
	}

	results := make([]*pkgservice.BackendMerkleNode, len(merkleNodeList))
	for i, eachMerkleNode := range merkleNodeList {
		results[i] = pkgservice.MerkleNodeToBackendMerkleNode(eachMerkleNode)
	}

	return results, nil
}

func (b *Backend) UploadFile(entityIDBytes []byte, filename []byte, bytes []byte) (*BackendUploadFile, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	media, err := pm.UploadFile(filename, bytes)
	if err != nil {
		return nil, err
	}

	return mediaToBackendUploadFile(media), nil
}

func (b *Backend) GetFile(entityIDBytes []byte, mediaIDBytes []byte) (*BackendGetFile, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	mediaID, err := types.UnmarshalTextPttID(mediaIDBytes)
	if err != nil {
		return nil, err
	}
	if mediaID == nil {
		return nil, types.ErrInvalidID
	}

	f, err := pm.GetMedia(mediaID)
	if err != nil {
		return nil, err
	}

	return mediaToBackendGetFile(f), nil
}

func (b *Backend) UploadImage(entityIDBytes []byte, fileType string, bytes []byte) (*BackendUploadImg, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	media, err := pm.UploadImage(fileType, bytes)
	if err != nil {
		return nil, err
	}

	return mediaToBackendUploadImg(media), nil
}

func (b *Backend) GetImage(entityIDBytes []byte, mediaIDBytes []byte) (*BackendGetImg, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	mediaID, err := types.UnmarshalTextPttID(mediaIDBytes)
	if err != nil {
		return nil, err
	}
	if mediaID == nil {
		return nil, types.ErrInvalidID
	}

	media, err := pm.GetMedia(mediaID)
	if err != nil {
		return nil, err
	}

	return mediaToBackendGetImg(media), nil

}

func (b *Backend) GetArticleSummary(entityIDBytes []byte, articleInfo *BackendArticleSummaryParams) (*ArticleBlock, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	articleID, err := types.UnmarshalTextPttID([]byte(articleInfo.ArticleID))
	if err != nil {
		return nil, err
	}
	if articleID == nil {
		return nil, types.ErrInvalidID
	}

	blockInfoID, err := types.UnmarshalTextPttID([]byte(articleInfo.ContentBlockID))
	if err != nil {
		return nil, err
	}

	articleBlockList, err := pm.GetArticleBlockList(articleID, blockInfoID, ContentTypeArticle, 0, 1, pttdb.ListOrderNext)
	if err != nil {
		return nil, err
	}
	if len(articleBlockList) != 1 {
		return nil, ErrInvalidBlock
	}

	return articleBlockList[0], nil
}

func (b *Backend) GetArticleSummaryByIDs(entityIDBytes []byte, articleInfos []*BackendArticleSummaryParams) (map[string]*ArticleBlock, error) {

	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	articleBlocks := make(map[string]*ArticleBlock)
	var articleID *types.PttID
	var contentBlockID *types.PttID
	var articleBlockList []*ArticleBlock
	for _, articleInfo := range articleInfos {
		articleID, err = types.UnmarshalTextPttID([]byte(articleInfo.ArticleID))
		if err != nil {
			continue
		}
		if articleID == nil {
			continue
		}

		contentBlockID, err = types.UnmarshalTextPttID([]byte(articleInfo.ContentBlockID))
		if err != nil {
			continue
		}

		articleBlockList, err = pm.GetArticleBlockList(articleID, contentBlockID, ContentTypeArticle, 0, 1, pttdb.ListOrderNext)
		if err != nil {
			continue
		}
		if len(articleBlockList) != 1 {
			continue
		}
		articleBlocks[articleInfo.ArticleID] = articleBlockList[0]
	}

	return articleBlocks, nil
}

func (b *Backend) MarkBoardSeen(entityIDBytes []byte) (types.Timestamp, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return types.ZeroTimestamp, err
	}
	if entityID == nil {
		return types.ZeroTimestamp, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return types.ZeroTimestamp, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.SaveLastSeen(types.ZeroTimestamp)
}

func (b *Backend) MarkArticleSeen(entityIDBytes []byte, articleIDBytes []byte) (types.Timestamp, error) {

	return types.ZeroTimestamp, types.ErrNotImplemented
}

func (b *Backend) SetTitle(entityIDBytes []byte, title []byte) (*BackendGetBoard, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	log.Debug("GetRawBoard: after unmarshal", "entityIDBytes", entityIDBytes, "entityID", entityID, "e", err)
	if err != nil {
		return nil, err
	}
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	err = pm.SetTitle(title)
	if err != nil {
		return nil, err
	}

	board := entity.(*Board)

	myID := b.Ptt().GetMyEntity().GetID()
	idBytes, err := myID.MarshalText()
	if err != nil {
		return nil, err
	}

	userName, err := b.accountBackend.GetRawUserName(idBytes)
	if err != nil {
		return nil, err
	}
	myName := userName.Name

	theTitle, err := b.GetRawTitle(entityIDBytes)
	if err == leveldb.ErrNotFound {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return boardToBackendGetBoard(board, string(myName), theTitle, myID), nil
}

func (b *Backend) GetRawTitle(entityIDBytes []byte) (*Title, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	log.Debug("GetRawBoard: after unmarshal", "entityIDBytes", entityIDBytes, "entityID", entityID, "e", err)
	if err != nil {
		return nil, err
	}

	return b.GetRawTitleByID(entityID)
}

func (b *Backend) GetRawTitleByID(entityID *types.PttID) (*Title, error) {
	if entityID == nil {
		return nil, types.ErrInvalidID
	}

	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.GetTitle()
}

func (b *Backend) GetJoinKeys(entityIDBytes []byte) ([]*pkgservice.KeyInfo, error) {
	entityID, err := types.UnmarshalTextPttID(entityIDBytes)
	if err != nil {
		return nil, err
	}
	entity := b.SPM().Entity(entityID)
	if entity == nil {
		return nil, types.ErrInvalidID
	}
	pm := entity.PM().(*ProtocolManager)

	return pm.JoinKeyList(), nil
}
