package test

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"runtime/debug"
	"testing"

	"github.com/taiyuechain/taiyuechain/cim"

	"strings"

	"github.com/taiyuechain/taiyuechain/accounts/abi"
	"github.com/taiyuechain/taiyuechain/common"
	"github.com/taiyuechain/taiyuechain/consensus/minerva"
	"github.com/taiyuechain/taiyuechain/core"
	"github.com/taiyuechain/taiyuechain/core/state"
	"github.com/taiyuechain/taiyuechain/core/types"
	"github.com/taiyuechain/taiyuechain/core/vm"
	"github.com/taiyuechain/taiyuechain/crypto"
	taicert "github.com/taiyuechain/taiyuechain/cert"
	"github.com/taiyuechain/taiyuechain/log"
	"github.com/taiyuechain/taiyuechain/params"
	"github.com/taiyuechain/taiyuechain/yuedb"
)

var (
	pbft1Name = "pbft1priv"
	pbft2Name = "pbft2priv"
	pbft3Name = "pbft3priv"
	pbft4Name = "pbft4priv"
	pbft1path = "../testcert/" + pbft1Name + ".pem"
	pbft2path = "../testcert/" + pbft2Name + ".pem"
	pbft3path = "../testcert/" + pbft3Name + ".pem"
	pbft4path = "../testcert/" + pbft4Name + ".pem"

	db       = yuedb.NewMemDatabase()
	gspec    = DefaulGenesisBlock()
	abiCA, _ = abi.JSON(strings.NewReader(vm.PermissionABIJSON))
	signer   = types.NewSigner(gspec.Config.ChainID)

	// pbft 1
	priKey, _ = crypto.HexToECDSA("7631a11e9d28563cdbcf96d581e4b9a19e53ad433a53c25a9f18c74ddf492f75")
	// pbft 2
	prikey2, _ = crypto.HexToECDSA("bab8dbdcb4d974eba380ff8b2e459efdb6f8240e5362e40378de3f9f5f1e67bb")
	// pbft 3
	prikey3, _ = crypto.HexToECDSA("122d186b77a030e04f5654e13d934b21af2aac03b942c3ecda4632364d81cbab")
	prikey4, _ = crypto.HexToECDSA("fe44cbc0e164092a6746bd57957422ab165c009d0299c7639a2f4d290317f20f")
	prikey5, _ = crypto.HexToECDSA("77b4e6383502fd145cae5c2f8db28a9b750394bd70c0c138b915bb1327225489")

	saddr1 = crypto.PubkeyToAddress(priKey.PublicKey)
	saddr2 = crypto.PubkeyToAddress(prikey2.PublicKey)
	saddr3 = crypto.PubkeyToAddress(prikey3.PublicKey)
	saddr4 = crypto.PubkeyToAddress(prikey4.PublicKey)
	saddr5 = crypto.PubkeyToAddress(prikey5.PublicKey)

	pbft1Byte, _ = taicert.ReadPemFileByPath(pbft1path)
	pbft2Byte, _ = taicert.ReadPemFileByPath(pbft2path)
	pbft3Byte, _ = taicert.ReadPemFileByPath(pbft3path)
	pbft4Byte, _ = taicert.ReadPemFileByPath(pbft4path)

	pkey2, _ = crypto.HexToECDSA("ea4297749d514cc476fe971a7fe20100cbd29f010864341b3e624e8744d46cec")
	paddr2   = crypto.PubkeyToAddress(pkey2.PublicKey)

	pkey3, _ = crypto.HexToECDSA("86937006ac1e6e2c846e160d93f86c0d63b0fcefc39a46e9eaeb65188909fbdc")
	paddr3   = crypto.PubkeyToAddress(pkey3.PublicKey)

	pkey4, _ = crypto.HexToECDSA("cbddcbecd252a8586a4fd759babb0cc77f119d55f38bc7f80a708e75964dd801")
	paddr4   = crypto.PubkeyToAddress(pkey4.PublicKey)
)

func DefaulGenesisBlock() *core.Genesis {
	i, _ := new(big.Int).SetString("10000000000000000000000", 10)
	key1 := crypto.FromECDSAPub(&priKey.PublicKey)
	key2 := crypto.FromECDSAPub(&prikey2.PublicKey)
	key3 := crypto.FromECDSAPub(&prikey3.PublicKey)
	key4 := crypto.FromECDSAPub(&prikey4.PublicKey)

	var certList = [][]byte{pbft1Byte, pbft2Byte, pbft3Byte, pbft4Byte}

	return &core.Genesis{
		Config:       params.DevnetChainConfig,
		GasLimit:     20971520,
		UseGas:       1,
		IsCoin:   1,
		KindOfCrypto: 2,
		Timestamp:    1537891200,
		Alloc: map[common.Address]types.GenesisAccount{
			saddr1: {Balance: i},
		},
		Committee: []*types.CommitteeMember{
			&types.CommitteeMember{Coinbase: saddr1, Publickey: key1},
			&types.CommitteeMember{Coinbase: saddr2, Publickey: key2},
			&types.CommitteeMember{Coinbase: saddr3, Publickey: key3},
			&types.CommitteeMember{Coinbase: saddr4, Publickey: key4},
		},
		CertList: certList,
	}
}

func newTestPOSManager(sBlocks int, executableTx func(uint64, *core.BlockGen, *core.BlockChain, *types.Header, *state.StateDB, *cim.CimList)) {

	//new cimList
	cimList := cim.NewCIMList(uint8(crypto.CryptoType))
	engine := minerva.NewFaker(cimList)

	params.MinTimeGap = big.NewInt(0)
	params.SnailRewardInterval = big.NewInt(3)
	params.EnablePermission = 1

	genesis := gspec.MustCommit(db)
	vm.SetPermConfig(true,true)
	blockchain, _ := core.NewBlockChain(db, nil, gspec.Config, engine, vm.Config{}, cimList)
	//init cert list to
	// need init cert list to statedb
	stateDB, err := blockchain.State()
	if err != nil {
		panic(err)
	}
	err = cimList.InitCertAndPermission(blockchain.CurrentBlock().Number(), stateDB)
	if err != nil {
		panic(err)
	}

	chain, _ := core.GenerateChain(gspec.Config, genesis, engine, db, sBlocks*60, func(i int, gen *core.BlockGen) {

		header := gen.GetHeader()
		stateDB := gen.GetStateDB()
		executableTx(header.Number.Uint64(), gen, blockchain, header, stateDB, cimList)
	})
	if _, err := blockchain.InsertChain(chain); err != nil {
		panic(err)
	}
}

//neo test
func sendGrantPermissionTranscation(height uint64, gen *core.BlockGen, from, to,group common.Address, permission *big.Int, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 25 {
		nonce, _ := getNonce(gen, from, state, "grantPermission", txPool)
		input := packInput(abiStaking, "grantPermission", "grantPermission", common.Address{}, to, group, permission, true)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendGrantContractPermissionTranscation(height uint64, gen *core.BlockGen, from, to,contract common.Address, permission *big.Int, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 25 {
		nonce, _ := getNonce(gen, from, state, "grantPermission", txPool)
		input := packInput(abiStaking, "grantPermission", "grantPermission", contract, to, common.Address{}, permission, true)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendRevokePermissionTranscation(height uint64, gen *core.BlockGen, from, to common.Address, permission *big.Int, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 40 {
		nonce, _ := getNonce(gen, from, state, "sendRevokePermissionTranscation", txPool)
		input := packInput(abiStaking, "revokePermission", "sendRevokePermissionTranscation", from, to, common.Address{}, permission, true)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendRevokeContractPermissionTranscation(height uint64, gen *core.BlockGen, from, to,contract common.Address, permission *big.Int, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 25 {
		nonce, _ := getNonce(gen, from, state, "sendRevokePermissionTranscation", txPool)
		input := packInput(abiStaking, "revokePermission", "sendRevokePermissionTranscation", contract, to, common.Address{}, permission, true)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendCreateGroupPermissionTranscation(height uint64, gen *core.BlockGen, from common.Address, gropName string, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 25 {
		nonce, _ := getNonce(gen, from, state, "sendCreateGroupPermissionTranscation", txPool)
		input := packInput(abiStaking, "createGroupPermission", "sendCreateGroupPermissionTranscation", gropName)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendDelGroupPermissionTranscation(height uint64, gen *core.BlockGen, from, GroupAddr common.Address, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 70 {
		nonce, _ := getNonce(gen, from, state, "sendDelGroupPermissionTranscation", txPool)
		input := packInput(abiStaking, "delGroupPermission", "sendDelGroupPermissionTranscation", GroupAddr)
		addTx(gen, blockchain, nonce, nil, input, txPool, priKey, signer)
	}
}

//neo test
func sendIsApproveCACertTranscation(height uint64, gen *core.BlockGen, from common.Address, cert []byte, priKey *ecdsa.PrivateKey, signer types.Signer, state *state.StateDB, blockchain *core.BlockChain, abiStaking abi.ABI, txPool txPool) {
	if height == 30 {
		input := packInput(abiStaking, "isApproveCaCert", "sendIsApproveCACertTranscation", cert)
		var args bool
		readTx(gen, blockchain, 0, big.NewInt(0), input, txPool, priKey, signer, "isApproveCaCert", &args)
		printTest("get Cert Amount is ", args)
	}
}

func addTx(gen *core.BlockGen, blockchain *core.BlockChain, nonce uint64, value *big.Int, input []byte, txPool txPool, priKey *ecdsa.PrivateKey, signer types.Signer) {
	//2426392 1000000000
	//866328  1000000
	//2400000
	tx, _ := types.SignTx(types.NewTransaction(nonce, types.PermiTableAddress, value, 2446392, big.NewInt(1000000000), input), signer, priKey)

	if gen != nil {
		gen.AddTxWithChain(blockchain, tx)
	} else {
		txPool.AddRemotes([]*types.Transaction{tx})
	}
}

func readTx(gen *core.BlockGen, blockchain *core.BlockChain, nonce uint64, value *big.Int, input []byte, txPool txPool, priKey *ecdsa.PrivateKey, signer types.Signer, abiMethod string, result interface{}) {
	tx, _ := types.SignTx(types.NewTransaction(nonce, types.PermiTableAddress, value, 866328, big.NewInt(1000000), input), signer, priKey)

	if gen != nil {
		output, gas := gen.ReadTxWithChain(blockchain, tx)
		err := abiCA.Unpack(result, abiMethod, output)
		if err != nil {
			printTest(abiMethod, " error ", err)
		}
		printTest("readTx gas ", gas)
	} else {
		txPool.AddRemotes([]*types.Transaction{tx})
	}
}
func packInput(abiStaking abi.ABI, abiMethod, method string, params ...interface{}) []byte {
	input, err := abiStaking.Pack(abiMethod, params...)
	if err != nil {
		printTest(method, " error ", err)
	}
	return input
}

type txPool interface {
	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*types.Transaction) []error
	State() *state.ManagedState
}

func printTest(a ...interface{}) {
	log.Info("test", "SendTX", a)
}

func getNonce(gen *core.BlockGen, from common.Address, state1 *state.StateDB, method string, txPool txPool) (uint64, *state.StateDB) {
	var nonce uint64
	var stateDb *state.StateDB
	if gen != nil {
		nonce = gen.TxNonce(from)
		stateDb = gen.GetStateDB()
	} else {
		stateDb = state1
		nonce = txPool.State().GetNonce(from)
	}
	return nonce, stateDb
}

func sendTranction(height uint64, gen *core.BlockGen, state *state.StateDB, from, to common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, signer types.Signer, txPool txPool, header *types.Header, cimList *cim.CimList) {
	if height == 10 {
		nonce, statedb := getNonce(gen, from, state, "sendTranction", txPool)
		balance := statedb.GetBalance(to)
		printTest("sendTranction ", balance.Uint64(), " height ", height, " current ", header.Number.Uint64(), " from ", types.ToTai(state.GetBalance(from)))
		tx, _ := types.SignTx(types.NewTransaction(nonce, to, value, params.TxGas, new(big.Int).SetInt64(1000000), nil), signer, privateKey)
		if gen != nil {
			if check, err := cimList.VerifyPermission(tx, signer, *statedb); !check {
				fmt.Println(header.Number.Uint64(), " --------------------------------- ", err, " ---------------------------------------")
			} else {
				gen.AddTx(tx)
			}
		} else {
			txPool.AddRemotes([]*types.Transaction{tx})
		}
	}
}

func sendContractTranction(height uint64, gen *core.BlockGen, state *state.StateDB, from common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, signer types.Signer, txPool txPool, header *types.Header, cimList *cim.CimList) {
	if height == 10 {
		nonce, statedb := getNonce(gen, from, state, "sendTranction", txPool)
		printTest("sendTranction ", " height ", height, " current ", header.Number.Uint64(), " from ", types.ToTai(state.GetBalance(from)))

		tx, _ := types.SignTx(types.NewContractCreation(nonce, value, params.TxGasContractCreation, new(big.Int).SetInt64(1000000), nil), signer, privateKey)
		if gen != nil {
			if check, err := cimList.VerifyPermission(tx, signer, *statedb); !check {
				fmt.Println(header.Number.Uint64(), " --------------------------------- ", err, " ---------------------------------------")
			} else {
				gen.AddTx(tx)
			}
		} else {
			txPool.AddRemotes([]*types.Transaction{tx})
		}
	}
}

func loadPermissionTable(state *state.StateDB) *vm.PerminTable {
	ptable := vm.NewPerminTable()
	ptable.Load(state)
	return ptable
}

func checkBaseCrtContractPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkCreateContractTxPermission(from,t,has,ptable)

	checkAddContractPermission(from,t,false,ptable)
	checkDelContractPermission(from,t,false,ptable)
	checkAddCrtContractManagerPermission(from,t,false,ptable)
	checkDelCrtContractManagerPermission(from,t,false,ptable)
}

func checkBaseCrtManagerContractPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkCreateContractTxPermission(from,t,has,ptable)
	checkAddContractPermission(from,t,has,ptable)
	checkDelContractPermission(from,t,has,ptable)
	checkAddCrtContractManagerPermission(from,t,has,ptable)
	checkDelCrtContractManagerPermission(from,t,has,ptable)
}

func checkBaseContractPermission(from,contract common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkAccessContractPermission(from,contract,t,has,ptable)

	checkAddContractMemberPermission(from,contract,t,false,ptable)
	checkDelContractMemberPermission(from,contract,t,false,ptable)
	checkAddContractManagerPermission(from,contract,t,false,ptable)
	checkDelContractManagerPermission(from,contract,t,false,ptable)
}

func checkBaseManagerContractPermission(from,contract common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkAccessContractPermission(from,contract,t,has,ptable)

	checkAddContractMemberPermission(from,contract,t,has,ptable)
	checkDelContractMemberPermission(from,contract,t,has,ptable)
	checkAddContractManagerPermission(from,contract,t,has,ptable)
	checkDelContractManagerPermission(from,contract,t,has,ptable)
}

func checkCreateContractTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	//if has {
	//	checkSendTxPermission(from,t,true)
	//}

	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.PerminType_CreateContract) != has {
		printStack("CheckActionPerm err PerminType_CreateContract",t)
	}
}

func checkAddContractPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_AddCrtContractPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddCrtContractPerm",t)
	}
}

func checkDelContractPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_DelCrtContractPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelCrtContractPerm",t)
	}
}

func checkAddCrtContractManagerPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_AddCrtContractManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddCrtContractManagerPerm",t)
	}
}

func checkDelCrtContractManagerPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_DelCrtContractManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelCrtContractManagerPerm",t)
	}
}

func checkAddContractMemberPermission(from, contractAddr common.Address, t *testing.T, has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from, common.Address{}, contractAddr, vm.ModifyPerminType_AddContractMemberPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddContractMemberPerm",t)
	}
}

func checkDelContractMemberPermission(from, contractAddr common.Address, t *testing.T, has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from, common.Address{}, contractAddr, vm.ModifyPerminType_DelContractMemberPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelContractMemberPerm",t)
	}
}

func checkAddContractManagerPermission(from,contract common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},contract,vm.ModifyPerminType_AddContractManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddCrtContractManagerPerm",t)
	}
}

func checkDelContractManagerPermission(from,contract common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},contract,vm.ModifyPerminType_DelContractManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelCrtContractManagerPerm",t)
	}
}

func checkAccessContractPermission(from,contract common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	//if has {
	//	checkSendTxPermission(from,t,true)
	//}
	if ptable.CheckActionPerm(from,common.Address{},contract,vm.PerminType_AccessContract) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelCrtContractManagerPerm",t)
	}
}

func printStack(err string,t *testing.T) {
	debug.PrintStack()
	t.FailNow()
}

func printResError(res bool,err error,t *testing.T,str string) {
	if !res{
		fmt.Println(err)
		printStack(str,t)
	}
}

func checkBothTxGroupPermission(from,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkBaseManagerSendTxPermission(from,t,true,ptable)
	checkBaseGroupManagerPermission(from,gropAddr,t,true,ptable)
}

func checkNoBothTxGroupPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkNoBaseSendTxPermission(from,t,false,ptable)
	checkNoBaseGroupPermission(from,common.Address{},t,false,ptable)

}

func checkNoBaseSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkSendTxPermission(from,t,false,ptable)
	checkAddSendTxPermission(from,t,false,ptable)
	checkDelSendTxPermission(from,t,false,ptable)
	checkSendTxManagerPermission(from,t,false,ptable)
	checkDelSendTxManagerPermission(from,t,false,ptable)
}

func checkBaseSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkSendTxPermission(from,t,has,ptable)
	checkAddSendTxPermission(from,t,false,ptable)
	checkDelSendTxPermission(from,t,false,ptable)
	checkSendTxManagerPermission(from,t,false,ptable)
	checkDelSendTxManagerPermission(from,t,false,ptable)
}

func checkBaseManagerSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkSendTxPermission(from,t,has,ptable)
	checkAddSendTxPermission(from,t,has,ptable)
	checkDelSendTxPermission(from,t,has,ptable)
	checkSendTxManagerPermission(from,t,has,ptable)
	checkDelSendTxManagerPermission(from,t,has,ptable)
}

func checkSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.PerminType_SendTx) != has {
		printStack("CheckActionPerm err PerminType_SendTx",t)
	}
}

func checkAddSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_AddSendTxPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddSendTxManagerPerm",t)
	}
}

func checkDelSendTxPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_DelSendTxPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddSendTxManagerPerm",t)
	}
}

func checkSendTxManagerPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_AddSendTxManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddSendTxManagerPerm",t)
	}
}

func checkDelSendTxManagerPermission(from common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,common.Address{},common.Address{},vm.ModifyPerminType_DelSendTxManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelSendTxManagerPerm",t)
	}
}

func checkNoBaseGroupPermission(from, gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkGroupSendTxPermission(from,gropAddr,t,false,ptable)
	checkAddGroupMemberPermission(from,gropAddr,t,false,ptable)
	checkDelGroupMemberPermission(from,gropAddr,t,false,ptable)
	checkAddGroupManagerPermission(from,gropAddr,t,false,ptable)
	checkDelGroupManagerPermission(from,gropAddr,t,false,ptable)
	checkDelGropPermission(from,gropAddr,t,false,ptable)
}

func checkBaseGroupPermission(from, gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkGroupSendTxPermission(from,gropAddr,t,has,ptable)
	checkAddGroupMemberPermission(from,gropAddr,t,false,ptable)
	checkDelGroupMemberPermission(from,gropAddr,t,false,ptable)
	checkAddGroupManagerPermission(from,gropAddr,t,false,ptable)
	checkDelGroupManagerPermission(from,gropAddr,t,false,ptable)
	checkDelGropPermission(from,gropAddr,t,false,ptable)
}

func checkBaseGroupManagerPermission(from, gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	checkGroupSendTxPermission(from,gropAddr,t,true,ptable)
	checkAddGroupMemberPermission(from,gropAddr,t,true,ptable)
	checkDelGroupMemberPermission(from,gropAddr,t,true,ptable)
	checkAddGroupManagerPermission(from,gropAddr,t,true,ptable)
	checkDelGroupManagerPermission(from,gropAddr,t,true,ptable)
	checkDelGropPermission(from,gropAddr,t,true,ptable)
}

func checkGroupSendTxPermission(from,group common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(from,group,common.Address{},vm.PerminType_SendTx) != has {
		printStack("CheckActionPerm err PerminType_SendTx",t)
	}
}

func checkAddGroupMemberPermission(member,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(member,gropAddr,common.Address{},vm.ModifyPerminType_AddGropMemberPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddGropManagerPerm",t)
	}
}

func checkDelGroupMemberPermission(member,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(member,gropAddr,common.Address{},vm.ModifyPerminType_DelGropMemberPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddGropManagerPerm",t)
	}
}

func checkAddGroupManagerPermission(member,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(member,gropAddr,common.Address{},vm.ModifyPerminType_AddGropManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddGropManagerPerm",t)
	}
}

func checkDelGroupManagerPermission(member,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(member,gropAddr,common.Address{},vm.ModifyPerminType_DelGropManagerPerm) != has {
		printStack("CheckActionPerm err ModifyPerminType_AddGropManagerPerm",t)
	}
}

func checkDelGropPermission(member,gropAddr common.Address,t *testing.T,has bool,ptable *vm.PerminTable) {
	if ptable.CheckActionPerm(member,gropAddr,common.Address{},vm.ModifyPerminType_DelGrop) != has {
		printStack("CheckActionPerm err ModifyPerminType_DelGrop",t)
	}
}