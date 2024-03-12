package core

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	// "transfer/wallet"

	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Transaction UTXO结构
// Type表示UTXO的类型，默认为0：普通的UTXO，1：转账区 -> 轻计算区 跨区UTXO
type Transaction struct {
	ID   []byte // hash of transaction
	Vin  []TXInput
	Vout []TXOutput
	Type int
	// TODO:需要添加一个字段账户，表明这个交易是谁发出来的，也需要提供一个专门的查询函数来查询账户对应的公钥
	Account string // 发送者账户
}

// TransactionWallet 为了解决循环引用的结构体
type TransactionWallet struct {
	Account    string           // 发送者账户
	PrivateKey ecdsa.PrivateKey // 钱包私钥 (账号)
	PublicKey  ecdsa.PublicKey  // 钱包公钥 (密码)
}

// NewWallet 生成一个已有的钱包
func NewWallet(Account string, pr ecdsa.PrivateKey, pu ecdsa.PublicKey) *TransactionWallet {
	w := TransactionWallet{Account, pr, pu}

	return &w
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(wallet *TransactionWallet, to common.Address, amount int, UTXOSet *UTXOSet1) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	from := wallet.GetAddress() // 获得交易地址

	// (based public key byte)Find the unspent, valid outputs to reference in the inputs
	pubKeyByte := crypto.FromECDSAPub(&wallet.PublicKey) // 将PublicKey转为字节形式，为之后做准备
	// 寻找符合余额条件的UTXO
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyByte, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds") // 其实已经提前检查过了
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, from, false}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs

	outputs = append(outputs, *NewTXOutput(amount, to))

	// change
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs, 0, wallet.Account}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, &wallet.PrivateKey)

	return &tx
}

// NewUTXOut 用于UTXO按照余额从高到低排序
type NewUTXOut struct {
	txID  string
	outID int // out的相对位置
	TXOutput
}

// GetAddress returns wallet address
func (w TransactionWallet) GetAddress() common.Address {
	address := crypto.PubkeyToAddress(w.PublicKey)
	return address
}

// NewTransaction 现已支持多钱包多地址转账
func NewTransaction(wallet *TransactionWallet, AddressMoney map[common.Address]int, AUTXO map[string]TXOutputs) (*Transaction, error) {
	// 必要的数据
	AAddress := wallet.GetAddress()                      // 发送方地址
	APrivatekey := wallet.PrivateKey                     // 发送方私钥
	var sortedUTXO []NewUTXOut                           // 排序后的UTXO数组 新
	var FinalUTXO []NewUTXOut                            // 选中的UTXO集合
	pubKeyByte := crypto.FromECDSAPub(&wallet.PublicKey) // 将PublicKey转为字节形式，为之后做准备
	var Accumulated int = 0                              // 记录总余额
	var AllNeed int = 0                                  // 记录需要的总金额
	var Inputs []TXInput                                 // input集合
	var Outputs []TXOutput                               // output集合

	// 计算需要的总金额
	for _, v := range AddressMoney {
		AllNeed += v
	}

	//// UTXO按照余额从高到低排序
	//// 统计有多少out
	//var Num int = 0
	//for _, a := range AUTXO {
	//	for _, _ = range a.Outputs {
	//		Num++
	//	}
	//}
	//// 冒泡排序 复杂度过高，弃用
	//for i := 0; i < Num; i++ {
	//	v4 := TXOutput{Value: 0} // 最终的结果
	//	var TxID string
	//	for k, v := range AUTXO {
	//		v3 := v.Outputs[0] // 第一个out
	//		//v4 := v.Outputs[0] // 最终的结果
	//		for _, v2 := range v.Outputs {
	//			if v2.Value > v3.Value {
	//				v3 = v2
	//			}
	//		}
	//		// for循环结束后，v3就是当前交易ID下余额最高的out
	//		if v3.Value > v4.Value {
	//			// 更换第一名
	//			v4 = v3
	//			TxID = k
	//		}
	//	}
	//	// 将目前余额最高的out放入数组中
	//	AAUTXO = append(AAUTXO, NewUTXOut{
	//		TxID:     TxID,
	//		TXOutput: v4,
	//	})
	//}

	// UTXO按照余额从高到低排序
	// 将UTXO映射到一个临时的切片中进行排序
	for txID, outputs := range AUTXO {
		for outID, output := range outputs.Outputs {
			sortedUTXO = append(sortedUTXO, NewUTXOut{
				txID:     txID,
				outID:    outID,
				TXOutput: output,
			})
		}
	}

	// 对切片进行排序
	sort.Slice(sortedUTXO, func(i, j int) bool {
		return sortedUTXO[i].Value > sortedUTXO[j].Value
	})

	// 选择合适的out
	for _, v := range sortedUTXO {
		if v.TXOutput.IsLockedWithKey(pubKeyByte) {
			FinalUTXO = append(FinalUTXO, v) // 添加到选中的UTXO结果中
			Accumulated += v.TXOutput.Value  // 当前累加的余额
			if Accumulated > AllNeed {
				// 余额足够
				break
			}
		}
	}

	// 构造交易
	// input
	for _, out := range FinalUTXO {
		// txID转为字节
		txID, err := hex.DecodeString(out.txID)
		if err != nil {
			fmt.Println("! txID转为字节出错")
			return nil, err
		}
		// 构造一个交易的其中一个input结构
		input := TXInput{
			Txid:      txID,
			Vout:      out.outID,
			Signature: nil,         // 不知道要填什么字段，源程序也没有填写字段
			Address:   out.Address, // 来源地址
		}
		Inputs = append(Inputs, input) // 添加到最终Inputs集合中
	}
	// output
	for k, v := range AddressMoney {
		Outputs = append(Outputs, *NewTXOutput(v, k))
	}
	//Outputs = append(Outputs, *NewTXOutput(Amount, BAddress))
	// 找零
	if Accumulated > AllNeed {
		Outputs = append(Outputs, *NewTXOutput(Accumulated-AllNeed, AAddress)) // TODO: 当前版本只支持将零钱转回自己的账户
	}
	// 构造交易
	TX := Transaction{
		ID:   nil,
		Vin:  Inputs,
		Vout: Outputs,
		Type: 0, // Type值为0表示普通交易
	}
	TX.ID = TX.Hash() // 交易ID
	// 交易签名
	err := SignTransactionNoBlockchain(&TX, &APrivatekey)
	if err != nil {
		fmt.Println("! 交易签名方法出现错误")
		return nil, err
	}
	// 新的交易构造完成

	// 返回交易
	return &TX, nil
}

// NewOutToLight 用于跨区转账交易To轻计算区Output按照余额排序
type NewOutToLight struct {
	TxID  string
	OutID int // out的相对位置
	Out   TXOutputsTran
}

// NewTransactionToLight 跨区转账To轻计算区使用的交易构造函数
func NewTransactionToLight(wallet *TransactionWallet, BAddress common.Address, Amount int, AUTXO map[string][]TXOutputsTran) (*Transaction, []NewOutToLight, error) {
	// 必要的数据
	AAddress := wallet.GetAddress()                      // 发送方地址
	APrivatekey := wallet.PrivateKey                     // 发送方私钥
	var SortedUTXO []NewOutToLight                       // 排序后的UTXO数组 新
	var FinalUTXO []NewOutToLight                        // 选中的UTXO集合
	pubKeyByte := crypto.FromECDSAPub(&wallet.PublicKey) // 将PublicKey转为字节形式，为之后做准备
	var Accumulated int = 0                              // 记录总余额
	var Inputs []TXInput                                 // input集合
	var Outputs []TXOutput                               // output集合

	// UTXO按照余额从高到低排序
	// 将UTXO映射到一个临时的切片中进行排序
	for txID, outputs := range AUTXO {
		for outID, output := range outputs {
			SortedUTXO = append(SortedUTXO, NewOutToLight{
				TxID:  txID,
				OutID: outID,
				Out:   output,
			})
		}
	}

	// 对切片进行排序
	sort.Slice(SortedUTXO, func(i, j int) bool {
		return SortedUTXO[i].Out.Out.Value > SortedUTXO[j].Out.Out.Value
	})

	// 选择合适的out
	for _, v := range SortedUTXO {
		if v.Out.Out.IsLockedWithKey(pubKeyByte) {
			FinalUTXO = append(FinalUTXO, v) // 添加到选中的UTXO结果中
			Accumulated += v.Out.Out.Value   // 当前累加的余额
			if Accumulated > Amount {
				// 余额足够
				break
			}
		}
	}

	// 构造交易 特殊交易ToLight，需要将UTXO结构体中的Type值设置为1
	// input
	for _, out := range FinalUTXO {
		// txID转为字节
		txID, err := hex.DecodeString(out.TxID)
		if err != nil {
			fmt.Println("! txID转为字节出错")
			return nil, nil, err
		}
		// 构造一个交易的其中一个input结构
		input := TXInput{
			Txid:      txID,
			Vout:      out.OutID,
			Signature: nil,                 // 不知道要填什么字段，源程序也没有填写字段，后续用签名函数签名
			Address:   out.Out.Out.Address, // 地址填写来自于哪里的out地址
		}
		Inputs = append(Inputs, input) // 添加到最终Inputs集合中
	}
	// output
	Outputs = append(Outputs, TXOutput{
		Value:   Amount,
		Address: BAddress,
		IsUse:   true,
	}) // BAddress是轻计算区对应地址，true表示该out属于跨区转账类型的out，已被使用
	// 找零
	if Accumulated > Amount {
		Outputs = append(Outputs, TXOutput{
			Value:   Accumulated - Amount,
			Address: AAddress,
			IsUse:   false,
		}) // AAddress是转账区对应地址，false表示该out还留在转账区中，没有被使用
	}
	// 构造交易
	TX := Transaction{
		ID:      nil,
		Vin:     Inputs,
		Vout:    Outputs,
		Type:    1, // Type为1表示是跨区转账的特殊交易ToLight
		Account: wallet.Account,
	}
	TX.ID = TX.Hash() // 交易ID
	// 交易签名
	err := SignTransactionNoBlockchain(&TX, &APrivatekey)
	if err != nil {
		fmt.Println("! 交易签名方法出现错误")
		return nil, nil, err
	}
	// 新的交易构造完成

	// 返回交易
	return &TX, FinalUTXO, nil
}

// NewCoinbaseTX 新建ToTran Coinbase交易
// 目前铸币交易能自己构造的也只有ToTran的跨链交易
func NewCoinbaseTX(FromAddress common.Address, Address common.Address, Amount int, UserAccount string) *Transaction {
	// tx.Vin只有一个，且没有Txid，Vout = -1
	var Inputs []TXInput   // input集合
	var Outputs []TXOutput // output集合
	// input
	input := TXInput{
		Txid:      nil,
		Vout:      -1,
		Signature: nil,
		Address:   FromAddress, // 对应轻计算区的账户地址
		IsToTran:  true,        // true表示是跨链交易ToTran
	}
	Inputs = append(Inputs, input)
	// output
	Outputs = append(Outputs, *NewTXOutput(Amount, Address)) // 不用找零，跨链交易ToTran将所有钱全部转入转账区对应账户
	// Coinbase交易不用签名
	// 构造交易
	TX := Transaction{
		ID:      nil,
		Vin:     Inputs,
		Vout:    Outputs,
		Type:    2, // 2表示跨链交易ToTran
		Account: UserAccount,
	}
	TX.ID = TX.Hash() // 交易ID
	return &TX
}

// NewTransactionToTran 跨链交易ToTran构造新交易
// BAddress是转账区对应的地址，Amount是目标转账金额
func NewTransactionToTran(wallet *TransactionWallet, BAddress common.Address, Amount int) (*Transaction, error) {
	// 必要的数据
	AAddress := wallet.GetAddress()  // 发送方地址
	APrivatekey := wallet.PrivateKey // 发送方私钥
	var Inputs []TXInput             // input集合
	var Outputs []TXOutput           // output集合

	// 构造交易
	// input
	input := TXInput{
		Txid:      nil,
		Vout:      0,
		Signature: nil,      // 不知道要填什么字段，源程序也没有填写字段
		Address:   AAddress, // 对应轻计算区的账户地址
		IsToTran:  true,     // true表示是跨链交易ToTran
	}
	Inputs = append(Inputs, input)
	// output
	Outputs = append(Outputs, *NewTXOutput(Amount, BAddress)) // 不用找零，跨链交易ToTran将所有钱全部转入转账区对应账户
	// 构造交易
	TX := Transaction{
		ID:      nil,
		Vin:     Inputs,
		Vout:    Outputs,
		Type:    2, // 2表示跨链交易ToTran
		Account: wallet.Account,
	}
	TX.ID = TX.Hash() // 交易ID
	// 交易签名
	err := SignTransactionNoBlockchain(&TX, &APrivatekey)
	if err != nil {
		fmt.Println("! 交易签名方法出现错误")
		return nil, err
	}
	// 新的交易构造完成

	// 返回交易
	return &TX, nil
}

// SignTransactionNoBlockchain 不使用blockchain结构体的交易签名方法
func SignTransactionNoBlockchain(TX *Transaction, privKey *ecdsa.PrivateKey) error {
	// 交易Inputs中对应的前置完整交易
	PrevTXs := make(map[string]Transaction)

	// 模拟获取目前阶段的blockchain
	err, blockchain := GetBlockChain()
	if err != nil {
		fmt.Println("! 获取区块链信息出现错误")
		return err
	}

	for _, vin := range TX.Vin {
		PrevTX, err := blockchain.FindTransaction(vin.Txid)
		if err != nil {
			fmt.Println("! 按照TXid寻找交易方法出现错误")
			return err
		}
		// 前置交易数组
		PrevTXs[hex.EncodeToString(PrevTX.ID)] = PrevTX
	}

	// 交易签名
	TX.Sign(privKey, PrevTXs)
	// 返回
	return nil
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, vin.Address, false})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.Address, false})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, 0, tx.Account}

	return txCopy
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	// ! 原来代码的逻辑是对每一个tx in 都做一次签名

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].Address = prevTx.Vout[vin.Vout].Address // ?

		dataToSign := fmt.Sprintf("%x\n", txCopy) // 将交易对象 txCopy 格式化为十六进制字符串

		r, s, err := ecdsa.Sign(rand.Reader, privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		// TODO:现在知道了可以使用公钥来进行签名验证 ecdsa.Verify()方法

		tx.Vin[inID].Signature = signature
		// txCopy.Vin[inID].Address = nil
	}
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// GetTXByCoordinates TODO:根据坐标返回交易
func GetTXByCoordinates(CoordinatesX int, CoordinatesY int) (Transaction, error) {
	// 模拟获取目前阶段的blockchain
	err, blockchain := GetBlockChain()
	if err != nil {
		fmt.Println("! 模拟获取区块链出现错误")
		return Transaction{}, err
	}

	bci := blockchain.Iterator() // 迭代器
	var block *Block
	var i int

	// 找到目标区块
	for i = 0; i <= CoordinatesX; i++ {
		block = bci.Next() // 新区块
	}

	if CoordinatesY > len(block.Body.Transactions) {
		return Transaction{}, fmt.Errorf("! 坐标范围错误")
	}

	return *block.Body.Transactions[CoordinatesY], nil
}
