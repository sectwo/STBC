/*
    This project is to development of Blockchain core(bitcoin)

	Author: sectwo@gmail.com
	Date: 26 Dec, 2022

*/
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

// Setting the global constant value "targetBits" for difficulty control
const targetBits = 12
const maxNonce = math.MaxInt64

func main() {

	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Timestamp : %d\n", block.Timestamp)
		fmt.Println()
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}

}

/*
	0. Set Block informaintion with SHA-256 function for hash value calculation
*/
func (b *ST_Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

/*
	1. Function for Create New block
	ver0.1 : We can remove the "SetHash" function and add the "PoW" function in NewBlock()
*/
func NewBlock(data string, prevBlockHash []byte) *ST_Block {
	block := &ST_Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

/*
	2. Implementation of blockchain functionality
	#1. addBlock : Block Add-in
	#2. NewGenesisBlock : Genesis block generation capability for adding new blocks
	#3. NewBlockchain : Start New Blockchain
*/
func (bc *ST_Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

func NewGenesisBlock() *ST_Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *ST_Blockchain {
	return &ST_Blockchain{[]*ST_Block{NewGenesisBlock()}}
}

//=========================================================================================================
/*
	Adding PoW Function with PoW Validate
*/
func NewProofOfWork(b *ST_Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

func IntToHex(int_value int64) []byte {
	hex_value := strconv.FormatInt(int64(int_value), 16)
	return []byte(hex_value)
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
