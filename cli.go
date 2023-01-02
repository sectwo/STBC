package main

import (
	"flag"
	"fmt"
	"os"
)

func (c *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.db.Close()
}

// 새로운 블록을 추가하기 위한 메서드
// Blockchain.AddBlock()을 호출하며, 여기서 가져오는 블록체인은 기존에 있던 체인에 추가하는 것이므로 NewBlockchain() 사용
// func (c *CLI) addBlock(data string) {
// 	bc := NewBlockchain()
// 	defer bc.db.Close()

// 	bc.AddBlock(data)
// }

// 블록체인에 있는 데이터 출력을 위한 메서드
// 이미 있는 블록체인을 출력하는 것이니 NewBlockchain()을 사용
func (c *CLI) list() {
	bc := NewBlockchain()
	defer bc.db.Close()

	bc.List()
}

// 어플리케이션 사용을 위한 메서드
func (c *CLI) Run() {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	newWalletCmd := flag.NewFlagSet("newwallet", flag.ExitOnError)

	sendValue := sendCmd.Uint64("value", 0, "")
	sendFrom := sendCmd.String("from", "", "")
	sendTo := sendCmd.String("to", "", "")

	newAddress := newCmd.String("address", "", "")
	getBalanceAddress := getBalanceCmd.String("address", "", "")

	switch os.Args[1] {
	case "new":
		newCmd.Parse(os.Args[2:])
	case "send":
		sendCmd.Parse(os.Args[2:])
	case "getbalance":
		getBalanceCmd.Parse(os.Args[2:])
	case "newwallet":
		newWalletCmd.Parse(os.Args[2:])
	default:
		os.Exit(1)
	}

	if newCmd.Parsed() {
		if *newAddress == "" {
			newCmd.Usage()
			os.Exit(1)
		}
		c.createBlockchain(*newAddress)
	}
	if sendCmd.Parsed() {
		if *sendValue == 0 || *sendFrom == "" || *sendTo == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		c.send(*sendValue, *sendFrom, *sendTo)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		fmt.Printf("Balance of '%s': %d\n", *getBalanceAddress, c.getBalance(*getBalanceAddress))
	}
	if newWalletCmd.Parsed() {
		fmt.Printf("Address: %s", c.newWallet())
	}
}

// 거래를 위한 기능
// 블록 하나당 트랜잭션을 한 개만 가지며, 블록은 채굴되지만 Coinbase Transaction 을 지정해주지 않았기 때문에 보상은 주어지지 않음
func (c *CLI) send(value uint64, from, to string) {
	bc := NewBlockchain()
	defer bc.db.Close()

	tx := bc.Send(value, from, to)
	bc.AddBlock([]*Transaction{tx})
}

// 특정 주소의 자금을 보기 위한 기능
// 특정 주소의 UTXO 의 합을 보여줌
func (c *CLI) getBalance(address string) uint64 {
	bc := NewBlockchain()
	defer bc.db.Close()

	return bc.GetBalance(address)
}

// 지갑을 만들기 위한 Cli 메서드
// 지갑을 만들고 주소를 반환
func (c *CLI) newWallet() string {
	wallet := NewKeyStore().CreateWallet().GetAddress()
	return wallet
}
