package main

// 트랜잭션은 기본적으로 Block 에 포함
//  Data 라는 필드를 블록에 포함시켰었는데 그것 대신에 Transaction들로 변경될 것임
// 먼저 거래는 입력 값(Input Transaction)과 출력 값(Output Transaction)이 있으며, 하나의 트랜잭션은 다수의 입력과 출력을 가질 수 있기 때문에 배열로 표현
// 록의 해시값이 있듯이 거래도 ID 로 불리는 해시값이 있음

// 트랜잭션의 경우 입력보다 출력이 우선
// 출력 : 자금을 얼마나 어디로 송금할 것인가? (ex. 내가 가진 돈을 얼마만큼 누군가에게 지불하는 것)
// 입력 : 돈이 어디에서 왔는지 그 원천에 대한 이야기(TXOutput을 참조하는 필드가 필요)
// 10. 주소를 이용한 거래기능 추가로 인한 변경점
//		- 거래에서 주소를 사용
//		- 서명을 통한 서명검증을 위해 TXInput과 TXOutput을 변경
//		- 스크립트언어가 아닌 공개키해시 사용(Base58CheckDecode)
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// P2PKH(Pay-To-Public-Key-Hash), P2SH(Pay-To-Script-Hash)의 추가적인 내용 숙지 필요
type TXOutput struct {
	Value      uint64 // 코인
	PubKeyHash []byte // 어디로(받는사람의 공개키)
	//ScriptPubKey string // 어디로(받는사람의 공개키)
}

// 과거에 내게 들어온 자본의 흐름, 즉 이전 트랜잭션의 출력값을 참조를 위한 Txid와 Vout 필드
// vout의 경우 하나의 트랜잭션은 다수의 출력을 가질 수 있기때문에 지목하기 위한 인덱스를 위한 필드가 필요
type TXInput struct {
	Txid      []byte // 참조한 트랜잭션의 ID
	Vout      int    // 해당 트랜잭션이 가진 출력값의 인덱스
	Signature []byte // 디지털 서명(개인키를 사용하여 생성)
	PubKey    []byte // 서명을 검증하기 위한 발신자의 공개키
}
