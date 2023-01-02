package main

import "log"

// 거래위한 페이지
// 거래 조건 : 현재 가지고 있는 자금을 추적하고, 현재 자금보다 보내는 자금이 더 큰 경우 실행을 중지하고, 충분하다면 금액을 보낸뒤 잔액(Change)를 만듬
// (코인은 화폐처럼 분리할 수 없으며, 모두 소비한 뒤 잔액을 표시해야함)
// 거래 순서
// 		1. 거래를 위해 TXInput 을 구성

func (bc *Blockchain) Send(value uint64, from, to string) *Transaction {
	var txin []TXInput
	var txout []TXOutput

	UTXs := bc.FindUnspentTransactions(from)
	var acc uint64

Work:
	for _, tx := range UTXs {
		for outIdx, out := range tx.Vout {
			if out.ScriptPubKey == from && acc < value {
				acc += out.Value
				txin = append(txin, TXInput{tx.ID, outIdx, from})
			}
			if acc >= value {
				break Work
			}
		}
	}

	if value > acc {
		log.Panic("ERROR: NOT enough funds")
	}

	txout = append(txout, TXOutput{value, to})
	if acc > value {
		txout = append(txout, TXOutput{acc - value, from})
	}

	return NewTransaction(txin, txout)
}
