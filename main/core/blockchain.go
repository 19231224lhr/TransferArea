package core

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

// GetBlockChain 获取当前区块链BlockChain结构体的方法
func GetBlockChain() (error, *BlockChain) {
	// 模拟获取当前区块链的区块，前两个空块，第三个区块随即填入一些数据
	block1 := NewBlock()
	block2 := NewBlock()
	// 创建一个块
	block3 := &Block{
		Header: &Header{
			Version:    1,
			TimeStamp:  1630041600,
			PrevBlock:  []byte("00000000000000000000000000000000"),
			MerkelRoot: []byte("5fa9f46d4f3d0b775a3b2ea1c59252f2"),
			state:      0,
		},
		Body: &Body{
			Transactions: []*Transaction{},
		},
	}

	// 新建区块链
	blockchain, err := NewBlockChain()
	if err != nil {
		fmt.Println("创建区块链时出错")
		return err, nil
	}

	// 放入区块
	err = blockchain.AddBlock(block1)
	err = blockchain.AddBlock(block2)
	err = blockchain.AddBlock(block3)
	if err != nil {
		fmt.Println("添加区块时出错")
		return err, nil
	}

	fmt.Println("区块添加完成")
	h, err := block3.Hash()
	if err != nil {
		fmt.Println("计算区块哈希值出错")
		return err, nil
	}

	// 设置blockchain.tip为最新区块哈希值
	blockchain.tip = h

	// 返回
	return err, blockchain
}

// NewBlockChain 创建一个新的区块链
// 首先打开或创建一个 Bolt 数据库，并在其中创建一个名为 "blocks" 的 bucket。然后，我们返回一个新的 BlockChain 实例，其中包含了连接到数据库的指针。
func NewBlockChain() (*BlockChain, error) {
	// 打开或创建一个 Bolt 数据库
	db, err := bolt.Open("myblockchain.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	// 开始一个读写事务
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer func(tx *bolt.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

	// 检查数据库中是否已经有 "blocks" bucket
	b := tx.Bucket([]byte("blocks"))
	if b == nil {
		// 如果不存在，则创建一个新的 bucket
		_, err := tx.CreateBucket([]byte("blocks"))
		if err != nil {
			return nil, err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 返回新的 BlockChain 实例
	return &BlockChain{db: db}, nil
}

// AddBlock 添加一个区块到区块链中
func (bc *BlockChain) AddBlock(block *Block) error {
	return bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		if b == nil {
			return fmt.Errorf("Bucket 'blocks' does not exist")
		}

		// 将区块序列化为字节数组
		blockData := block.Serialize()

		// 将区块数据保存到数据库中，以区块哈希值作为键
		// 先获取区块的哈希值
		h, err := block.Hash()
		if err != nil {
			fmt.Println("区块哈希值计算存在错误")
			return err
		}
		err = b.Put(h, blockData)
		if err != nil {
			return err
		}

		// 更新区块链的 tip
		err = b.Put([]byte("l"), h) // 在示例代码中使用 []byte("l") 作为键存储最新区块的哈希值，是一种简单的做法，它只是作为一个标识符，用来表示最新区块的哈希值。 // 通常情况下，我们可以选择任何唯一的标识符来表示最新区块的哈希值。在实际应用中，可以根据具体需求选择更具描述性的标识符。例如，可以使用 "latest_block_hash"、"tip"、"current_block_hash" 等等。选择一个合适的标识符可以使代码更易读和易于理解
		if err != nil {
			return err
		}

		// 更新区块链结构体中的 tip
		bc.tip = h

		return nil
	})
}

// FindTransaction finds a transaction by its ID
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Body.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.Header.PrevBlock) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

//// FindUTXONoBlockchain 所有未花费的交易的记录(曾经是也会记录)，和所有花费的交易的记录，理论上来说这两个集合的差集就是所有现在未花费的交易记录
//func FindUTXONoBlockchain() (map[string]TXOutputs, error) {
//	UTXO := make(map[string]TXOutputs)  // 未花费(改变)的UTXO，TODO: 注意，并不是现在的未使用输出的集合
//	SpentTXOs := make(map[string][]int) // 已花费(Vout已经有部分被花费)的UTXO(已经不能叫做UTXO了)，key - TXID，value - int数组，存储的是TX的Vout部分已经花费的输出相对编号
//	// 模拟获取目前阶段的blockchain
//	bc, err := NewBlockChain()
//	if err != nil {
//		fmt.Println("! 模拟获取区块链出现错误")
//		return nil, err
//	}
//	bci := bc.Iterator()
//
//	for {
//		block := bci.Next()
//
//		for _, tx := range block.Body.Transactions {
//			txID := hex.EncodeToString(tx.ID) // TXID解码，string结构
//
//		Outputs: // 这段代码使用了标签（label）Outputs:，它用于标记一个循环，以便在内部循环中使用continue语句跳转到外部循环的指定位置
//			// 处理一个交易中的out部分
//			for outIdx, out := range tx.Vout {
//				if SpentTXOs[txID] != nil { // 如果当前交易已经在已花费交易数组中，则检查Vout部分哪些输出已经被花费了
//					for _, spentOutIdx := range SpentTXOs[txID] {
//						if spentOutIdx == outIdx { // 使用相对编号检查
//							continue Outputs // 存在一样的已花费输出跳到外部for循环从下一个重新开始
//						}
//					}
//				}
//
//				// 运行到这一步的都是没有被使用过的output
//				outs := UTXO[txID] // outs是一个结构体，outs.Outputs是一个包含UTXO的数组
//				outs.Outputs = append(outs.Outputs, out)
//				UTXO[txID] = outs // TODO: 注意UTXO集合只进不出，因此只是一个记录，后续如果某个已经被记录的交易被花费了，那么UTXO集合并不会删去这笔交易
//			}
//
//			// 处理一个交易中的in部分(全部加入已使用交易列表中)
//			if tx.IsCoinbase() == false {
//				for _, in := range tx.Vin {
//					inTxID := hex.EncodeToString(in.Txid)
//					// 加入每个UTXO的in中引用的上一个UTXO的out
//					SpentTXOs[inTxID] = append(SpentTXOs[inTxID], in.Vout)
//				}
//			}
//		}
//
//		if len(block.Header.PrevBlock) == 0 {
//			break
//		}
//	}
//
//	return UTXO, nil
//}

// FindUTXOutputs 找出所有的还没有使用的UTXO的outputs，不包括曾经是但后来被使用的UTXO，是真正的UTXO集合
// TODO:为了满足验证的需求，需要新建一个FindUTXOutputs函数，返回带txid和outid的out集合
func (bc *BlockChain) FindUTXOutputs() map[string]TXOutputs {
	USet := make(map[string]TXOutputs) // 存储UTXO的集合
	bci := bc.Iterator()               // 迭代器

	for {
		block := bci.Next() // 新区块
		for _, tx := range block.Body.Transactions {
			txID := hex.EncodeToString(tx.ID) // TXID解码，string结构

			// 先处理out
			for _, out := range tx.Vout {
				a := USet[txID].Outputs
				a = append(a, out)
				b := USet[txID]
				b.Outputs = a
				USet[txID] = b // 这个步骤真的麻烦
				//USet[TxID].Outputs = append(USet[TxID].Outputs, Out)
			}

			for _, in := range tx.Vin {
				// 处理in
				if tx.IsCoinbase() == true {
					fmt.Println("> 出块交易，没有in来源")
				} else {
					if _, ok := USet[hex.EncodeToString(in.Txid)]; !ok {
						fmt.Println("> 在UTXO集合中没有找到使用的输出")
						// 但是这种情况不太可能，coinbase交易除外，普通交易必须使用已有的未使用的交易输出来构造交易
					} else {
						for index, value := range USet[hex.EncodeToString(in.Txid)].Outputs {
							// value.IsUse == true表示该output是特殊类型的UTXO，需要从UTXO集合中去除
							if value.Address == in.Address || value.IsUse == true {
								fmt.Println("> 使用了对应的out，out地址是：", value.Address)
								// 删除对应元素，利用切片
								c := USet[hex.EncodeToString(in.Txid)].Outputs
								c = append(c[:index-1], c[index:]...)
								d := USet[hex.EncodeToString(in.Txid)]
								d.Outputs = c
								USet[hex.EncodeToString(in.Txid)] = d
								//USet[hex.EncodeToString(in.Txid)].Outputs = append(USet[hex.EncodeToString(in.Txid)].Outputs[:index-1], USet[hex.EncodeToString(in.Txid)].Outputs[index:]...)
							}
						}
					}
				}
			}
		}
		if len(block.Header.PrevBlock) == 0 {
			break
		}
	}
	return USet
}

// FindUTXOutputs2 找出所有的还没有使用的UTXO的outputs，不包括曾经是但后来被使用的UTXO，是真正的UTXO集合
// TODO:为了满足验证的需求，需要新建一个FindUTXOutputs函数，返回带txid和outid的out集合
func (bc *BlockChain) FindUTXOutputs2() map[string]TXOutputs2 {
	USet := make(map[string]TXOutputs2) // 存储UTXO的集合
	bci := bc.Iterator()                // 迭代器

	for {
		block := bci.Next() // 新区块
		for _, tx := range block.Body.Transactions {
			txID := hex.EncodeToString(tx.ID) // TXID解码，string结构

			// 先处理out
			for i, out := range tx.Vout {
				a := USet[txID].Outputs
				a = append(a, TXOutput2{
					Value:   out.Value,
					Address: out.Address,
					Outid:   i,
					IsUse:   out.IsUse,
				})
				b := USet[txID]
				b.Outputs = a
				USet[txID] = b // 这个步骤真的麻烦
				//USet[TxID].Outputs = append(USet[TxID].Outputs, Out)
			}

			for _, in := range tx.Vin {
				// 处理in
				if tx.IsCoinbase() == true {
					fmt.Println("> 出块交易，没有in来源")
				} else {
					if _, ok := USet[hex.EncodeToString(in.Txid)]; !ok {
						fmt.Println("> 在UTXO集合中没有找到使用的输出")
						// 但是这种情况不太可能，coinbase交易除外，普通交易必须使用已有的未使用的交易输出来构造交易
					} else {
						for index, value := range USet[hex.EncodeToString(in.Txid)].Outputs {
							// value.IsUse == true表示该output是特殊类型的UTXO，需要从UTXO集合中去除
							if value.Address == in.Address || value.IsUse == true {
								fmt.Println("> 使用了对应的out，out地址是：", value.Address)
								// 删除对应元素，利用切片
								c := USet[hex.EncodeToString(in.Txid)].Outputs
								c = append(c[:index-1], c[index:]...)
								d := USet[hex.EncodeToString(in.Txid)]
								d.Outputs = c
								USet[hex.EncodeToString(in.Txid)] = d
								//USet[hex.EncodeToString(in.Txid)].Outputs = append(USet[hex.EncodeToString(in.Txid)].Outputs[:index-1], USet[hex.EncodeToString(in.Txid)].Outputs[index:]...)
							}
						}
					}
				}
			}
		}
		if len(block.Header.PrevBlock) == 0 {
			break
		}
	}
	return USet
}

// FindUTXOutputsForTran 跨区转账使用，找出所有的还没有使用的UTXO的outputs，不包括曾经是但后来被使用的UTXO，是真正的UTXO集合
func (bc *BlockChain) FindUTXOutputsForTran() map[string][]TXOutputsTran {
	USet := make(map[string][]TXOutputsTran) // 存储UTXO的集合
	bci := bc.Iterator()                     // 迭代器

	var blockNum int = 0 // 区块编号

	for {
		block := bci.Next() // 新区块
		for txNum, tx := range block.Body.Transactions {
			txID := hex.EncodeToString(tx.ID) // TXID解码，string结构

			// 先处理out
			for _, out := range tx.Vout {
				// 填写UTXO的out，交易坐标
				USet[txID] = append(USet[txID], TXOutputsTran{
					Out:          out,
					CoordinatesX: blockNum,
					CoordinatesY: txNum,
				})
				//USet[TxID].Outputs = append(USet[TxID].Outputs, Out)
			}

			for _, in := range tx.Vin {
				// 处理in
				if tx.IsCoinbase() == true {
					fmt.Println("> 出块交易，没有in来源")
				} else {
					if _, ok := USet[hex.EncodeToString(in.Txid)]; !ok {
						fmt.Println("> 在UTXO集合中没有找到使用的输出")
						// 但是这种情况不太可能，coinbase交易除外，普通交易必须使用已有的未使用的交易输出来构造交易
					} else {
						for index, value := range USet[hex.EncodeToString(in.Txid)] {
							// value.IsUse == true表示该output是特殊类型的UTXO，需要从UTXO集合中去除
							if value.Out.Address == in.Address || value.Out.IsUse == true {
								fmt.Println("> 使用了对应的out，out地址是：", value.Out.Address)
								// 删除对应元素，利用切片
								USet[hex.EncodeToString(in.Txid)] = append(USet[hex.EncodeToString(in.Txid)][:index-1], USet[hex.EncodeToString(in.Txid)][index:]...)
								//USet[hex.EncodeToString(in.Txid)].Outputs = append(USet[hex.EncodeToString(in.Txid)].Outputs[:index-1], USet[hex.EncodeToString(in.Txid)].Outputs[index:]...)
							}
						}
					}
				}
			}
		}
		if len(block.Header.PrevBlock) == 0 {
			break
		}
	}
	return USet
}

// Iterator returns a BlockchainIterat
func (bc *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// SignTransaction signs inputs of a Transaction
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey *ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

//// FindUTXO2 查找指定账户所有未使用的UTXO，以便计算账户余额
//func (bc *BlockChain) FindUTXO2(PublicKey ecdsa.PublicKey) (int, map[string]TXOutputs) {
//	UTXO := make(map[string]TXOutputs)  // UTXO集合
//	spentTXOs := make(map[string][]int) // 所有已花费的 UTXO
//	bci := bc.Iterator()
//
//	for {
//		block := bci.Next()
//
//		for _, tx := range block.Body.Transactions {
//			TxID := hex.EncodeToString(tx.ID)
//
//		Outputs:
//			for outIdx, Out := range tx.Vout {
//				// Was the output spent?
//				if spentTXOs[TxID] != nil {
//					for _, spentOutIdx := range spentTXOs[TxID] {
//						if spentOutIdx == outIdx {
//							continue Outputs
//						}
//					}
//				}
//
//				// 是我们需要的钱包地址吗
//
//				// 运行到这一步的都是没有被使用过的output
//				outs := UTXO[TxID]
//				if Out.Address != PublicKey {
//					continue
//				}
//				outs.Outputs = append(outs.Outputs, Out)
//				UTXO[TxID] = outs
//			}
//
//			if tx.IsCoinbase() == false {
//				for _, in := range tx.Vin {
//					inTxID := hex.EncodeToString(in.Txid)
//					// 加入每个UTXO的in中引用的上一个UTXO的out
//					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
//				}
//			}
//		}
//
//		if len(block.Header.PrevBlock) == 0 {
//			break
//		}
//	}
//
//	// 计算钱包余额
//	var WalletValue int = 0
//	for _, u := range UTXO {
//		for _, o := range u.Outputs {
//			WalletValue += o.Value
//		}
//	}
//
//	// 返回钱包余额
//	return WalletValue, UTXO
//}

func (bc *BlockChain) FindTXOut(txid []byte, outid int) TXOutput {
	// 找到对应的交易
	TX, err := bc.FindTransaction(txid)
	if err != nil {
		fmt.Println("! 按照交易id寻找对应交易方法出现错误")
		return TXOutput{}
	}
	// 返回交易对应的out
	return TX.Vout[outid]
}
