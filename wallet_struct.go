package main

import "crypto/ecdsa"

const walletFile = "wallet.json"

type Wallet struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  []byte
}

// 키를 가지고 있는 지갑들을 다수 보관하기 위한 저장소
// Wallets 의 키로는 생성된 주소가 들어갈 것이며 값으로는 그에 해당하는 *Wallet 이 들어감
type KeyStore struct {
	Wallets map[string]*Wallet
}
