package main

import (
	"bytes"
	"log"
)

// 거래위한 페이지
// 거래 조건 : 현재 가지고 있는 자금을 추적하고, 현재 자금보다 보내는 자금이 더 큰 경우 실행을 중지하고, 충분하다면 금액을 보낸뒤 잔액(Change)를 만듬
// (코인은 화폐처럼 분리할 수 없으며, 모두 소비한 뒤 잔액을 표시해야함)
// 거래 순서
// 		1. 거래를 위해 TXInput 을 구성
//
//  10. 주소를 이용한 거래기능으로 인한 변경점
//		- 지갑간의 주소를 통한 거래를 위해 KeyStore 를 이곳에서 사용하도록 변경
//		- 공개키해시를 사용하여 검증
//		- NewTXOutput() 메서드를 사용하여 출력을 구성

func (bc *Blockchain) Send(value uint64, from, to string) *Transaction {
	var txin []TXInput
	var txout []TXOutput
	keyStore := NewKeyStore()

	wallet := keyStore.Wallets[from]
	UTXs := bc.FindUnspentTransactions(HashPubKey(wallet.PubKey))
	var acc uint64

Work:
	for _, tx := range UTXs {
		for outIdx, out := range tx.Vout {
			// if out.ScriptPubKey == from && acc < value {
			// 	acc += out.Value
			// 	txin = append(txin, TXInput{tx.ID, outIdx, from})
			// }
			if bytes.Compare(out.PubKeyHash, HashPubKey(wallet.PubKey)) == 0 && acc < value {
				acc += out.Value
				txin = append(txin, TXInput{tx.ID, outIdx, nil, wallet.PubKey})
			}
			if acc >= value {
				break Work
			}
		}
	}

	if value > acc {
		log.Panic("ERROR: NOT enough funds")
	}

	// txout = append(txout, TXOutput{value, to})
	// if acc > value {
	// 	txout = append(txout, TXOutput{acc - value, from})
	// }
	txout = append(txout, *NewTXOutput(value, to))
	if acc > value {
		txout = append(txout, *NewTXOutput(acc-value, from))
	}

	tx := NewTransaction(txin, txout)
	bc.SignTransaction(wallet.PrivKey, tx)

	return tx
}
