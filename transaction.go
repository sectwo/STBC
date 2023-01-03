package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
)

const (
	subsidy = 10 // BTC
)

// 새로운 트랜잭션 생성을 위한 함수
// 트랜잭션의 ID(해시값)의 경우 별도로 생성
func NewTransaction(vin []TXInput, vout []TXOutput) *Transaction {
	tx := Transaction{nil, vin, vout}
	tx.SetID()

	return &tx
}

// 트랜잭션의 ID 생성을 위한 함수
// 트랜잭션을 직렬화하고 해시화 함
func (tx *Transaction) SetID() {
	result, err := json.Marshal(tx)
	if err != nil {
		fmt.Println("error msg : ", err.Error())
		log.Panic(err)
	}

	hash := sha256.Sum256(result)
	tx.ID = hash[:]
}

// 트랜잭션의 ID 를 묶어서 해싱하기 위한 메서드
// 작업증명을 위해 사용되며, 작업증명을 위한 데이터를 준비할때 Block.Data를 사용하였지만 트랜잭션 기능이 추가되며 Block.Transactions로 변경
func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// 블록을 채굴하면 채굴자에게 보상을 주기위한 제일 첫 번째 트랜잭션을 위한 함수
// 입력이 없고 채굴자에게 보상을 지급하기 위한 출력값만 존재
// 10. 주소를 이용한 거래기능으로 인한 변경점
//   - 코인베이스 트랜잭션을 생성할 때 주소를 받아옴
//   - TXOutput 을 생성할 때 Base58CheckDecode 로 처리하고 공개키 해시를 출력에 넣어함
//   - txout = TXOutput -> NewTXOutput()으로 변경
func NewCoinbaseTX(data, to string) *Transaction {
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)

	return NewTransaction([]TXInput{txin}, []TXOutput{*txout})
}

// 새로운 TXOutput 을 생성을 위한 함수
// .Lock() 메서드는 주소에 해당하는 공개키 해시로 출력을 잠그기위해 사용
func NewTXOutput(value uint64, address string) *TXOutput {
	txo := &TXOutput{value, []byte(address)}
	txo.Lock(address)

	return txo
}

// 주소로 부터 공개키 해시(Public Key Hash)를 얻어온 다음 출력을 잠그기 위한 메서드
// 잠긴 것을 해제하여 소비할 수 있는 것은 지불을 받는 당사자 밖에 없음
func (out *TXOutput) Lock(address string) {
	pubKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	out.PubKeyHash = pubKeyHash
}
