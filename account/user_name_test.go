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

import (
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestUserName_Marshal(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u:    tUserNameA,
			want: tUserNameMarshal,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			got, err := u.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserName.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserName.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestUserName_Unmarshal(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		theBytes []byte
	}

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		args    args
		want    *UserName
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u:    &UserName{},
			args: args{tUserNameMarshal},
			want: tUserNameA,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			err := u.Unmarshal(tt.args.theBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserName.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			u.SetDB(u.DB(), u.DBLock(), tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)

			if !reflect.DeepEqual(u.BaseObject.ID, tt.want.BaseObject.ID) {
				t.Errorf("UserName.Unmarshal() u = %v, tt.want %v", u, tt.want)
			}
		})
	}

	// teardown test
}

func TestUserName_Save(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u: tUserNameA,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			u.SetDB(dbAccount, tLockMap, tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
			if err := u.Save(true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			key, _ := u.MarshalKey()
			if isHas, _ := dbAccountCore.Has(key); !isHas {
				t.Errorf("UserName.Save() id not exists: u: %v", u)
			}
		})
	}

	// teardown test
}

func TestUserName_Get(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tUserNameA.Save(true)
	tUserNameB.Save(true)
	tUserNameC.Save(true)

	// define test-structure
	type args struct {
		isLocked bool
	}

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		args    args
		want    *UserName
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserIDA}},
			args: args{isLocked: true},
			want: tUserNameA,
		},
		{
			u:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserIDB}},
			args: args{isLocked: true},
			want: tUserNameB,
		},
		{
			u:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserIDC}},
			args: args{isLocked: true},
			want: tUserNameC,
		},
		{
			u:       &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserIDD}},
			args:    args{isLocked: true},
			want:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserIDD}},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			u.SetDB(dbAccount, tLockMap, tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
			if err := u.Get(true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.want.SetDB(dbAccount, tLockMap, tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
			if !reflect.DeepEqual(u, tt.want) {
				t.Errorf("UserName.Get() u = %v, want %v", u, tt.want)
			}
		})
	}

	// teardown test
}

func TestUserName_Delete(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tUserNameA.Save(true)

	// define test-structure
	type args struct {
		id *types.PttID
	}

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserNameA.ID}},
			args: args{tUserNameA.ID},
		},
		{
			u:    &UserName{BaseObject: &pkgservice.BaseObject{ID: tUserNameA.ID}},
			args: args{tUserNameA.ID},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			u.SetDB(dbAccount, tLockMap, tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
			if err := u.Delete(true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			newU := &UserName{BaseObject: &pkgservice.BaseObject{ID: tt.args.id}}
			newU.SetDB(dbAccount, tLockMap, tEntityID, DBUserNamePrefix, DBUserNameIdxPrefix, nil, nil)
			err := newU.Get(true)
			if err != leveldb.ErrNotFound {
				t.Errorf("UserName.Delete() unable to delete: id: %v newU: %v e: %v", tt.args.id, newU, err)
			}
		})
	}

	// teardown test
}
