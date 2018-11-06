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

package account

import (
	"reflect"
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestNewUserName(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	// define test-structure
	type args struct {
		id *types.PttID
		ts types.Timestamp
	}

	// prepare test-cases
	tests := []struct {
		name    string
		args    args
		want    *UserName
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{id: tUserIDA, ts: tTsA},
			want: tUserNameA,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserName(tt.args.id, tt.args.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUserName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserName() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

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
			want: []byte("{\"V\":2,\"ID\":\"f8FnBNeGR37bqtFqZ4zZjXGYdKpoyDbWLRrv8qRSevKjQzeWpdeX46\",\"CT\":{\"T\":1,\"NT\":5},\"UT\":{\"T\":1,\"NT\":5},\"S\":0,\"N\":null,\"bID\":null,\"l\":null}"),
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
			args: args{[]byte("{\"V\":2,\"ID\":\"f8FnBNeGR37bqtFqZ4zZjXGYdKpoyDbWLRrv8qRSevKjQzeWpdeX46\",\"CT\":{\"T\":1,\"NT\":5},\"UT\":{\"T\":1,\"NT\":5},\"S\":0,\"N\":null}")},
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
			if !reflect.DeepEqual(u, tt.want) {
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
		id *types.PttID
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
			args: args{id: tUserIDA},
			want: tUserNameA,
		},
		{
			u:    &UserName{},
			args: args{id: tUserIDB},
			want: tUserNameB,
		},
		{
			u:    &UserName{},
			args: args{id: tUserIDC},
			want: tUserNameC,
		},
		{
			u:       &UserName{},
			args:    args{id: tUserIDD},
			want:    &UserName{ID: tUserIDD},
			wantErr: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			if err := u.Get(tt.args.id, true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(u, tt.want) {
				t.Errorf("UserName.Get() u = %v, want %v", u, tt.want)
			}
		})
	}

	// teardown test
}

func TestUserName_Update(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tUserNameA.Save(true)

	origUserNameA := &UserName{}
	newUserNameA := &UserName{}

	origUserNameA.Get(tUserNameA.ID, true)
	newUserNameA.Get(tUserNameA.ID, true)

	newUserNameA.Name = []byte("test2")

	// define test-structure
	type args struct {
		name []byte
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
			u:    tUserNameA,
			args: args{[]byte("test2")},
			want: newUserNameA,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u

			if err := u.Update(tt.args.name, true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			// update-ts is updated
			tt.want.UpdateTS = u.UpdateTS

			if !reflect.DeepEqual(u, tt.want) {
				t.Errorf("UserName.Update() u = %v, want %v", u, tt.want)
			}

			loadUserName := &UserName{}
			loadUserName.Get(u.ID, true)
			if !reflect.DeepEqual(loadUserName, tt.want) {
				t.Errorf("UserName.Update() loadUserName = %v, want %v", loadUserName, tt.want)
			}
		})
	}
	tUserNameA = origUserNameA
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
			u:    &UserName{},
			args: args{tUserNameA.ID},
		},
		{
			u:    &UserName{},
			args: args{tUserNameA.ID},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u

			if err := u.Delete(tt.args.id, true); (err != nil) != tt.wantErr {
				t.Errorf("UserName.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			newU := &UserName{}
			err := newU.Get(tt.args.id, true)
			if err != leveldb.ErrNotFound {
				t.Errorf("UserName.Delete() unable to delete: id: %v newU: %v e: %v", tt.args.id, newU, err)
			}
		})
	}

	// teardown test
}

func Test_isValidName(t *testing.T) {
	// setup test

	// define test-structure
	type args struct {
		name []byte
	}

	// prepare test-cases
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			args: args{[]byte("01234567891123456789223456789323456789423456789")},
			want: false,
		},
		{
			args: args{[]byte("01234567891123456789")},
			want: true,
		},
		{
			args: args{[]byte("零一二三四五六七八九十壹貳參肆伍陸柒捌玖")},
			want: true,
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidName(tt.args.name); got != tt.want {
				t.Errorf("isValidName() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}

func TestUserName_GetList(t *testing.T) {
	// setup test
	setupTest(t)
	defer teardownTest(t)

	tUserNameA.Save(true)
	tUserNameB.Save(true)
	tUserNameC.Save(true)

	// define test-structure
	type fields struct {
		V        types.Version
		ID       *types.PttID
		CreateTS types.Timestamp
		UpdateTS types.Timestamp
		Name     []byte
	}
	type args struct {
		id    *types.PttID
		limit int
	}

	// prepare test-cases
	tests := []struct {
		name    string
		u       *UserName
		args    args
		want    []*UserName
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			u:    &UserName{},
			args: args{id: &types.EmptyID, limit: 4},
			want: []*UserName{tUserNameA, tUserNameB, tUserNameC},
		},
		{
			u:    &UserName{},
			args: args{id: &types.EmptyID, limit: 2},
			want: []*UserName{tUserNameA, tUserNameB},
		},
		{
			u:    &UserName{},
			args: args{id: &types.EmptyID, limit: 3},
			want: []*UserName{tUserNameA, tUserNameB, tUserNameC},
		},
		{
			u:    &UserName{},
			args: args{id: tUserIDA, limit: 3},
			want: []*UserName{tUserNameA, tUserNameB, tUserNameC},
		},
		{
			u:    &UserName{},
			args: args{id: tUserIDB, limit: 3},
			want: []*UserName{tUserNameB, tUserNameC},
		},
		{
			u:    &UserName{},
			args: args{id: tUserIDC, limit: 3},
			want: []*UserName{tUserNameC},
		},
	}

	// run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.u
			got, err := u.GetList(tt.args.id, tt.args.limit, pttdb.ListOrderNext)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserName.GetList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserName.GetList() = %v, want %v", got, tt.want)
			}
		})
	}

	// teardown test
}
