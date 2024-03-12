package core

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet1 represents UTXO set
type UTXOSet1 struct {
	Blockchain *BlockChain
}

// UTXOSet represents UTXO set
type UTXOSet []Transaction

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
// TODO : 考虑将pubkeyByte改为address
// 暂时不用这个方法
func (u UTXOSet1) FindSpendableOutputs(pubkeyByte []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		// open database
		// 通过 tx.Bucket([]byte(utxoBucket)) 来打开一个名为 utxoBucket 的 Bucket。这个 Bucket 用于存储 UTXO
		// 在 BoltDB 中，Bucket 是一个类似于命名空间的概念，用于存储一组相关的键值对数据。它类似于关系型数据库中的表或 NoSQL 数据库中的集合。Bucket 提供了一种将相关数据组织在一起的方式，使得在 BoltDB 数据库中可以更加有效地组织和检索数据。
		// 每个 Bucket 都有一个唯一的名称，用于在数据库中标识和访问它。在 BoltDB 中，可以通过名称来创建、获取或删除 Bucket
		b := tx.Bucket([]byte(utxoBucket))
		// 使用 c := b.Cursor() 创建了一个游标，这个游标可以遍历 Bucket 中的所有键值对
		// 在 BoltDB 中，游标（Cursor）是一种用于遍历 Bucket 中所有键值对的机制。使用游标可以方便地按顺序或逆序遍历 Bucket 中的数据，而不需要事先知道键的值。
		c := b.Cursor()

		// 使用 for 循环遍历游标中的每个键值对，其中 c.First() 用于获取第一个键值对，然后通过 c.Next() 不断获取下一个键值对，直到遍历完所有键值对为止。
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// 在循环中，对于每个键值对，将键转换为十六进制字符串表示的交易 ID（TxID），然后将值反序列化为一组交易输出（outs）
			txID := hex.EncodeToString(k)
			// 将存储在数据库中的序列化的交易输出数据反序列化为原始的交易输出对象。
			//在很多数据库中，包括 BoltDB 在内，数据存储的时候通常会以一种紧凑的格式进行序列化，以便于存储和传输。在这种情况下，当你需要使用这些数据时，你需要将其反序列化为原始的数据结构，以便于对其进行操作和处理
			outs := DeserializeOutputs(v)

			// 针对每个交易输出，检查它是否被指定的公钥（pubkeyByte）锁定，并且累计的未花费总额（accumulated）小于某个指定的金额（amount）。如果满足这些条件，则将该输出添加到未花费输出的列表中（unspentOutputs），并更新累计金额
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyByte) && accumulated < amount {
					// 每次累加UTXO的余额，直到超过余额后跳出循环
					accumulated += out.Value
					// 如果满足这些条件，则将该输出添加到未花费输出的列表中
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}

		}
		// 返回nil，函数结束
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet1) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet1) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex rebuilds the UTXO set
func (u UTXOSet1) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXOutputs()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet1) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Body.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
