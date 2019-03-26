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
	"crypto/ecdsa"
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
)

const ()

const (
	_ OpType = iota
	tDefaultOpType
)

type TType struct {
	A string
	B string
}

var (
	tDefaultTimestamp = types.Timestamp{Ts: 123456789, NanoTs: 2}

	tDefaultPtt = &BasePtt{
		myNodeID: tDefaultNodeID,
	}

	tMyKey, _      = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	tMyID, _       = types.NewPttIDFromKeyPostfix(tMyKey, []byte("0123456789abcdefghij"))
	tDefaultSPM, _ = NewBaseServiceProtocolManager(tDefaultPtt, nil)

	tDefaultKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	tDefaultID, _  = types.NewPttIDFromKeyPostfix(tDefaultKey, []byte("0123456789abcdefghij"))

	tDefaultKeyInfo = &KeyInfo{
		Key:         tDefaultKey,
		KeyBytes:    crypto.FromECDSA(tDefaultKey),
		PubKeyBytes: crypto.FromECDSAPub(&tDefaultKey.PublicKey),
	}

	tDefaultHash   = crypto.PubkeyToAddress(tDefaultKey.PublicKey)
	tDefaultNodeID = &discover.NodeID{}

	tDefaultData = TType{
		A: "test",
		B: "test2",
	}
	tDefaultDataBytes, _        = json.Marshal(tDefaultData)
	tDefaultOp           OpType = 3

	tDefaultEncData = []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 69, 69, 177, 160,
		222, 143, 223, 115, 69, 155, 171, 227, 29, 169,
		45, 135, 184, 129, 253, 45, 107, 99, 215, 47,
		175, 41, 199, 197, 220, 16, 143, 52,
	}

	tDefaultPttData = &PttData{
		Node:       nil,
		Code:       CodeTypeOp,
		Hash:       []byte{113, 86, 43, 113, 153, 152, 115, 219, 91, 40, 109, 249, 87, 175, 25, 158, 201, 70, 23, 247},
		EvWithSalt: []byte("{\"C\":4,\"H\":\"cVYrcZmYc9tbKG35V68ZnslGF/c=\",\"D\":\"AAECAwQFBgcICQoLDA0ODwAAAAN7IkEiOiJ0ZXN0IiwiQiI6InRlc3QyIn0EBAQE\"}01234567890123456789012345678901"),
		Checksum: []byte{
			204, 211, 90, 234, 168, 48, 66, 22, 62, 144,
			66, 158, 184, 174, 58, 199, 120, 241, 68, 163,
			239, 137, 107, 252, 222, 160, 111, 124, 52, 247,
			241, 58,
		},
		Relay: 0,
	}

	tDefaultMerkleNode = &MerkleNode{
		Level: MerkleTreeLevelHR,
		Addr: []byte{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		},
		UpdateTS:  types.Timestamp{Ts: 1234567890, NanoTs: 123456789},
		NChildren: 13,
		Key:       []byte{},
	}
	tDefaultMerkleNodeBytes = []byte("\"AgABAgMEBQYHCAkKCwwNDg8QERITAAAAAEmWAtIHW80VAAAADQ==\"")

	tDefaultOplog          *BaseOplog = nil
	tDefaultTimestamp1                = types.Timestamp{Ts: 1234567890, NanoTs: 0}
	tDefaultOplogMerkleKey            = []byte{
		46, 116, 116, 109, 107,
		113, 86, 43, 113, 153, 152, 115, 219, 91, 40,
		109, 249, 87, 175, 25, 158, 201, 70, 23, 247,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
		1,
		0, 0, 0, 0, 73, 150, 2, 210, 0, 0, 0, 0,
		246, 98, 165, 238, 232, 42, 189, 244, 74, 45,
		11, 117, 251, 24, 13, 175, 72, 167, 158, 224,
		177, 13, 57, 70, 81, 133, 15, 212, 161, 120,
		137, 46, 226, 133, 236, 225, 81, 20, 85, 120,
		0, 0, 0, 1,
	}
	tDefaultMerkleNode1Now = &MerkleNode{
		Level: MerkleTreeLevelNow,
		Addr: []byte{
			109, 106, 68, 249, 1, 105, 182, 107, 234, 50,
			192, 218, 27, 70, 8, 195, 36, 191, 156, 238,
		},
		UpdateTS:  types.Timestamp{Ts: 1234567890, NanoTs: 0},
		NChildren: 0,
		Key: []byte{
			46, 116, 116, 108, 103,
			113, 86, 43, 113, 153, 152, 115, 219, 91, 40,
			109, 249, 87, 175, 25, 158, 201, 70, 23, 247,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
			0, 0, 0, 0, 73, 150, 2, 210, 0, 0, 0, 0,
			246, 98, 165, 238, 232, 42, 189, 244, 74, 45,
			11, 117, 251, 24, 13, 175, 72, 167, 158, 224,
			177, 13, 57, 70, 81, 133, 15, 212, 161, 120,
			137, 46, 226, 133, 236, 225, 81, 20, 85, 120,
			0, 0, 0, 1,
		},
	}

	tDefaultOplog2          *BaseOplog = nil
	tDefaultTimestamp2                 = types.Timestamp{Ts: 1234567891, NanoTs: 0}
	tDefaultOplog2MerkleKey            = []byte{
		46, 116, 116, 109, 107,
		113, 86, 43, 113, 153, 152, 115, 219, 91, 40,
		109, 249, 87, 175, 25, 158, 201, 70, 23, 247,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
		1,
		0, 0, 0, 0, 73, 150, 2, 211, 0, 0, 0, 0,
		8, 117, 214, 78, 226, 211, 208, 208, 222, 107,
		248, 249, 180, 76, 232, 95, 240, 68, 198, 177,
		248, 59, 142, 136, 59, 191, 133, 122, 171, 153,
		197, 178, 82, 199, 66, 156, 50, 243, 168, 174,
		0, 0, 0, 1,
	}
	tDefaultMerkleNode2Now = &MerkleNode{
		Level: MerkleTreeLevelNow,
		Addr: []byte{
			204, 115, 234, 18, 43, 72, 250, 130, 156, 41,
			190, 100, 149, 69, 92, 238, 118, 154, 246, 82,
		},
		UpdateTS:  types.Timestamp{Ts: 1234567891, NanoTs: 0},
		NChildren: 0,
		Key: []byte{
			46, 116, 116, 108, 103,
			113, 86, 43, 113, 153, 152, 115, 219, 91, 40,
			109, 249, 87, 175, 25, 158, 201, 70, 23, 247,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
			0, 0, 0, 0, 73, 150, 2, 211, 0, 0, 0, 0,
			8, 117, 214, 78, 226, 211, 208, 208, 222, 107,
			248, 249, 180, 76, 232, 95, 240, 68, 198, 177,
			248, 59, 142, 136, 59, 191, 133, 122, 171, 153,
			197, 178, 82, 199, 66, 156, 50, 243, 168, 174,
			0, 0, 0, 1,
		},
	}

	tDefaultMerkleNodeDay = &MerkleNode{
		Level: MerkleTreeLevelDay,
		Addr: []byte{
			204, 99, 169, 130, 205, 128, 96, 0, 2, 125,
			192, 198, 207, 136, 119, 152, 20, 207, 118, 168,
		},
		UpdateTS:  types.Timestamp{Ts: 1234483200, NanoTs: 0},
		NChildren: 1,
		Key:       []byte{},
	}

	tDefaultMerkleNodeMonth = &MerkleNode{
		Level: MerkleTreeLevelMonth,
		Addr: []byte{
			175, 138, 104, 31, 200, 172, 68, 11, 101, 23,
			0, 182, 43, 40, 66, 255, 173, 90, 227, 26,
		},
		UpdateTS:  types.Timestamp{Ts: 1233446400, NanoTs: 0},
		NChildren: 1,
		Key:       []byte{},
	}

	tDefaultMerkleNodeYear = &MerkleNode{
		Level: MerkleTreeLevelYear,
		Addr: []byte{
			104, 254, 151, 251, 87, 216, 126, 172, 249, 32,
			67, 2, 36, 209, 51, 142, 133, 53, 58, 70,
		},
		UpdateTS:  types.Timestamp{Ts: 1230768000, NanoTs: 0},
		NChildren: 1,
		Key:       []byte{},
	}

	tDefaultMerkle *Merkle = nil

	tDBOplog             *pttdb.LDBBatch    = nil
	tDBOplogCore         *pttdb.LDBDatabase = nil
	tDBOplogPrefix                          = []byte(".ttlg")
	tDBOplogIdxPrefix                       = []byte(".ttig")
	tDBOplogMerklePrefix                    = []byte(".ttmk")

	origHandler      log.Handler
	origGenIV        func(iv []byte) error
	origGetTimestamp func() (types.Timestamp, error)
	origNewSalt      func() (*types.Salt, error)
	origRandRead     func(b []byte) (int, error)

	tKeyMe     *ecdsa.PrivateKey
	tUserIDMe  *types.PttID
	tTsMe      types.Timestamp
	tKeyInfoMe *KeyInfo

	tDBCountPrefix = []byte(".tcct")
	tPCount        = uint(12)

	tDBLock *types.LockMap

	tDefaultSignKeyBytes2 = []byte{
		208, 12, 190, 67, 195, 69, 44, 203, 240, 208,
		85, 102, 103, 66, 169, 55, 233, 240, 247, 167,
		251, 169, 18, 202, 108, 162, 116, 56, 4, 245,
		18, 96,
	}
	tDefaultSignKey2     *ecdsa.PrivateKey
	tDefaultSignKeyInfo2 *KeyInfo
	tDefaultDoerID2      *types.PttID

	tDefaultPubBytes2 = []byte{
		4, 59, 103, 1, 40, 236, 82, 148, 94, 111,
		140, 111, 130, 32, 57, 126, 118, 177, 147, 242,
		216, 88, 47, 237, 207, 190, 59, 209, 13, 226,
		90, 235, 57, 240, 64, 19, 97, 112, 216, 111,
		118, 249, 245, 187, 204, 66, 153, 73, 26, 221,
		187, 230, 61, 28, 46, 119, 49, 168, 78, 205,
		84, 119, 201, 39, 80,
	}

	tDefaultSig2 = []byte{
		243, 41, 18, 66, 195, 172, 71, 33, 91, 184,
		42, 251, 78, 117, 210, 90, 29, 166, 72, 69,
		2, 91, 186, 215, 212, 212, 248, 42, 81, 166,
		59, 174, 121, 198, 120, 191, 243, 246, 3, 38,
		41, 211, 42, 1, 255, 99, 103, 25, 244, 85,
		100, 150, 249, 173, 110, 108, 174, 119, 183, 80,
		254, 125, 10, 73, 1,
	}

	tDefaultHash2 = []byte{
		222, 42, 214, 145, 30, 223, 156, 203, 141, 156,
		82, 105, 150, 57, 95, 19, 153, 115, 48, 148,
		71, 140, 233, 219, 130, 80, 109, 242, 168, 24,
		26, 14,
	}

	tDefaultSalt2 = []byte{
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49,
	}

	tDefaultBytes2 = []byte(`{"V":1,"ID":"CeUiKhsBZPYj4JnsqfMYf3t6qju8YZHbotZfMyAp2mEKTvVEj3rkb6N","DID":"62Rm6MPdZs5WqccFHRNmRGDN6tgixqD888r6BXYSbsmQdvJtUhsr4E9","CT":{"T":1537239289,"NT":995126888},"OID":"BnaEYdCPBcDoqpBej7qzXYf5dR2iTufnv5N5MAFosuwohJ6Ag3QAkXm","O":2,"D":{"ms":null,"TH":"VUNkygal4IkEs7xk0mqhTyjO+BY=","bID":"AhvFYMttSMQh6t73GmKe8sKWN4GFsqR6zfJ8mNhxdB2fHQN1J9C9Cxj","NB":3,"H":[["7apLl20Cewyz3LKzb34DwiYU06udzf7NJLPEAQeHT5o=","Zf9Ac6acyTS9WMi5x8XIqGOFuIxe0zauuxuQvQ15Hxc="],["OCN78izVp2A0Svi0imLei2GmKIxhHz6X8bho9CZNrh4=","S6Gf6LQFEBp2TUnRnb06gqIPDx5Wr8eJql5+CtiDsAc="],["NPrr0uM+f6n0aMofaK6nSL3NhwgBy/85pOH7AcpaXyU=","Cd+N58buKvU2+adZe6LY50UnTnSJsWC0WHCN6EBx2fk="]]},"dH":null,"s":"11111111111111111111111111111111","S":null,"K":null,"UT":{"T":0,"NT":0},"H":null,"y":0}`)

	tDefaultBytesWithSalt2 = append(tDefaultBytes2, tDefaultSalt2...)

	tDefaultBytes3 = []byte(`{"V":1,"ID":"7k6CP6W7ScacvsnGcvKr1fyXtT23wRRKqG3QzMoarDaWjFoRsCSsbhs","DID":"7WsZbk3UedLyfMmF9zfcEKPYL19AjccdLepGC1wWHyvoKxJtMaQ89s","CT":{"T":1537240960,"NT":672491197},"OID":"AF72G4oxNV8eLsVom2o2DyvW41XLzZhnPSUnkSUmYFa4Ai1LTm97dvp","O":2,"D":{"H":[["uFkT1uGFkxUM6LhyQjj2Yia8FDgVrpHYvl+EcyWnd4c=","QWsL3xTsHYoJUN4g4VJkJZeRvCcWyYQtutm9YiCd29w="],["gjvq1pATPzT4VO+T5C+SRoiArlexB8K6lVY07G1Do30=","yao0B8nqBs0/3mWP1hu/zxJEkNjqHhC8KI/v5s/CJNQ="],["IQLpkO7ntCHOhy7WVAnuBZbHEbEgw9x9gfKQGmdPvfw=","qlreQk+vuKT/4J0pDUfXhg27RdQrBT+4Ciigcj8HiaA="]],"NB":3,"TH":"VUNkygal4IkEs7xk0mqhTyjO+BY=","bID":"Ap3cULQ8TK2uGqgbBWJbnEFjnpjPTPA9VjzV2W5J3E52DgtzLZXyxxY","ms":null},"dH":null,"s":"11111111111111111111111111111111","S":null,"K":null,"UT":{"T":0,"NT":0},"H":null,"y":0}`)
	tDefaultSalt3  = []byte{
		3, 224, 238, 136, 242, 71, 154, 49, 28, 102,
		106, 194, 8, 213, 9, 155, 17, 119, 178, 33,
		204, 133, 160, 85, 112, 218, 210, 252, 224, 34,
		109, 198,
	}

	tDefaultBytesWithSalt3 = append(tDefaultBytes3, tDefaultSalt3...)

	tDefaultSig3 = []byte{
		41, 236, 233, 222, 219, 250, 200, 133, 246, 13,
		179, 238, 42, 123, 50, 120, 96, 229, 224, 72,
		237, 162, 2, 173, 21, 78, 39, 24, 33, 23,
		116, 84, 61, 232, 41, 11, 227, 197, 220, 163,
		11, 110, 190, 103, 254, 155, 171, 64, 0, 160,
		117, 126, 114, 135, 89, 222, 192, 82, 211, 166,
		33, 24, 54, 173, 0,
	}

	tDefaultPubBytes3 = []byte{
		4, 243, 200, 196, 101, 2, 62, 57, 80, 121,
		4, 25, 202, 188, 252, 237, 71, 44, 60, 39,
		116, 181, 83, 139, 62, 227, 37, 220, 113, 172,
		238, 228, 104, 182, 174, 191, 191, 183, 207, 101,
		173, 52, 213, 29, 218, 90, 28, 38, 56, 8,
		25, 110, 212, 97, 196, 115, 26, 5, 7, 229,
		6, 72, 100, 168, 43,
	}

	tDefaultBytes4 = []byte(`{"V":2,"ID":"4UCUTWaPy1cY6x1yLcrVoL9Ee9Mp9HxCUaQwwHaXsmkRCKhrH2Kjuf6","CID":"UjyjA6W1GqF1xEQUX12QEUnxkHN1jvnYwJ7PkMZ3Tsrypd9eS1mmVE","CT":{"T":1551933027,"NT":564188767},"OID":"9pC8vTZkqNiFYutRKJC5ejZPGwejv6Lvq1YHzGQJvM2Fc7m2Qz89iJo","O":6,"D":{"BID":"3VaJvs24k3ZEvoRdttDNm98nU7YtYY7J84W5XpadUc12Qt555tqNEVh","H":[["JPqKXbHc102G5x9IC8KAhcqdDDZ3ELEfRGcSeIDSOBg=","RVfiPh27UhJOMjA+m89PJNJAGt3V92oviRezc+NSb+s="],["cPACVkCQuLJalXbpXr883ZIO1S11yHpm1IN2legnaCI=","KC3mGeH469HC5gRCSFEocAxNS2xJn57wiU0Y7ovbhCE="]],"NB":2,"th":"fHBGiGWLHupWfaqsvtRhIEd1BGo="},"s":"11111111111111111111111111111111","UT":{"T":0,"NT":0},"y":0}`)

	tDefaultSalt4 = []byte{
		206, 141, 120, 239, 75, 12, 176, 227, 67, 3,
		171, 171, 231, 10, 80, 97, 238, 34, 168, 214,
		232, 33, 212, 87, 121, 110, 35, 79, 205, 150,
		106, 98,
	}

	tDefaultBytesWithSalt4 = append(tDefaultBytes4, tDefaultSalt4...)

	tDefaultHash4 = []byte{
		54, 150, 125, 2, 35, 238, 75, 190, 241, 244,
		85, 61, 141, 250, 119, 55, 151, 184, 170, 20,
		178, 245, 10, 49, 55, 46, 43, 0, 65, 172,
		176, 2,
	}

	tDefaultSig4 = []byte{
		161, 187, 4, 84, 22, 149, 231, 143, 86, 68,
		90, 161, 11, 50, 225, 62, 115, 244, 167, 202,
		45, 2, 175, 185, 238, 41, 187, 245, 60, 198,
		46, 228, 22, 22, 32, 34, 10, 211, 67, 28,
		124, 107, 190, 173, 111, 205, 136, 21, 17, 238,
		237, 66, 201, 111, 210, 0, 125, 208, 163, 195,
		6, 2, 79, 84, 0,
	}

	tDefaultPubBytes4 = []byte{
		4, 106, 78, 173, 229, 144, 41, 189, 35, 235,
		160, 58, 58, 93, 7, 70, 182, 234, 226, 235,
		89, 66, 145, 108, 208, 221, 71, 118, 219, 153,
		27, 212, 186, 75, 201, 213, 6, 193, 21, 15,
		42, 166, 195, 35, 133, 129, 234, 245, 83, 85,
		99, 207, 147, 251, 98, 204, 242, 69, 40, 43,
		49, 129, 169, 113, 186,
	}

	tDefaultDoerID4 *types.PttID

	tDefaultBuf = [][]byte{
		[]byte("12345"),
		[]byte("6789"),
	}
	tDefaultScrambleBuf = [][]byte{
		[]byte{169, 250, 3, 58, 179, 16, 112, 182, 133, 3, 220, 116},
		[]byte{208, 150, 36, 13, 206, 37, 111, 207, 237, 28, 163, 88},
	}

	tDefaultBuf2 = [][]byte{
		[]byte("1"),
		[]byte(""),
	}
	tDefaultScrambleBuf2 = [][]byte{
		[]byte{169, 62, 147, 183, 253, 71},
		[]byte{208, 48, 252, 196, 157, 122},
	}

	tDefaultBuf3 = [][]byte{
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte(""),
	}
	tDefaultScrambleBuf3 = [][]byte{
		[]byte{111, 228, 65, 65, 185, 185, 135, 155, 135, 135, 121, 195},
		[]byte{117, 135, 65, 65, 197, 197, 228, 246, 228, 228, 101, 130},
	}

	tDefaultBuf4 = [][]byte{
		[]byte{},
		[]byte{},
	}
	tDefaultScrambleBuf4 = [][]byte{
		[]byte{169, 34, 220, 116},
		[]byte{208, 34, 163, 88},
	}
)

func setupTest(t *testing.T) {
	origHandler = log.Root().GetHandler()
	log.Root().SetHandler(log.Must.FileHandler("log.tmp.txt", log.TerminalFormat(true)))

	origGetTimestamp = types.GetTimestamp
	types.GetTimestamp = func() (types.Timestamp, error) {
		return tDefaultTimestamp, nil
	}

	origNewSalt = types.NewSalt
	types.NewSalt = func() (*types.Salt, error) {
		return &types.Salt{
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49,
		}, nil
	}

	origGenIV = genIV
	genIV = func(iv []byte) error {
		for i := 0; i < len(iv); i++ {
			iv[i] = uint8(i % 0xff)
		}

		return nil
	}

	origRandRead = types.RandRead
	rand.Seed(0)
	types.RandRead = func(b []byte) (int, error) {
		return rand.Read(b)
	}

	tDBLock, _ = types.NewLockMap(10)

	tKeyMe, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	tUserIDMe, _ = types.NewPttIDFromKey(tKeyMe)
	tTsMe = types.Timestamp{Ts: 123456789, NanoTs: 5}
	tKeyInfoMe = &KeyInfo{
		Key:         tKeyMe,
		KeyBytes:    crypto.FromECDSA(tKeyMe),
		PubKeyBytes: crypto.FromECDSAPub(&tKeyMe.PublicKey),
	}

	tDBOplogCore, _ = pttdb.NewLDBDatabase("oplog", "./test.out", 0, 0)
	tDBOplog, _ = pttdb.NewLDBBatch(tDBOplogCore)
	tDefaultOplog, _ = NewOplog(tDefaultID, tDefaultTimestamp1, tMyID, tDefaultOpType, nil, tDBOplog, tDefaultID, tDBOplogPrefix, tDBOplogIdxPrefix, tDBOplogMerklePrefix, tDBLock)
	tDefaultOplog.Sign(tKeyInfoMe)
	tDefaultOplog.MasterLogID = tUserIDMe
	tDefaultOplog2, _ = NewOplog(tDefaultID, tDefaultTimestamp2, tMyID, tDefaultOpType, nil, tDBOplog, tDefaultID, tDBOplogPrefix, tDBOplogIdxPrefix, tDBOplogMerklePrefix, tDBLock)
	tDefaultOplog2.Sign(tKeyInfoMe)
	tDefaultOplog2.MasterLogID = tUserIDMe
	tDefaultMerkle, _ = NewMerkle(tDBOplogPrefix, tDBOplogMerklePrefix, tDefaultID, tDBOplog, "default")

	tDefaultSignKey2, _ = crypto.ToECDSA(tDefaultSignKeyBytes2)
	tDefaultSignHash2 := crypto.PubkeyToAddress(tDefaultSignKey2.PublicKey)
	tDefaultSignID2 := keyInfoHashToID(&tDefaultSignHash2)
	tDefaultDoerID2, _ = types.NewPttIDFromKey(tDefaultSignKey2)
	ts, _ := types.GetTimestamp()
	tDefaultSignKeyInfo2 = &KeyInfo{
		BaseObject:  NewObject(tDefaultSignID2, ts, tDefaultDoerID2, nil, nil, types.StatusAlive),
		Key:         tDefaultSignKey2,
		KeyBytes:    tDefaultSignKeyBytes2,
		PubKeyBytes: crypto.FromECDSAPub(&tDefaultSignKey2.PublicKey),
	}

	ts, _ = types.GetTimestamp()
	t.Logf("after setup: GetTimestamp: %v", ts)

	pubKey, _ := crypto.UnmarshalPubkey(tDefaultPubBytes4)
	addr := crypto.PubkeyToAddress(*pubKey)
	tDefaultDoerID4 = &types.PttID{}
	copy(tDefaultDoerID4[:], addr[:])

}

func teardownTest(t *testing.T) {
	log.Root().SetHandler(origHandler)
	types.GetTimestamp = origGetTimestamp
	types.NewSalt = origNewSalt
	genIV = origGenIV
	types.RandRead = origRandRead

	rand.Seed(time.Now().UnixNano())

	if tDBOplogCore != nil {
		tDBOplogCore.Close()
		tDBOplogCore = nil
	}
	if tDBOplog != nil {
		tDBOplog = nil
	}

	if tDBLock != nil {
		tDBLock = nil
	}

	os.RemoveAll("./test.out")

	ts, _ := types.GetTimestamp()
	t.Logf("after teardown: GetTimestamp: %v", ts)
}
