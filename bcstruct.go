/*
Block Structure Design for Constructing Blockchain
#1. SBlock = Blockhcain transection informations using only essential elements
*/
package main

import "math/big"

type ST_Block struct {
	Timestamp     int64  // Time now
	Data          []byte // Actual information on the block
	PrevBlockHash []byte // Hash value of previous block
	Hash          []byte // Hash value
}

/*
	Use arrays and maps to implement this structure
	Arrays : Maintain aligned hashes
	Maps : Maintain Hash-Block Pairs
*/
type ST_Blockchain struct {
	blocks []*ST_Block
}

/*
	for PoW
*/
type ProofOfWork struct {
	block  *ST_Block
	target *big.Int
}
