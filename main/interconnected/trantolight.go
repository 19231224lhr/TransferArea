package interconnected

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"transfer/core"
	"transfer/wallet"
)

// 转账区 -> 轻计算区

// TXLog 交易记录
type TXLog struct {
	TX           core.Transaction // 交易
	CoordinatesX int              // 区块号
	CoordinatesY int              // 区块中的第几个交易
}

// ToLightComputeReturn 转账区转到轻计算区 返回结构体
type ToLightComputeReturn struct {
	Amount  int              // 转账金额
	Balance int              // 转账区剩余金额
	TxLogs  []TXLog          // 交易记录
	TX      core.Transaction // 构造的交易
}

// ToLightCompute 转账区转到轻计算区，返回TXInput和TXOutput，并返回余额，还有交易记录
// TODO: 目前只支持向单一地址转账
// 单纯转换，Money表示转多少钱过去，APublickey表示转账区用户的公钥，APrivatekey表示转账区用户的私钥，BAddress表示转到轻计算区的地址
func ToLightCompute(w wallet.Wallets, BAddress common.Address, Money int) (error, ToLightComputeReturn) {
	fmt.Println("> 开始执行 转账区 -> 轻计算区 转换函数")

	// 转账区对应的地址
	//AAddress := crypto.PubkeyToAddress(APublickey)

	// 查询目标用户当前余额与可用UTXO
	AUTXO, err := w.GetWalletsBalanceToLight()
	if err != nil {
		fmt.Println("! 跨区转账ToLight获取钱包余额方法出现错误")
		return err, ToLightComputeReturn{}
	}

	// 计算总余额
	var AllMoney int = 0
	for _, v := range AUTXO {
		AllMoney += v.Balance
	}
	if AllMoney < Money {
		fmt.Println("! 您的余额不满足您的跨区转账需求")
		return fmt.Errorf("! 您的余额不满足您的跨区转账需求"), ToLightComputeReturn{}
	}

	// 新建钱包
	ws := core.NewWallet(w.Account, w.Privatekey, w.Publickey)

	var AllUTXOs map[string][]core.TXOutputsTran // 可用的UTXO集合

	// 合并可用的UTXO TODO:有点麻烦，看后续有没有其他的解决方法
	for _, v := range AUTXO {
		for k1, v1 := range v.Txouts {
			for _, k3 := range v1 {
				//a := AllUTXOs[k1]
				//a = append(a, k3)
				//AllUTXOs[k1] = a
				AllUTXOs[k1] = append(AllUTXOs[k1], k3)
			}
		}
	}
	// 方法内包含了对UTXO out的签名
	TX, FinalUTXO, err := core.NewTransactionToLight(ws, BAddress, Money, AllUTXOs)
	if err != nil {
		fmt.Println("! 跨区转账ToLight构造新交易时出现错误")
	}

	// 交易上链
	// TODO: 暂时省略

	// 构造TxLog
	var txlog []TXLog
	for _, value := range FinalUTXO {
		// string -> byte
		txid, err := hex.DecodeString(value.TxID)
		if err != nil {
			fmt.Println("! 跨区交易ToLight string转byte方法出现错误")
			return err, ToLightComputeReturn{}
		}
		// 寻找vin对应的交易，用于轻计算区验证
		err, bc := core.GetBlockChain()
		if err != nil {
			fmt.Println("! 跨区转账ToLight获取区块链相关信息出现错误")
			return err, ToLightComputeReturn{}
		}
		tx, err := bc.FindTransaction(txid)
		if err != nil {
			fmt.Println("! 跨区转账ToLight按照txid寻找交易时出现错误")
			return err, ToLightComputeReturn{}
		}
		txlog = append(txlog, TXLog{
			TX:           tx,
			CoordinatesX: value.Out.CoordinatesX,
			CoordinatesY: value.Out.CoordinatesY,
		})
	}

	// 构造 ToLightComputeReturn
	rm := ToLightComputeReturn{
		Amount:  Money,
		Balance: AllMoney - Money,
		TxLogs:  txlog,
		TX:      *TX,
	}

	// 返回
	return nil, rm
}
