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

package main

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/ailabstw/go-pttai/node"
	"github.com/naoina/toml"
	cli "gopkg.in/urfave/cli.v1"
)

type Config struct {
	Node *node.Config
}

func NewConfig(ctx *cli.Context) (*Config, error) {
	return &Config{
		Node: &node.DefaultConfig,
	}, nil
}

func loadConfig(ctx *cli.Context) (*Config, error) {
	cfg := DefaultConfig

	filename := ctx.GlobalString(configFileFlag.Name)
	if filename == "" {
		return &cfg, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return &cfg, nil
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg)
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(filename + ", " + err.Error())
	}

	return &cfg, err
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	cfg, err := NewConfig(ctx)
	if err != nil {
		return err
	}

	comment := ""

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
