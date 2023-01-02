package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
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
func NewCoinbaseTX(data, to string) *Transaction {
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}

	return NewTransaction([]TXInput{txin}, []TXOutput{txout})
}
