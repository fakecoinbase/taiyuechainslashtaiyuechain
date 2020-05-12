package crypto

import (
	"encoding/hex"
	"fmt"
	"github.com/taiyuechain/taiyuechain/common"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
)

func TestDecrypt(t *testing.T) {
	//1 is guoji 2 is guomi
	CryptoType = CRYPTO_SM2_SM3_SM4
	ecdsapri, _ := GenerateKey()
	fmt.Println(ecdsapri)
	ecdsabyte := FromECDSA(ecdsapri)
	ecdsapri, _ = ToECDSA(ecdsabyte)
	fmt.Println(ecdsapri)
	h := sha3.NewLegacyKeccak256()
	//h:=sm3.New()
	hash := h.Sum(nil)
	//sign and verify test
	sign, _ := Sign(hash, ecdsapri)
	pubbyte := FromECDSAPub(&ecdsapri.PublicKey)
	ecdpub, _ := UnmarshalPubkey(pubbyte)
	fmt.Println(ecdpub)
	fmt.Println(ecdsapri.PublicKey)

	boolverify := VerifySignature(pubbyte, hash, sign)
	fmt.Println(boolverify)
	//	compress and uncompress test
	compreebyte := CompressPubkey(&ecdsapri.PublicKey)
	fmt.Println(compreebyte)
	ecdsapub, _ := DecompressPubkey(compreebyte)
	fmt.Println(ecdsapub)
	//	sigtopub
	pubkey, err := SigToPub(hash, sign)
	if err != nil {
		panic(err)
	}
	//     Encryt and Decrypt test
	src := "caoliang"
	data := []byte(src)
	//ct,_:=Encrypt(CryptoType,ecdsapub,data,nil,nil)
	ct, _ := Encrypt(pubkey, data, nil, nil)
	fmt.Println(ct)
	m, _ := Decrypt(ecdsapri, ct, nil, nil)
	fmt.Println(string(m))

}

func Test_zeroBytes(t *testing.T) {
	CryptoType = CRYPTO_SM2_SM3_SM4
	ecdsapri, _ := GenerateKey()
	pubkeybyte := FromECDSAPub(&ecdsapri.PublicKey)
	stringsm2pub := hex.EncodeToString(pubkeybyte)
	fmt.Println(stringsm2pub)
	CryptoType = CRYPTO_P256_SH3_AES
	ecdpub, _ := UnmarshalPubkey(pubkeybyte)
	fmt.Println(ecdpub)
	byte, _ := hex.DecodeString(stringsm2pub)
	ecdpub1, _ := UnmarshalPubkey(byte)
	fmt.Println(ecdpub1)

}
func TestSm2(t *testing.T) {
	CryptoType = CRYPTO_P256_SH3_AES
	priv, err := GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", priv.Curve.IsOnCurve(priv.X, priv.Y))
	pub := &priv.PublicKey
	msg := []byte("123456")
	d0, err := Encrypt(pub, msg, nil, nil)
	if err != nil {
		fmt.Printf("Error: failed to encrypt %s: %v\n", msg, err)
		return
	}
	fmt.Printf("Cipher text = %v\n", d0)
	d1, err := Decrypt(priv, d0, nil, nil)
	if err != nil {
		fmt.Printf("Error: failed to decrypt: %v\n", err)
	}
	fmt.Printf("clear text = %s\n", d1)

	msg, _ = ioutil.ReadFile("ifile")
	//Keccak256(msg)
	sign, err := Sign(Keccak256(msg), priv)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("ofile", sign, os.FileMode(0644))
	if err != nil {
		log.Fatal(err)
	}
	signdata, _ := ioutil.ReadFile("ofile")
	ok := VerifySignature(FromECDSAPub(pub), Keccak256(msg), signdata)
	if ok != true {
		fmt.Printf("Verify error\n")
	} else {
		fmt.Printf("Verify ok\n")
	}

}

func TestString(t *testing.T) {
	CryptoType = CRYPTO_SM2_SM3_SM4
	priv, err := GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(FromECDSA(priv)), " pub ", hex.EncodeToString(FromECDSAPub(&priv.PublicKey)))
	CryptoType = CRYPTO_P256_SH3_AES
	priv, err = GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(FromECDSA(priv)), " pub ", hex.EncodeToString(FromECDSAPub(&priv.PublicKey)))
}
func Testrlp(t *testing.T) {

}

const BloomByteLength = 256

type Header struct {
	ParentHash    common.Hash    `json:"parentHash"       gencodec:"required"`
	Root          common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash        common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash   common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	CommitteeHash common.Hash    `json:"committeeRoot"    gencodec:"required"`
	Proposer      common.Address `json:"maker"            gencodec:"required"`
	Bloom         Bloom          `json:"logsBloom"        gencodec:"required"`
	SnailHash     common.Hash    `json:"snailHash"        gencodec:"required"`
	SnailNumber   *big.Int       `json:"snailNumber"      gencodec:"required"`
	Number        *big.Int       `json:"number"           gencodec:"required"`
	GasLimit      uint64         `json:"gasLimit"         gencodec:"required"`
	GasUsed       uint64         `json:"gasUsed"          gencodec:"required"`
	Time          *big.Int       `json:"timestamp"        gencodec:"required"`
	Extra         []byte         `json:"extraData"        gencodec:"required"`
}
type Bloom [BloomByteLength]byte

func TestHash256(t *testing.T) {
	//var he interface{}
	var header Header
	h := RlpHash(header)
	fmt.Println(h)
	doublebyte := Double256(h.Bytes())
	fmt.Println(doublebyte)
	hexstring := Hex(doublebyte)
	fmt.Println(hexstring)
	byte := Hash256Byte(doublebyte, doublebyte)
	fmt.Println(byte)
	hash256 := Hash256(doublebyte, doublebyte, doublebyte)
	fmt.Println(hash256)
}