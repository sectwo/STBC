package main

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
)

// UTXO(Unspent Transaction Output) :  소비되지 않은 거래 출력 값
// 소비(spent) : 내가 가지고 있는 자금을 다른 주체에게 지불하는 행위, 이를위해 내가 가지고 있는 자금 현황에 대해 파악
// 내가 가지고 있는 자금(Balance) : 나의 공개키(Public Key)로 묶여있는 TXOutput이며, TXOutput 은 다른 거래의 TXInput 에 참조된 적이 없어야함
// UTXO에 대한 결론 : 나의 공개키로 묶여있으며 다른 트랜잭션에 참조된 적이 없는 트랜잭션 출력 값의 합이 내가 가지고 있는 총 자금

// UTXO를 찾기위한 메서드
// UTXO를 찾기위한 과정 :
//  1. 맨 마지막 블록에서 역순으로 제네시스 블록까지 진행
//  2. 모든 TXOutput 에서 TXInput 에 참조된 적이 있는 트랜잭션 조사(입력값이 없는 코인베이스 트랜잭션을 제외한 트랜잭션에 대해 TXInput을 조사)
//  3. 소비(Spent)된 트랜잭션 출력(Spent Transaction Outputs) 집합을 찾기
//  4. 체인을 따라가며 이미 소비된 트랜잭션 집합을 제외하게 되면 소비되지 않은 트랜잭션 집합인 UTXO를 찾을 수 있음
// 변경점 추가
// 10. 주소를 이용한 거래기능으로 인한 변경점
//		- 이제 거래는 지갑 주소를 이용하여 이루어 지기때문에 입력 파라메타들과 조건문들 변경
//		- UTX(Unspent Transaction)을 찾을 떄 조건문 변경
//		- 소비된 출력을 찾을 때 쓰는 조건문 변경
//			- 기존에는 TXInput.ScriptSig 값으로 출력의 사용여부를 찾았으나 지금은 구현에 변화를 주었기 때문에 공개키 해시(PubKeyHash)로 비교
//		-  UTXO 를 찾을 때 조건문 변경
//			- 기존에는 TXOuput.ScriptPubKey를 비교하였지만, 이제는 TXOutput.PubKeyHash값으로 출력값을 잠그기때문에 이 부분을 변경
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []*Transaction {
	bci := NewBlockchainIterator(bc)

	spentTXOs := make(map[string][]int)
	var unspentTXs []*Transaction

	// 다음 블럭이 존재(TURE)면 반복 그렇지 않으면 반복정지
	for bci.HasNext() {
		for _, tx := range bci.Next().Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// TXOutput 에서 이미 소비된 트랜잭션에 대해서는 처리하지 않는다.
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// address 의 공개키로 출력이 되었다는 것은 address 에게 자금을 보냈다는 이야기다.
				// 그 이외의 트랜잭션은 아직 소비되지 않은 트랜잭션이다.
				// 10. 주소를 이용한 거래기능으로 인한 변경점
				//		- UTX(Unspent Transaction)을 찾을 떄 조건문 변경
				// if out.ScriptPubKey == address {
				// 	unspentTXs = append(unspentTXs, tx)
				// }
				if bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			// 입력이 없는 코인베이스 트랜잭션은 제외.
			if !tx.IsCoinbase() {
				// TXInput 을 조사하여 이미 소비된 출력 집합을 얻는다.
				for _, in := range tx.Vin {
					// 서명을 address 가 했음은 address 가 지불을 위해
					// 해당 트랜잭션 출력을 사용했다는 뜻이다.
					// if in.ScriptSig == address {
					// 	hash := hex.EncodeToString(in.Txid)
					// 	spentTXOs[hash] = append(spentTXOs[hash], in.Vout)
					// }
					if in.UsesKey(pubKeyHash) {
						hash := hex.EncodeToString(in.Txid)
						spentTXOs[hash] = append(spentTXOs[hash], in.Vout)
					}
				}
			}
		}
	}

	return unspentTXs
}

// 코인베이스 트랜잭션 확인을 위한 메서드
func (tx *Transaction) IsCoinbase() bool {
	return bytes.Compare(tx.Vin[0].Txid, []byte{}) == 0 && tx.Vin[0].Vout == -1 && len(tx.Vin) == 1
}

// 특정 주소가 가진 자금을 확인하기 위한 메서드
// 10. 주소를 이용한 거래기능으로 인한 변경점
//		- 기존 address에서 공개키 해시를 받는걸로 변경
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			// if out.ScriptPubKey == address {
			// 	UTXOs = append(UTXOs, out)
			// }
			if bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// 특정 주소가 가지고 있는 자금의 총 합을 구하기위한 메서드
// 10. 주소를 이용한 거래기능으로 인한 변경점
//		- 기존 address에서 공개키 해시를 받는걸로 변경
//		- Base58CheckDecode 를 통해 공개키 해시를 얻어서 .FindUTXO() 를 호출
func (bc *Blockchain) GetBalance(address string) uint64 {
	var balance uint64

	pubKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	for _, out := range bc.FindUTXO(pubKeyHash) {
		balance += out.Value
	}

	return balance
}
