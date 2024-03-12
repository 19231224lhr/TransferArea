package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

type Header struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// time when block was created
	TimeStamp int64

	// Height    uint64 // block height

	// hash of the previous block
	PrevBlock []byte

	// hash of the block state
	MerkelRoot []byte

	// state of the block : 0->commit ; 1->valid ; 2->invalid
	state byte
}

type Body struct {
	Transactions []*Transaction
}

type Block struct {
	Header *Header
	Body   *Body
}

func NewBlock() *Block {
	return &Block{
		Header: &Header{},
		Body:   &Body{},
	}
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// Hash 返回块的哈希值
func (b *Block) Hash() ([]byte, error) {
	// 连接块头部字段
	headers := fmt.Sprintf("%d%d%x%x%d", b.Header.Version, b.Header.TimeStamp, b.Header.PrevBlock, b.Header.MerkelRoot, b.Header.state)

	// 创建 SHA-256 哈希对象
	hasher := sha256.New()

	// 将连接后的数据写入哈希对象
	_, err := hasher.Write([]byte(headers))
	if err != nil {
		return nil, err
	}

	// 计算哈希值
	hash := hasher.Sum(nil)

	return hash, nil
}
