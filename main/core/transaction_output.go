package core

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TXOutput represents a transaction output
// IsUse 是表明该UTXO是否被使用的标识，用于跨区转账中，默认值false表示还没有被使用，true表示已经被使用
type TXOutput struct {
	Value   int
	Address common.Address
	IsUse   bool
}

// TXOutput2 用于验证
type TXOutput2 struct {
	Value   int
	Address common.Address
	Outid   int // 用于验证
	IsUse   bool
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address common.Address) *TXOutput {
	txo := &TXOutput{value, address, false}
	return txo
}

// Lock signs the output
func (out *TXOutput) Lock(address common.Address) {
	out.Address = address
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyByte []byte) bool {
	pk, _ := crypto.UnmarshalPubkey(pubKeyByte)
	addr := crypto.PubkeyToAddress(*pk)
	return bytes.Compare(out.Address[:], addr[:]) == 0
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// TXOutputs2 collects TXOutput2
type TXOutputs2 struct {
	Outputs []TXOutput2
}

// TXOutputsTran 用于跨区交易时的UTXO Output
type TXOutputsTran struct {
	Out          TXOutput // UTXO的out
	CoordinatesX int      // 区块号
	CoordinatesY int      // 区块中的第几个交易
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
