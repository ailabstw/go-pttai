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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type PrivateAPI struct {
	b *Backend
}

func NewPrivateAPI(b *Backend) *PrivateAPI {
	return &PrivateAPI{b}
}

func (api *PrivateAPI) CreateBoard(title []byte, isPublic bool) (*Board, error) {
	return api.b.CreateBoard(title, isPublic)
}

func (api *PrivateAPI) CreateArticle(boardID string, title []byte, article [][]byte, mediaIDs []string) (*BackendCreateArticle, error) {
	return api.b.CreateArticle(
		[]byte(boardID),
		title,
		article,
		mediaIDs,
	)
}

func (api *PrivateAPI) CreateComment(boardID string, articleID string, commentType CommentType, comment []byte, mediaID string) (*BackendCreateComment, error) {
	return api.b.CreateComment(
		[]byte(boardID),
		[]byte(articleID),
		commentType,
		comment,
		[]byte(mediaID),
	)
}

func (api *PrivateAPI) CreateReply(boardID string, articleID string, commentID string, reply [][]byte, mediaID string) (*BackendCreateReply, error) {
	return api.b.CreateReply(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(commentID),
		reply,
		[]byte(mediaID),
	)
}

func (api *PrivateAPI) SetTitle(boardID string, title []byte) (*BackendGetBoard, error) {
	return api.b.SetTitle([]byte(boardID), title)
}

func (api *PrivateAPI) UpdateArticle(boardID string, articleID string, article [][]byte, mediaIDs []string) (*BackendUpdateArticle, error) {
	return api.b.UpdateArticle(
		[]byte(boardID),
		[]byte(articleID),
		article,
		mediaIDs,
	)
}

func (api *PrivateAPI) UpdateReply(boardID string, articleID string, commentID string, reply [][]byte, mediaID string) (*BackendUpdateReply, error) {
	return api.b.UpdateReply(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(commentID),
		reply,
		[]byte(mediaID),
	)
}

func (api *PrivateAPI) DeleteBoard(boardID string) (*BackendDeleteBoard, error) {
	return api.b.DeleteBoard([]byte(boardID))
}

func (api *PrivateAPI) DeleteArticle(boardID string, articleID string) (*BackendDeleteArticle, error) {
	return api.b.DeleteArticle(
		[]byte(boardID),
		[]byte(articleID),
	)
}

func (api *PrivateAPI) DeleteComment(boardID string, articleID string, commentID string) (*BackendDeleteComment, error) {
	return api.b.DeleteComment(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(commentID),
	)
}

func (api *PrivateAPI) DeleteReply(boardID string, articleID string, commentID string) (*BackendDeleteReply, error) {
	return api.b.DeleteReply(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(commentID),
	)
}

func (api *PrivateAPI) LeaveBoard(boardID string) (*BackendLeaveBoard, error) {
	return api.b.LeaveBoard([]byte(boardID))
}

func (api *PrivateAPI) InviteMaster(boardID string, userID string, nodeURL string) (*BackendInviteMaster, error) {
	return api.b.InviteMaster(
		[]byte(boardID),
		[]byte(userID),
		[]byte(nodeURL),
	)
}

func (api *PrivateAPI) RevokeMaster(boardID string, userID string) (*BackendRevokeMaster, error) {
	return api.b.RevokeMaster(
		[]byte(boardID),
		[]byte(userID),
	)
}

func (api *PrivateAPI) GetJoinKeyInfos(boardID string) ([]*pkgservice.KeyInfo, error) {
	return api.b.GetJoinKeys([]byte(boardID))
}

func (api *PrivateAPI) GetRawBoard(boardID string) (*Board, error) {
	return api.b.GetRawBoard([]byte(boardID))
}

func (api *PrivateAPI) GetRawTitle(boardID string) (*Title, error) {
	return api.b.GetRawTitle([]byte(boardID))
}

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) GetBoard(boardID string) (*BackendGetBoard, error) {
	return api.b.GetBoard([]byte(boardID))
}

func (api *PublicAPI) GetBoardList(startingBoardID string, limit int, listOrder pttdb.ListOrder) ([]*BackendGetBoard, error) {
	return api.b.GetBoardList(
		[]byte(startingBoardID),
		limit,
		listOrder,
	)
}

func (api *PublicAPI) GetArticle(boardID string, articleID string) (*BackendGetArticle, error) {
	return api.b.GetArticle(
		[]byte(boardID),
		[]byte(articleID),
	)
}

func (api *PrivateAPI) GetRawArticle(boardID string, articleID string) (*Article, error) {
	return api.b.GetRawArticle(
		[]byte(boardID),
		[]byte(articleID),
	)
}

func (api *PrivateAPI) GetRawComment(boardID string, commentID string) (*Comment, error) {
	return api.b.GetRawComment(
		[]byte(boardID),
		[]byte(commentID),
	)
}

func (api *PrivateAPI) GetRawReply(boardID string, articleID string, commentID string) (*Reply, error) {
	return api.b.GetRawReply(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(commentID),
	)
}

/*
GetArticleBlockList gets the list of the blocks-to-show of the article, including main-article, comment, reply.

Given the boardID, articleID, and the corresponding subContentID (article: ContentBlockID, comment: commentID, reply: replyID), and the blockID (for comment and reply: blockID as 0)
GetArticleBlockList will get the following blocks from the specified subContentID and blockID.
*/
func (api *PublicAPI) GetArticleBlockList(boardID string, articleID string, subContentID string, contentType ContentType, blockID uint32, limit int, listOrder pttdb.ListOrder) ([]*ArticleBlock, error) {
	return api.b.GetArticleBlockList(
		[]byte(boardID),
		[]byte(articleID),
		[]byte(subContentID),
		contentType,
		blockID,
		limit,
		listOrder,
	)
}

func (api *PublicAPI) GetArticleList(boardID string, startingArticleID string, limit int, listOrder pttdb.ListOrder) ([]*BackendGetArticle, error) {
	return api.b.GetArticleList(
		[]byte(boardID),
		[]byte(startingArticleID),
		limit,
		listOrder,
	)
}

func (api *PublicAPI) GetPokedArticleList(boardID string) ([]*BackendGetArticle, error) {
	return api.b.GetPokedArticleList([]byte(boardID))
}

func (api *PublicAPI) ShowBoardURL(boardID string) (*pkgservice.BackendJoinURL, error) {
	return api.b.ShowBoardURL([]byte(boardID))
}

/**********
 * BoardOplog
 **********/

func (api *PrivateAPI) GetBoardOplogList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {
	return api.b.GetBoardOplogList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingBoardOplogMasterList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {
	return api.b.GetPendingBoardOplogMasterList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingBoardOplogInternalList(profileID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*BoardOplog, error) {
	return api.b.GetPendingBoardOplogInternalList([]byte(profileID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetBoardOplogMerkleNodeList(profileID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetBoardOplogMerkleNodeList([]byte(profileID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

func (api *PrivateAPI) UploadFile(boardID string, filename string, bytes []byte) (*BackendUploadFile, error) {
	return api.b.UploadFile([]byte(boardID), []byte(filename), bytes)
}

func (api *PrivateAPI) GetFile(boardID string, mediaID string) (*BackendGetFile, error) {
	return api.b.GetFile([]byte(boardID), []byte(mediaID))
}

func (api *PrivateAPI) UploadImage(boardID string, fileType string, bytes []byte) (*BackendUploadImg, error) {
	return api.b.UploadImage([]byte(boardID), fileType, bytes)
}

func (api *PrivateAPI) GetImage(boardID string, imgID string) (*BackendGetImg, error) {
	return api.b.GetImage([]byte(boardID), []byte(imgID))
}

func (api *PublicAPI) GetArticleSummary(boardID string, articleInfo *BackendArticleSummaryParams) (*ArticleBlock, error) {
	return api.b.GetArticleSummary([]byte(boardID), articleInfo)
}

func (api *PublicAPI) GetArticleSummaryByIDs(boardID string, articleInfos []*BackendArticleSummaryParams) (map[string]*ArticleBlock, error) {
	return api.b.GetArticleSummaryByIDs([]byte(boardID), articleInfos)
}

func (api *PrivateAPI) MarkBoardSeen(boardID string) (types.Timestamp, error) {
	return api.b.MarkBoardSeen([]byte(boardID))
}

func (api *PrivateAPI) MarkArticleSeen(boardID string, articleID string) (types.Timestamp, error) {
	return api.b.MarkArticleSeen([]byte(boardID), []byte(articleID))
}

/**********
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMasterOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetMasterOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMasterOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MasterOplog, error) {
	return api.b.GetPendingMasterOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMasterOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMasterOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MemberOplog
 **********/

func (api *PrivateAPI) GetMemberOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetMemberOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMemberOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.MemberOplog, error) {
	return api.b.GetPendingMemberOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMemberOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {
	return api.b.GetMemberOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * OpKeyOplog
 **********/

func (api *PrivateAPI) GetOpKeyOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetOpKeyOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetPendingOpKeyOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {
	return api.b.GetPendingOpKeyOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

/**********
 * Master
 **********/

func (api *PrivateAPI) GetMasterListFromCache(entityID string) ([]*pkgservice.Master, error) {
	return api.b.GetMasterListFromCache([]byte(entityID))
}

func (api *PrivateAPI) GetMasterList(entityID string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Master, error) {
	return api.b.GetMasterList([]byte(entityID), []byte(startID), limit, listOrder)
}

/**********
 * Member
 **********/

func (api *PrivateAPI) GetMemberList(entityID string, startID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.Member, error) {
	return api.b.GetMemberList([]byte(entityID), []byte(startID), limit, listOrder)
}

/**********
 * Op
 **********/

func (api *PrivateAPI) ShowValidateKey() (*types.PttID, error) {
	return api.b.ShowValidateKey()
}

func (api *PrivateAPI) ValidateValidateKey(key string) (bool, error) {
	return api.b.ValidateValidateKey([]byte(key))
}

func (api *PrivateAPI) GetOpKeyInfos(entityID string) ([]*pkgservice.KeyInfo, error) {
	return api.b.GetOpKeys([]byte(entityID))
}

func (api *PrivateAPI) RevokeOpKey(entityID string, keyID string, myKey string) (bool, error) {
	return api.b.RevokeOpKey([]byte(entityID), []byte(keyID), []byte(myKey))
}

func (api *PrivateAPI) GetOpKeyInfosFromDB(entityID string) ([]*pkgservice.KeyInfo, error) {
	return api.b.GetOpKeysFromDB([]byte(entityID))
}

/**********
 * Peer
 **********/

func (api *PrivateAPI) CountPeers(entityID string) (int, error) {
	return api.b.CountPeers([]byte(entityID))
}

func (api *PrivateAPI) GetPeers(entityID string) ([]*pkgservice.BackendPeer, error) {
	return api.b.GetPeers([]byte(entityID))
}
