package verify

import (
	"crypto/ecdsa"
	"fmt"
	"transfer/core"
	"transfer/wallet"
)

// 节点收到交易后的验证函数

// 金额，签名

func TxVerify(TX core.Transaction) bool {
	// 交易发起用户
	Account := TX.Account
	// 系统查询该Acount对应的钱包信息
	IsFind, W := wallet.GetAccountWallets(Account)
	if IsFind == false {
		fmt.Println("! 没有找到该交易对应的钱包信息")
		return false
	}
	// 找出当前钱包所有可用的UTXO
	UTXOs, err := W.GetWalletsBalance2()
	if err != nil {
		fmt.Println("验证交易时找出当前浅薄所有可用的UTXO出现错误")
		return false
	}
	// 当前交易的Input是否已经被花费了
	IsUsed := FindOutIsUsed(UTXOs, TX)
	if IsUsed == false {
		fmt.Println("当前交易的Input可能已经被花费了")
		return false
	}

	// 验证交易签名
	// 根据账户查询账户对应的公钥
	publickey, IsFind2 := wallet.GetPublickey(TX.Account)
	if IsFind2 == false {
		fmt.Println("没有找到账户对应的公钥")
		return false
	}
	IsSignVerify := VerifySign(TX, publickey)
	if IsSignVerify == false {
		fmt.Println("交易签名验证失败")
		return false
	}

	// 验证结束
	return true
}

// FindOutIsUsed 检查交易的Vin部分中的输入是否使用过了
func FindOutIsUsed(UTXOs []wallet.WalletsBalance2, TX core.Transaction) bool {
	// 模拟获取目前阶段的blockchain
	err, blockchain := core.GetBlockChain()
	if err != nil {
		fmt.Println("! 模拟获取区块链出现错误")
		return false
	}
	// 确认的input总数
	var inputnum int = 0
outerLoop:
	for _, t := range TX.Vin {
		// 根据交易ID与vout的相对序号寻找对应的vout
		Out := blockchain.FindTXOut(t.Txid, t.Vout)
		// 当前可用UTXO集合中是否存在该out
		for _, w := range UTXOs {
			outs, exist := w.Txouts[string(t.Txid)]
			if exist == true {
				// 找到了key，看是否存在value
				for _, out := range outs.Outputs {
					if out.Outid == t.Vout {
						if Out.IsUse == out.IsUse && Out.Value == out.Value && Out.Address == out.Address {
							// 找到了对应的output
							inputnum++
							break outerLoop
						}
					}
				}
			}
		}
	}
	if inputnum != len(TX.Vin) {
		fmt.Println("TX.Vin验证失败")
		return false
	}
	return true
}

// VerifySign 验证交易签名
func VerifySign(TX core.Transaction, Publickey ecdsa.PublicKey) bool {
	for _, in := range TX.Vin {
		// TODO:暂时不太清楚验证的方法与哈希值的选取，临时代替
		isvalid := ecdsa.VerifyASN1(&Publickey, make([]byte, 5), in.Signature)
		if isvalid == false {
			return false
		}
	}
	return true
}
