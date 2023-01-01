/*
Block Structure Design for Constructing Blockchain
#1. SBlock = Blockhcain transection informations using only essential elements
*/
package main

import (
	"math/big"

	"github.com/boltdb/bolt"
)

type ST_Block struct {
	Timestamp     int64             // Time now
	Transactions  []*ST_Transaction // Transactions info
	PrevBlockHash []byte            // Hash value of previous block
	Hash          []byte            // Hash value
	Nonce         int
}

/*
Use arrays and maps to implement this structure
Arrays : Maintain aligned hashes
Maps : Maintain Hash-Block Pairs
*/
type ST_Blockchain struct {
	tip []byte
	db  *bolt.DB
}

/*
for PoW
*/
type ProofOfWork struct {
	block  *ST_Block
	target *big.Int
}

/*
Blockchain Itorator for Bucket Key in BoltDB
*/
type ST_BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//==================================================================================
/*
	CLI Structuer
*/
type CLI struct {
	bc *ST_Blockchain
}

//==================================================================================
/*
	Transaction Structure
*/
type ST_Transaction struct {
	ID   []byte
	Vin  []ST_TXInput
	Vout []ST_TXOutput
}

type ST_TXOutput struct {
	Value        int    // Satoshis((0.00000001 BTC)
	ScriptPubKey string //
}

type ST_TXInput struct {
	Txid      []byte // Transaction ID
	Vout      int    // Output index for transactions
	ScriptSig string // Scripts that provide data used in "ScriptPubKey"
}
