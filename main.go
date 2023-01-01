/*
	    This project is to development of Blockchain core(bitcoin)

		Author: sectwo@gmail.com
		Date: 26 Dec, 2022
*/
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// Setting the global constant value "targetBits" for difficulty control
const targetBits = 16
const maxNonce = math.MaxInt64
const subsidy = 50
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const address = "user_address"

func main() {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}

/*
0. Set Block informaintion with SHA-256 function for hash value calculation
*/
// func (b *ST_Block) SetHash() {
// 	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
// 	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
// 	hash := sha256.Sum256(headers)
// 	b.Hash = hash[:]
// }

/*
1. Function for Create New block
ver0.1 : We can remove the "SetHash" function and add the "PoW" function in NewBlock()
*/
func NewBlock(transactions []*ST_Transaction, prevBlockHash []byte) *ST_Block {
	block := &ST_Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.NB_Run()

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
// func (bc *ST_Blockchain) AddBlock(data string) {
// 	prevBlock := bc.blocks[len(bc.blocks)-1]
// 	newBlock := NewBlock(data, prevBlock.Hash)
// 	bc.blocks = append(bc.blocks, newBlock)
// }

func (bc *ST_Blockchain) AddBlock(transactions []*ST_Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocksBucket"))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Println("err msg : ", err.Error())
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocksBucket"))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			fmt.Println("err msg : ", err.Error())
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		return nil
	})

}

func NewGenesisBlock(coinbase *ST_Transaction) *ST_Block {
	//return NewBlock("Genesis Block", []byte{})
	return NewBlock([]*ST_Transaction{coinbase}, []byte{})
}

func NewBlockchain(address string) *ST_Blockchain {
	var tip []byte
	db, err := bolt.Open("sample.db", 0600, nil)
	if err != nil {
		fmt.Println("error open DB Message : ", err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocksBucket"))
		if b == nil {
			cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)
			b, err := tx.CreateBucket([]byte("blocksBucket"))
			if err != nil {
				fmt.Println("error msg : ", err.Error())
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	bc := ST_Blockchain{tip, db}

	return &bc

	//return &ST_Blockchain{[]*ST_Block{NewGenesisBlock()}}
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
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (b *ST_Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func (pow *ProofOfWork) NB_Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.HashTransactions())

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

func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	//addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("error msg : ", err.Error())
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("error msg : ", err.Error())
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	// if addBlockCmd.Parsed() {
	// 	if *addBlockData == "" {
	// 		addBlockCmd.Usage()
	// 		os.Exit(1)
	// 	}
	// 	cli.addBlock(*addBlockData)
	// }

	// if addBlockCmd.Parsed() {
	// 	if *addBlockData == "" {
	// 		addBlockCmd.Usage()
	// 		os.Exit(1)
	// 	}
	// 	cli.addBlock(*addBlockData)
	// }

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

//=========================================================================================================
/*
	Adding Database
*/

func (b *ST_Block) Serialize() []byte {

	result, err := json.Marshal(b)
	if err != nil {
		fmt.Println("error : ", err.Error())
	}

	return result
}

func DeserializeBlock(d []byte) *ST_Block {

	var block ST_Block

	json.Unmarshal(d, &block)

	return &block
}

func (bc *ST_Blockchain) Iterator() *ST_BlockchainIterator {
	bci := &ST_BlockchainIterator{bc.tip, bc.db}
	return bci
}

func (i *ST_BlockchainIterator) Next() *ST_Block {
	var block *ST_Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocksBucket"))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		fmt.Println("error msg : ", err.Error())
	}

	i.currentHash = block.PrevBlockHash
	return block
}

//=============================================================================
/*
for CLI
*/

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

// func (cli *CLI) addBlock(data string) {
// 	cli.bc.AddBlock(data)
// 	fmt.Println("Success!")
// }

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.HashTransactions())
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

//=============================================================================
/*
Coin baseed Transactions (for Genesis Block)
*/
// func NewTransaction(vin []ST_TXInput, vout []ST_TXOutput) *ST_Transaction {
// 	tx := ST_Transaction{nil, vin, vout}
// 	tx.SetID()

// 	return &tx
// }

func (tx *ST_Transaction) SetID() {
	buf := new(bytes.Buffer)

	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buf.Bytes())
	tx.ID = hash[:]
}

func NewCoinbaseTX(to, data string) *ST_Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := ST_TXInput{[]byte{}, -1, data}
	txout := ST_TXOutput{subsidy, to}
	tx := ST_Transaction{nil, []ST_TXInput{txin}, []ST_TXOutput{txout}}
	tx.SetID()
	return &tx
}
