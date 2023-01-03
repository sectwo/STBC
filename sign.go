package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"log"
	"math/big"
)

// 서명을 위한 메서드
// 서명 생성 방법은 다음과 같음 명할 데이터의 해시에 개인키를 넣고 서명 알고리즘을 사용하여 서명 생성
// =====================================================================================================================
// 서명 알고리즘
// Sig = F`sig(F`hash(M), k)
// Sig = 서명, F`sig = 서명 알고리즘, F`hash = 해시함수, M = 서명할 데이터, k = 개인키
// =====================================================================================================================
// 서명을 위한 데이터는 송신자와 수신자의 식별정보를 사용하며, 공개키 해시(Public Key Hash)로 표현
// 이를위해, 거래를 바로 해싱하지 않고 거래를 복사한 뒤 값을 일부 수정하여 해싱
// ECDSA 알고리즘 사용
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]*Transaction) {
	if tx.IsCoinbase() {
		return
	}
	// 거래의 복사본 생성
	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Vin {
		// 서명 대상 데이터 구성 및 초기화, 데이터를 대상으로 해싱 .SetID()
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTXs[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash
		txCopy.SetID()
		txCopy.Vin[inID].PubKey = nil

		// 서명 생성, 개인키와 서명한 데이터의 해시를 넣자.
		r, s, err := ecdsa.Sign(rand.Reader, privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature

		// tx.Vin[inID].Signature = append(r.Bytes(), s.Bytes()...)
	}
}

// 대상 트랜잭션의 복사본을 생성을 위한 메서드
func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Vin {
		inputs = append(inputs, TXInput{in.Txid, in.Vout, nil, nil})
	}
	for _, out := range tx.Vout {
		outputs = append(outputs, TXOutput{out.Value, out.PubKeyHash})
	}

	return &Transaction{nil, inputs, outputs}
}

// 서명 검증을 위한 메서드
// 서명을 검증하기 위해서는 해시된 데이터, 서명(R,S), 공개키(X,Y)가 필요하며, 파라매터로는 .Sign() 과 마찬가지로 이전 트랜잭션들을 받음
// 검증을 위해 서명에 사용된 데이터를 해시해서 비교
func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, in := range tx.Vin {
		// 서명에 사용할 데이터를 생성하고 해싱.
		// 여기서 생성된 해시는 검증을 위해 만든 것.
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTXs[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash
		txCopy.SetID()
		txCopy.Vin[inID].PubKey = nil

		// 서명의 R, S 값 얻기
		var r, s big.Int

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:sigLen/2])
		s.SetBytes(in.Signature[sigLen/2:])

		// 공개키의 X, Y 값 얻기
		var x, y big.Int

		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:keyLen/2])
		y.SetBytes(in.PubKey[keyLen/2:])

		// 공개키 생성
		pubKey := ecdsa.PublicKey{curve, &x, &y}

		// 검증
		if isVerified := ecdsa.Verify(&pubKey, txCopy.ID, &r, &s); !isVerified {
			return false
		}
	}

	return true
}

// 11. 디지털 서명기능 추가로 인한 메서드
// 블록체인에서 파라매터로 넘어온 txid 에 해당하는 트랜잭션을 얻어옴
func (bc *Blockchain) FindTransaction(txid []byte) *Transaction {
	bci := NewBlockchainIterator(bc)
	for bci.HasNext() {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, txid) == 0 {
				return tx
			}
		}
	}

	return nil
}

// 트랜잭션에 서명을 하기 위한 메서드
func (bc *Blockchain) SignTransaction(privKey *ecdsa.PrivateKey, tx *Transaction) {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Vin {
		prevTXs[hex.EncodeToString(in.Txid)] = bc.FindTransaction(in.Txid)
	}

	tx.Sign(privKey, prevTXs)
}

// 해당 트랜잭션의 서명을 검증하기 위한 메서드
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Vin {
		prevTXs[hex.EncodeToString(in.Txid)] = bc.FindTransaction(in.Txid)
	}

	return tx.Verify(prevTXs)
}
