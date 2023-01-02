/*
	This project is to development of Blockchain core(bitcoin)

	순서 :
		1) 새로운 블록 생성(MewBlock)
		2) 블록의 해시 생성(SetHash)
		3) 새로운 블록체인의 생성(NewBlockchain)
		4) 블록체인에 블록추가
		5) 작업증명 추가
		6) 영속성 부여(BlotDB 사용) - 기존 Blockchain이 가진 Block을 db로 변경
		7) 테스트를 위한 CLI 추가
		8) Transaction 기능 추가

	Author: sectwo@gmail.com
	Date: 26 Dec, 2022
*/
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const (
	BlocksBucket = "blocks"
	dbFile       = "chain.db"
	targetBits   = 16
)

func main() {
	cli := CLI{}
	cli.Run()
}

// 8) 트랜잭션 기능으로 인한 변경점
// 		- 기존 입력파라메타의 data를 trasaction으로 변경
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{prevBlockHash, []byte{}, time.Now().Unix(), transactions, 0}
	pow := NewProofOfWork(block)
	block.Nonce, block.Hash = pow.Run()

	return block
}

func (b *Block) SetHash() {
	header := bytes.Join([][]byte{
		b.PrevBlockHash,
		b.HashTransaction(),
		IntToHex(b.Timestamp),
	}, []byte{})

	hash := sha256.Sum256(header)
	b.Hash = hash[:]
}

func IntToHex(int_value int64) []byte {
	hex_value := strconv.FormatInt(int64(int_value), 16)
	return []byte(hex_value)
}

// 새로운 블록체인 생성 - 제네시스 블록 생성으로 시작
// 6) 영속성으로 인한 변경점
//		- 기존 블록 정보가 아닌 db의 정보와 LastHash값을 가져야함
//		- 이를 위해 버킷(RDB에서의 테이블)에 블록을 담고 조회 할 수 있어야함
// 7) Cli 추가로 인한 변경점
// 		- 기존 이미 블록체인이 존재하는 경우에 대한 Genesis Block 생성은 사라지고 기존의 블록체인이 존재하는 경우, 기존 블록체인을 얻어오기 위해 사용됨
func NewBlockchain() *Blockchain {

	blockchain := new(Blockchain)
	var l []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Println("error msg : ", err.Error())
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		// 이미 블록체인이 존재하는 경우
		l = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Println("error msg : ", err.Error())
		log.Panic(err)
	}
	blockchain.db = db
	blockchain.l = l

	return blockchain
}

// 새로운 블록체인에 블록연결([제네시스블록]-[새롭게 생성되는 블록]-[...])
// 6) 영속성으로 인한 변경점
// 		- .blocks에 저장하던 것을 boltDB에 저장할 수 있도록 변경
//		- 마지막 블록해시 l
// 8) 트랜잭션 기능으로 인한 변경점
// 		- NewBlock() 변경으로 입력 파라메타와 기능 수정
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	block := NewBlock(transactions, bc.l)
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		err := b.Put(block.Hash, block.Serialize())
		if err != nil {
			fmt.Println("error : ", err.Error())
			log.Panic(err)
		}

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			fmt.Println("error : ", err.Error())
			log.Panic(err)
		}
		bc.l = block.Hash

		return nil
	})
	if err != nil {
		fmt.Println("error : ", err.Error())
		log.Panic(err)
	}
}

//================================================================================
// 5) 작업증명
// - 작업증명은 채굴이라 말할 수 있으며, 끝 자리(0x000000~~~)과 같은 비트수에 맞는 해시값을 찾는 작업
// - 해시값을 비교해가며 찾아야하며, targetbits를 통해 난이도 설정 가능
// - 난이도는 16진수를 나타내며 24의 경우 24bit 즉, 끝자리 0이 6개를 의미함

// target 지정을 우선하며 Shift 연산자를 사용하여 target을 지정함
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-targetBits)

	pow := &ProofOfWork{b, target}
	return pow
}

// target과 prepareData를 통해 준비한 데이터를 해시화 한 값과 비교하여 더 작으면 완료시킴
// 작업 증며을위한 실질적인 메서드
// 이때 nonce는 반복을 위한 단순한 counter용도로 사용
// 8) 트랜잭션 기능으로 인한 변경점
// 		- 작업증명을 위한 준비데이터를 data에서 Block.HashTransaction()을 사용하여 트랜잭션을 해싱
func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransaction(),
		IntToHex(pow.block.Timestamp),
		IntToHex(nonce),
		IntToHex(targetBits),
	}, []byte{})

	return data
}

// 작업 증명을위한 실질적인 메서드
// target과 prepareData를 통해 준비한 데이터를 해시화 한 값과 비교하여 더 작으면 완료시킴
// 	- target보다 더 작은값을 찾기 위해 nonce를 증가 시키며 반복
func (pow *ProofOfWork) Run() (int64, []byte) {
	var nonce int64

	var hashInt big.Int
	var hash [32]byte

	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

// 작업증명(PoW)를 통해 나온 것인지를 증명하기위한 메서드
// 블록에 존재하는 Nonce값을 사용하여 한번의 사이클로 증명 가능
func (pow *ProofOfWork) Validate(b *Block) bool {
	var hashInt big.Int
	data := pow.prepareData(b.Nonce)
	hash := sha256.Sum256(data)

	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

//================================================================================
// 6) 영속성 추가

// BoltDB로 데이터를 전송하기위해 블록정보를 직렬화 하기 위한 메서드
func (b *Block) Serialize() []byte {

	result, err := json.Marshal(b)
	if err != nil {
		fmt.Println("error : ", err.Error())
		log.Panic(err)
	}

	return result
}

// BoltDB에서 조회시 BlotDB의 데이터를 전송하기위해 블록정보를 역직렬화 하기 위한 메서드
func DeserializeBlock(d []byte) *Block {

	var block Block

	json.Unmarshal(d, &block)

	return &block
}

// 블록 조회를 위한 블록체인 내부 순회 반복자 함수
func NewBlockchainIterator(bc *Blockchain) *blockchainIterator {
	return &blockchainIterator{bc.db, bc.l}
}

// BoltDB를 조회하여 버킷(블록)을 반환하며 가장 마지막 블록-> 최초의 블록 순서로 조회
func (i *blockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		encodedBlock := b.Get(i.hash)
		block = DeserializeBlock(encodedBlock)

		i.hash = block.PrevBlockHash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}

// 다음 블록이 존재하는지 검사하기 위한 메서드
// 반복자의 hash가 다음블록과 같은지 같지 않은지 비교 (작을땐 -1, 클땐 1, 같을땐 0)
func (i *blockchainIterator) HasNext() bool {
	result := bytes.Compare(i.hash, []byte{}) != 0
	fmt.Println(result)
	return result
}

//================================================================================
// 7) Cli 기능 추가

// 블록체인을 새로 생성(제네시스 블록 생성)
// 8) 트랜잭션 기능 추가로 인한 변경점
//		- 입력 파라메타 "address string" 추가 : 블록체인을 생성하고 제네시스 블록을 채굴한 사람에게 보상을 지급을 위함
func CreateBlockchain(address string) *Blockchain {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var l []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(BlocksBucket))
		if err != nil {
			log.Panic(err)
		}

		//genesis := NewBlock("Genesis Block", []byte{})
		genesis := NewBlock([]*Transaction{NewCoinbaseTX("", address)}, []byte{})

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// "l" 키는 마지막 블록해시를 저장합니다.
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		l = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{db, l}
}

// BlockchainIterator 를 사용하여 블록체인을 순회
func (bc *Blockchain) List() {
	bci := NewBlockchainIterator(bc)

	for bci.HasNext() {
		block := bci.Next()

		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Transactions)

		pow := NewProofOfWork(block)
		fmt.Println("pow:", pow.Validate(block))

		fmt.Println()
	}
}
