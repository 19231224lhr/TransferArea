package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"transfer/core"
)

// 测试更改

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey // 钱包私钥 (账号)
	PublicKey  ecdsa.PublicKey  // 钱包公钥 (密码)
}

// AllWalletSet 从数据库找到的所有钱包数据集合，用于查询(但是怎么获取到这个数据还需要讨论)
var AllWalletSet []Wallet

// LoginWallet 钱包登录程序(单钱包)
// 输入账户名和密码即可，系统会自动验证并返回对应的公私钥对（加密），早期版本先不考虑加密 TODO:后期换成多钱包登录方法
func LoginWallet() (int, error, ecdsa.PublicKey, ecdsa.PrivateKey, string) {
	var PrivateKey ecdsa.PrivateKey
	var PublicKey ecdsa.PublicKey

	var account string
	var password string
	// 1 输入钱包账号和密码
	fmt.Println("> 请输入您的钱包账号：")
	num1, err := fmt.Scanln(&account)
	if num1 != 1 || err != nil {
		return 0, err, PublicKey, PrivateKey, account // 0表示错误，1表示正确
	} else {
		fmt.Println("> 您输入的钱包账号是：", account)
	}
	fmt.Println("> 请输入您的钱包密码：")
	num2, err := fmt.Scanln(&password)
	if num2 != 1 || err != nil {
		return 0, err, PublicKey, PrivateKey, account
	} else {
		fmt.Println("> 您输入的钱包密码是：", password)
	}

	// 2 验证钱包账号和密码
	v, PublicKey, PrivateKey := WalletVerify(account, password)
	if v == false {
		fmt.Println("! 您输入的钱包公私钥对有误，登录程序退出")
		return 0, nil, PublicKey, PrivateKey, account
	} else {
		fmt.Println("> 登录成功")
		return 1, nil, PublicKey, PrivateKey, account
	}
}

// AccountECDSA 模拟钱包服务器端数据
type AccountECDSA struct {
	Password   string           // 账户密码
	Publickey  ecdsa.PublicKey  // 账户公钥
	PrivateKey ecdsa.PrivateKey // 账户私钥
}

// AccountData key - 账户账号，value - struct{账户密码,账户公钥,账户私钥}
var AccountData = make(map[string]AccountECDSA)

// WalletVerify 查询账号密码，返回值第一个值为true表示成功，第一个值为false表示失败
func WalletVerify(a, p string) (bool, ecdsa.PublicKey, ecdsa.PrivateKey) {
	// TODO: 这里应该接入服务器查询账号密码是否正确，为了模拟就不用了
	for k, v := range AccountData {
		if a == k {
			if p == v.Password {
				fmt.Println("账号密码验证通过")
				// 这里服务器验证成功后，会使用账户密码解锁存储在服务器的钱包公私钥对并返回，这里我们模拟返回一个
				return true, v.Publickey, v.PrivateKey
			}
		} else {
			continue
		}
	}
	var emptyPublicKey ecdsa.PublicKey
	var emptyPrivateKey ecdsa.PrivateKey
	return false, emptyPublicKey, emptyPrivateKey
}

// GetBalance 查询钱包余额
func (w Wallet) GetBalance() (int, map[string]core.TXOutputs, error) {
	// 方法：查询这个公钥地址所有的UTXO的转入转出，并进行合并
	// return core.BlockChain{}.FindUTXO2(PublicKey) // some bugs

	var Changes int = 0 // 钱包余额

	AUTXO := make(map[string]core.TXOutputs, 10) // 当前用户可用的UTXO集合

	// 将Publickey转为Address
	Address := crypto.PubkeyToAddress(w.PublicKey) // 钱包公钥对应的地址
	// 模拟获取目前阶段的blockchain
	err, blockchain := core.GetBlockChain()
	if err != nil {
		fmt.Println("! 模拟获取区块链出现错误")
		return 0, nil, err
	}
	// 寻找目前所有可用的UTXO
	UTXOSet := blockchain.FindUTXOutputs()
	// 遍历所有的可用UTXO，寻找地址一样的未使用交易输出
	for key, value := range UTXOSet {
		for _, out := range value.Outputs {
			if out.Address == Address {
				fmt.Println("> 找到一个可用的未使用交易输出")
				Changes += out.Value // 累加余额

				// 角色可用的UTXO集合
				a := AUTXO[key].Outputs
				a = append(a, out)
				b := AUTXO[key]
				b.Outputs = a
				AUTXO[key] = b
				//AUTXO[key].Outputs = append(AUTXO[key].Outputs, out)
			}
		}
	}
	// 返回
	return Changes, AUTXO, nil
}

// GetBalance2 查询钱包余额
func (w Wallet) GetBalance2() (int, map[string]core.TXOutputs2, error) {
	// 方法：查询这个公钥地址所有的UTXO的转入转出，并进行合并
	// return core.BlockChain{}.FindUTXO2(PublicKey) // some bugs

	var Changes int = 0 // 钱包余额

	AUTXO := make(map[string]core.TXOutputs2, 10) // 当前用户可用的UTXO集合

	// 将Publickey转为Address
	Address := crypto.PubkeyToAddress(w.PublicKey) // 钱包公钥对应的地址
	// 模拟获取目前阶段的blockchain
	err, blockchain := core.GetBlockChain()
	if err != nil {
		fmt.Println("! 模拟获取区块链出现错误")
		return 0, nil, err
	}
	// 寻找目前所有可用的UTXO
	UTXOSet := blockchain.FindUTXOutputs2()
	// 遍历所有的可用UTXO，寻找地址一样的未使用交易输出
	for key, value := range UTXOSet {
		for _, out := range value.Outputs {
			if out.Address == Address {
				fmt.Println("> 找到一个可用的未使用交易输出")
				Changes += out.Value // 累加余额

				// 角色可用的UTXO集合
				a := AUTXO[key].Outputs
				a = append(a, out)
				b := AUTXO[key]
				b.Outputs = a
				AUTXO[key] = b
				//AUTXO[key].Outputs = append(AUTXO[key].Outputs, out)
			}
		}
	}
	// 返回
	return Changes, AUTXO, nil
}

// GetBalanceToLight 跨区转账时使用的查询余额的函数，可以返回带交易坐标的Output集合
func (w Wallet) GetBalanceToLight() (int, map[string][]core.TXOutputsTran, error) {
	// 方法：查询这个公钥地址所有的UTXO的转入转出，并进行合并
	// return core.BlockChain{}.FindUTXO2(PublicKey) // some bugs

	var Changes int = 0 // 钱包余额

	AUTXO := make(map[string][]core.TXOutputsTran, 10) // 当前用户可用的UTXO集合

	// 将Publickey转为Address
	Address := crypto.PubkeyToAddress(w.PublicKey) // 钱包公钥对应的地址
	// 模拟获取目前阶段的blockchain
	err, blockchain := core.GetBlockChain()
	if err != nil {
		fmt.Println("! 模拟获取区块链出现错误")
		return 0, nil, err
	}
	// 寻找目前所有可用的UTXO
	UTXOSet := blockchain.FindUTXOutputsForTran()
	// 遍历所有的可用UTXO，寻找地址一样的未使用交易输出
	for key, value := range UTXOSet {
		for _, out := range value {
			if out.Out.Address == Address {
				fmt.Println("> 找到一个可用的未使用交易输出")
				Changes += out.Out.Value // 累加余额

				// 角色可用的UTXO集合
				AUTXO[key] = append(AUTXO[key], out)
				//AUTXO[key].Outputs = append(AUTXO[key].Outputs, out)
			}
		}
	}
	// 返回
	return Changes, AUTXO, nil
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, _ := crypto.GenerateKey()
	public := private.PublicKey
	wallet := Wallet{*private, public}

	return &wallet
}

// NewWallet2 生成一个已有的钱包
func NewWallet2(pr ecdsa.PrivateKey, pu ecdsa.PublicKey) *Wallet {
	wallet := Wallet{pr, pu}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() common.Address {
	address := crypto.PubkeyToAddress(w.PublicKey)
	return address
}
