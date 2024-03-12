package interconnected

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"transfer/core"
	"transfer/wallet"
)

// 轻计算区 -> 转账区
// 先进行账户转换，再执行具体交易

// 传入 目标地址AAd dress，转账金额AMoney 生成 Type类型为2的UTXO，监测目标账户是否存在并创建
// 返回 生成的转账交易，操作结果，执行错误

// ToTransferAccount 用于跨区转账ToTransfer向轻计算区返回新建账户信息的结构体，注意是新建账户才使用这个结构体
type ToTransferAccount struct {
	Account    string
	Password   string
	Publickey  ecdsa.PublicKey
	Privatekey ecdsa.PrivateKey
}

// ToTransfer 跨区交易ToTransfer轻计算区向转账区转钱
// 第一个返回值bool表示是否成功转账，默认向Publickey对应的地址转账，没有就创建新的地址，并返回ToTransferAccount信息 TODO: 但是目前无法根据已知的公钥创建对应的私钥
// 不需要私钥信息，按照Coinbase交易构造 TODO:现在互相信任，认为对方转账一定能成功，所以去掉第一个bool值
func ToTransfer(FromAddress common.Address, Publickey ecdsa.PublicKey, Money int) (ToTransferAccount, core.Transaction) {
	fmt.Println("> 开始执行轻计算区 转 转账区 账户转换")

	// TODO:存在一点问题，输入的应该是账号和密码，系统显示子地址让用户选择，现在输入的是多钱包的公私钥但是多钱包地址无法转账
	// 因为存在旧帐户和新账户两种情况，所以需要统一收集信息
	var BAddress common.Address // 目标地址
	var Pubkey ecdsa.PublicKey  // 账户公钥
	var UserAccount string      // 发送者账户
	var IsNew bool = true       // 是否需要新建账户
	var reAccount ToTransferAccount

	// 使用已有账户，检查目标公钥是否存在
	//BAddress := crypto.PubkeyToAddress(Publickey)
	// 更改为多钱包账户
	for _, v := range wallet.AccountData2 {
		// k是Account，v是密码，公钥，私钥的结构体
		if v.Publickey == Publickey {
			// 公钥正确，说明存在对应地址
			fmt.Println("> 目标地址存在")
			UserAccount = v.Account
			IsNew = false
			Pubkey = Publickey
		}
	}
	if IsNew == true {
		fmt.Println("> 目标地址不存在")
		// 将目标地址设置为新建钱包的公钥 TODO: 暂时无法实现，创建新账户
		// 创建新的账户与地址 TODO: 后续可以设置为用户自定义账户信息
		a, _ := crypto.GenerateKey()
		wallet.AccountData["TestA"] = wallet.AccountECDSA{Password: "TestA", Publickey: a.PublicKey, PrivateKey: *a}
		UserAccount = "TestA"
		Pubkey = a.PublicKey

		// 构造返回账户信息
		reAccount = ToTransferAccount{
			Account:    "TestA",
			Password:   "TestA",
			Publickey:  Pubkey,
			Privatekey: *a,
		}
	}

	BAddress = crypto.PubkeyToAddress(Pubkey)

	// 构造新的UTXO Coinbase交易
	TX := core.NewCoinbaseTX(FromAddress, BAddress, Money, UserAccount)

	// 返回 reAccount如果是空的，就代表没有新建账户
	return reAccount, *TX
}

// TODO:根据坐标返回交易

// TODO:空白交易占位符？

// TODO:变色龙哈希

// TODO:互联 2 phase commit
