// Copyright 2019 The go-pttai Authors
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

package account

func (pm *ProtocolManager) CleanObject() error {

	profile := pm.Entity().(*Profile)

	// user-node
	pm.CleanUserNode()

	// user-name
	userName := NewEmptyUserName()
	pm.SetUserNameDB(userName)
	userName.SetID(profile.MyID)
	err := userName.Get(false)
	if err == nil {
		userName.Delete(false)
	}

	// user-img
	userImg := NewEmptyUserImg()
	pm.SetUserImgDB(userImg)
	userImg.SetID(profile.MyID)
	err = userImg.Get(false)
	if err == nil {
		userImg.Delete(false)
	}

	// name-card
	nameCard := NewEmptyNameCard()
	pm.SetNameCardDB(nameCard)
	nameCard.SetID(profile.MyID)
	err = nameCard.Get(false)
	if err == nil {
		nameCard.Delete(false)
	}

	return nil
}
