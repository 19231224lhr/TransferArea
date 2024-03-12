package core

import (
	"github.com/ethereum/go-ethereum/common"
)

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte // Tx Hash
	Vout      int    // Previous Tx Output Index
	Signature []byte
	Address   common.Address // 来源output的地址，如果是跨链交易ToTran则表示来源于轻计算区哪个地址
	IsToTran  bool           // 是否是跨链交易ToTran的input，true表示是跨链交易ToTran的input，false表示是正常交易
}

// TXInputs 全面对标TXOutputs
type TXInputs struct {
	Inputs []TXInput
}

// // UsesKey checks whether the address initiated the transaction
// func (in *TXInput) UsesKey(pubKeyByte []byte) bool {
// 	pk, _ := crypto.UnmarshalPubkey(pubKeyByte)
// 	lockingHash := crypto.PubkeyToAddress(*pk)

// 	return bytes.Compare(lockingHash[:], in.Address[:]) == 0
// }
