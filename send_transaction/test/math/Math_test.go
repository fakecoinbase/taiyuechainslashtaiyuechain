package math

import (
	"encoding/hex"
	"fmt"
	taicert "github.com/taiyuechain/taiyuechain/cert"
	"github.com/taiyuechain/taiyuechain/accounts/abi/bind"
	"github.com/taiyuechain/taiyuechain/accounts/abi/bind/backends"
	"github.com/taiyuechain/taiyuechain/common"
	"github.com/taiyuechain/taiyuechain/core"
	"github.com/taiyuechain/taiyuechain/core/types"
	"github.com/taiyuechain/taiyuechain/crypto"
	"github.com/taiyuechain/taiyuechain/log"
	"github.com/taiyuechain/taiyuechain/params"
	"golang.org/x/crypto/sha3"
	"math/big"
	"os"
	"testing"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

var (
	pbft1Name = "pbft1priv"
	p2p1Name  = "p2p1cert"
	pbft1path = "../../../cim/testdata/testcert/" + pbft1Name + ".pem"
	p2p1path  = "../../../cim/testdata/testcert/" + p2p1Name + ".pem"

	gspec = DefaulGenesisBlock()

	//p2p 1
	priKey, _ = crypto.HexToECDSA("41c8bcf352894b132db095b0ef67b1c7ea9f4d7afd72a36b16c62c9fc582a5df")
	// p2p 2
	skey1, _ = crypto.HexToECDSA("200854f6bdcd2f94ecf97805ec95f340026375b347a6efe6913d5287afbabeed")
	// pbft 1
	dkey1, _ = crypto.HexToECDSA("8c2c3567667bf29509afabb7e1178e8a40a849b0bd22e0455cff9bab5c97a247")
	mAccount = crypto.PubkeyToAddress(priKey.PublicKey)
	saddr1   = crypto.PubkeyToAddress(skey1.PublicKey)
	daddr1   = crypto.PubkeyToAddress(dkey1.PublicKey)

	p2p1Byte, _  = taicert.ReadPemFileByPath(p2p1path)
	pbft1Byte, _ = taicert.ReadPemFileByPath(pbft1path)
)

func DefaulGenesisBlock() *core.Genesis {
	i, _ := new(big.Int).SetString("10000000000000000000000", 10)
	key1 := crypto.FromECDSAPub(&dkey1.PublicKey)

	var certList = [][]byte{pbft1Byte}
	coinbase := daddr1

	return &core.Genesis{
		Config:     params.DevnetChainConfig,
		Nonce:      928,
		ExtraData:  nil,
		GasLimit:   88080384,
		Difficulty: big.NewInt(20000),
		Alloc: map[common.Address]types.GenesisAccount{
			mAccount: {Balance: i},
		},
		Committee: []*types.CommitteeMember{
			&types.CommitteeMember{Coinbase: coinbase, Publickey: key1},
		},
		CertList: certList,
	}
}

func TestMath(t *testing.T) {
	contractBackend := backends.NewSimulatedBackend(gspec, 10000000)
	transactOpts := bind.NewKeyedTransactor(priKey, p2p1Byte, gspec.Config.ChainID)

	// Deploy the ENS registry
	ensAddr, _, _, err := DeployToken(transactOpts, contractBackend)
	if err != nil {
		t.Fatalf("can't DeployContract: %v", err)
	}
	ens, err := NewToken(ensAddr, contractBackend)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111111111111")
	_, err = ens.Add(transactOpts, big.NewInt(50000))
	if err != nil {
		log.Error("Failed to request token transfer", ": %v", err)
	}
	fmt.Println("2222222222222222222222222222222222222222222222222222222222222222222222222222222")

	//fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())
	contractBackend.Commit()
}

func TestMethod(t *testing.T) {
	method := []byte("add(uint256)")
	sig := crypto.Keccak256(method)[:4]
	fmt.Println(" ", hex.EncodeToString(sig))
	d := sha3.NewLegacyKeccak256()
	d.Write(method)
	fmt.Println(" ", hex.EncodeToString(d.Sum(nil)[:4]))
}
