package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

func NewWallet() *Wallet {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return &Wallet{privKey, pubKey}
}

// 지갑의 주소 생성을 위한 메서드
// 주소는 개인키로부터 도출되며, 비트코인 주소의 경우 주소의 접두사로 1 이 붙음
// 공개키를 더블 해싱(Double-Hashing)하여 SHA256, RIPEMD160 를 각각 한 번씩 해주고, 비트코인 주소를 의미하는 버전 접두어 0x00 을 붙인 다음, 마지막으로 Base58CheckEncode를 하여 주소 생성
func (w *Wallet) GetAddress() string {
	publicRIPEMD160 := HashPubKey(w.PubKey)
	version := byte(0x00)

	return base58.CheckEncode(publicRIPEMD160, version)
}

// 공개키를 더블 해싱 하기 위한 함수
// SHA256과 RIPEMD160로 해성 처리 후 반환
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}

	return RIPEMD160Hasher.Sum(nil)
}

// wallet.dat 파일을 만들기 위한 함수
// 함수의 이름이 소문자로 시작하기때문에 외부에서 접근하지 않는 것을 전재로 함
func createKeyStore() error {
	file, err := os.OpenFile(walletFile, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return file.Close()
}

// wallet.dat 파일을 읽어와 새로운 KeyStore 를 반환하는 함수
// 만약 wallet.dat 파일이 없다면 생성하고 비어있는 KeyStore 를 반환
func NewKeyStore() *KeyStore {
	keyStore := KeyStore{make(map[string]*Wallet)}

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		err := createKeyStore()
		if err != nil {
			log.Panic(err)
		}
	} else {
		fileContent, err := ioutil.ReadFile(walletFile)
		if err != nil {
			log.Panic(err)
		}

		gob.Register(elliptic.P256())

		decoder := gob.NewDecoder(bytes.NewReader(fileContent))
		err = decoder.Decode(&keyStore)
		if err != nil {
			log.Panic(err)
		}
	}

	return &keyStore
}

// .Wallets 와 wallet.dat 파일을 내용을 동기화시키기위한 함수
// KeyStore 자체를 인코딩하여 저장함
// 2023.01.02_sectwo : 저장시 오류 발생 수정 필요(.dat 파일에 재대로 저장되지 않는 오류) 해결을 위해 json 파일로 변경 시도 예정 ver0.6에서
func (ks *KeyStore) Save() {

	result, err := json.Marshal(elliptic.P256())
	if err != nil {
		fmt.Println("error : ", err.Error())
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, result, 0644)
	if err != nil {
		log.Panic(err)
	}
}

// func (ks *KeyStore) Save() {
// 	buf := new(bytes.Buffer)

// 	gob.Register(elliptic.P256())

// 	encoder := gob.NewEncoder(buf)
// 	err := encoder.Encode(ks)
// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	err = ioutil.WriteFile(walletFile, buf.Bytes(), 0644)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// }

// 지갑을 만들기 위한 메서드
// 지갑을 만들고 키스토어에 저장
func (ks *KeyStore) CreateWallet() *Wallet {
	wallet := NewWallet()

	ks.Wallets[wallet.GetAddress()] = wallet
	ks.Save()

	return wallet
}
