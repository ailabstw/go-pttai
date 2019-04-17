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

package me

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/key"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Config struct {
	DataDir string

	PrivateKey *ecdsa.PrivateKey `toml:"-"`
	ID         *types.PttID      `toml:"-"` // we also need ID because other services need to know ID, but cannot directly acccess private-key and postfix.
	Postfix    string
}

func (c *Config) SetMyKey(hex string, file string, postfix string, isSave bool) error {
	var (
		key *ecdsa.PrivateKey
		id  *types.PttID
		err error
	)

	switch {
	case file != "" && hex != "":
		return ErrInvalidPrivateKeyFileHex
	case file != "":
		if key, err = crypto.LoadECDSA(file); err != nil {
			return ErrInvalidPrivateKeyFile
		}
		c.PrivateKey = key
	case hex != "":
		if key, err = crypto.HexToECDSA(hex); err != nil {
			return ErrInvalidPrivateKeyHex
		}
		c.PrivateKey = key
	}

	if postfix != "" {
		c.Postfix = postfix
	}

	key, postfix, id, err = c.myKey()
	if err != nil {
		return err
	}

	c.PrivateKey = key
	c.Postfix = postfix
	c.ID = id

	if isSave {
		c.saveKeyFile(DataDirPrivateKey, key, postfix, id)
	}

	return nil

}

func (c *Config) saveKeyFile(DataDirPrivateKey string, key *ecdsa.PrivateKey, postfix string, id *types.PttID) error {

	// save DataDirPrivKey
	keyfile := c.ResolvePath(DataDirPrivateKey)
	if err := c.SaveKey(keyfile, key, postfix); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return err
	}

	// save DataDirPrivKeyWithID
	keyfile, err := c.ResolvePrivateKeyWithIDPath(id)
	if err != nil {
		return err
	}
	if err := c.SaveKey(keyfile, key, postfix); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return err
	}

	return nil
}

func (c *Config) myKey() (*ecdsa.PrivateKey, string, *types.PttID, error) {
	// Use any specifically configured key.
	if c.PrivateKey != nil && c.Postfix == "" {
		return nil, "", nil, ErrInvalidMe
	}

	if c.PrivateKey == nil && c.Postfix != "" {
		return nil, "", nil, ErrInvalidMe
	}

	if c.PrivateKey != nil {
		id, err := types.NewPttIDFromKeyPostfix(c.PrivateKey, []byte(c.Postfix))
		if err != nil {
			return nil, "", nil, ErrInvalidMe
		}

		return c.PrivateKey, c.Postfix, id, nil
	}

	// Generate ephemeral key if no datadir is being used.
	if c.DataDir == "" {
		key, err := key.GenerateKey()
		if err != nil {
			log.Crit(fmt.Sprintf("Failed to generate ephemeral node key: %v", err))
			return nil, "", nil, ErrInvalidMe
		}

		id, err := types.NewPttIDFromKey(key)
		if err != nil {
			return nil, "", nil, ErrInvalidMe
		}

		postfix := string(id[common.AddressLength:])

		return key, postfix, id, nil
	}

	// retrieve key / id from file
	keyfile := c.ResolvePath(DataDirPrivateKey)
	key, err := crypto.LoadECDSA(keyfile)
	postfixBytes, err2 := ioutil.ReadFile(keyfile + ".postfix")
	if err == nil && err2 == nil {
		id, err := types.NewPttIDFromKeyPostfix(key, postfixBytes)
		if err != nil {
			return nil, "", nil, ErrInvalidMe
		}

		return key, string(postfixBytes), id, nil
	}

	log.Warn(fmt.Sprintf("Failed to load key: %v. create a new one.", err))
	// No persistent key found, generate and store a new one.
	key, err = key.GenerateKey()
	if err != nil {
		log.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
		return nil, "", nil, ErrInvalidMe
	}

	id, err := types.NewPttIDFromKey(key)
	if err != nil {
		return nil, "", nil, ErrInvalidMe
	}
	postfix := string(id[common.AddressLength:])

	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return nil, "", nil, err
	}

	err = c.saveKeyFile(DataDirPrivateKey, key, postfix, id)
	if err != nil {
		return nil, "", nil, err
	}

	return key, postfix, id, err
}

func (c *Config) GetDataPrivateKeyByID(myID *types.PttID) (*ecdsa.PrivateKey, error) {
	keyfile, err := c.ResolvePrivateKeyWithIDPath(myID)
	if err != nil {
		return nil, err
	}
	return crypto.LoadECDSA(keyfile)
}

func (c *Config) ResolvePrivateKeyWithIDPath(myID *types.PttID) (string, error) {
	idBytes, err := myID.MarshalText()
	if err != nil {
		return "", err
	}
	idStr := string(idBytes)

	dataDirPrivateKeyPostfix := DataDirPrivateKey + "." + idStr

	keyfile := c.ResolvePath(dataDirPrivateKeyPostfix)

	return keyfile, nil
}

// resolvePath resolves path in the instance directory.
func (c *Config) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, path)
}

func (c *Config) SaveKey(filename string, key *ecdsa.PrivateKey, postfix string) error {
	err := crypto.SaveECDSA(filename, key)
	if err != nil {
		return err
	}

	postfixFilename := filename + ".postfix"
	err = ioutil.WriteFile(postfixFilename, []byte(postfix), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) LoadKey(filename string) (*ecdsa.PrivateKey, *types.PttID, error) {
	key, err := crypto.LoadECDSA(filename)
	if err != nil {
		return nil, nil, err
	}

	id, err := types.NewPttIDFromKey(key)
	if err != nil {
		return nil, nil, ErrInvalidMe
	}

	return key, id, nil
}

func (c *Config) DeleteKey() error {
	keyfile := c.ResolvePath(DataDirPrivateKey)

	tsStr := time.Now().UTC().Format("2006-01-02_15-04-05.000")

	deleteFile := keyfile + "." + tsStr + ".deleted"

	log.Warn("to Remove keyfile", "keyfile", keyfile, "deleteFile", deleteFile)

	return os.Rename(keyfile, deleteFile)
}

func (c *Config) RevokeMyKey(myID *types.PttID) error {
	keyfile, err := c.ResolvePrivateKeyWithIDPath(myID)
	if err != nil {
		return err
	}

	os.Remove(keyfile)

	postfixFilename := keyfile + ".postfix"
	os.Remove(postfixFilename)

	return nil
}

func (c *Config) RevokeKey() error {
	keyfile := c.ResolvePath(DataDirPrivateKey)

	log.Warn("to Revoke keyfile", "keyfile", keyfile)

	return os.Remove(keyfile)
}
