package main

import (
	"math/big"

	"github.com/boltdb/bolt"
)

// 8) 트랜잭션 기능으로 변경점
//   - Data 필드 대신 Transactions 필드로 대체
type Block struct {
	PrevBlockHash []byte
	Hash          []byte
	Timestamp     int64
	//Data          []byte
	Transactions []*Transaction
	Nonce        int64
}

// 블록체인은 다수의 블록을 가짐 - 블록체인은 블록의 연결
// Block을 가지기지만 블록의 직접적 정보가 아닌 db의 정보와 lastHash 값만을 가짐
type Blockchain struct {
	//blocks []*Block
	db *bolt.DB
	l  []byte
}

// 작업증명(PoW) - 채굴을 위한 작업으로 난이도(Target) 설정
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// 영속성 추가시 블록체인 내부 순회를 위한 구조체
type blockchainIterator struct {
	db   *bolt.DB
	hash []byte
}
