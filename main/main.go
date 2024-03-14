package main

import (
	"fmt"
	"strconv"
	"transfer/core"
	"transfer/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// 核心流程文件

// Main 开机启动界面
func main() {
	// 模拟服务器的账户密码数据，在这里我们新建三个用户账户数据
	NewAccount()
	a, _ := crypto.GenerateKey()
	fmt.Println(a)
	fmt.Println("> 欢迎登录盘古区块链转账系统")
	fmt.Println("> 请您输入自己的钱包密码登录钱包")
	result, Account, w, err := wallet.LoginWallets()
	if err != nil {
		fmt.Println(err)
		return
	}
	if result == false {
		fmt.Println("! 登录错误，请您重新登录")
		return
	}

	fmt.Println("> 欢迎您进入磐古区块链系统，您的登录账户为：", Account)

	// 获取钱包余额，AUTXO是当前用户可用的UTXO集合
	UTXOs, err := w.GetWalletsBalance()
	if err != nil {
		fmt.Println("! 获取钱包余额方法出现错误")
		return
	}
	// 计算钱包总金额
	var AllMoney1 int = 0
	// 输出钱包余额
	for _, u := range UTXOs {
		AllMoney1 += u.Balance
		fmt.Printf("> 您的 %v 钱包余额为 %d \n", u.Address, u.Balance)
	}
	fmt.Println("您的钱包总余额为 ", AllMoney1)

	// 转账功能 A to B 一个账户可能有多个地址 后续可以新增联系人列表功能
	fmt.Println("> 即将进行转账功能，请输入您想要转给的 用户账户")
	var BAccouont string // 目标账户
	//var BAddress []common.Address            // 目标地址集合
	var BMoney int                           // 目标金额
	var inputnum string                      // 用户输入的字符串序号
	var IsEnd string                         // 输入转账目标地址循环是否结束
	var BAddressMoney map[common.Address]int // 转账目标地址与目标金额
	var AllUTXOs map[string]core.TXOutputs   // 可用的UTXO集合

	// TODO: 改成输入账户后显示该账户拥有的钱包地址集合，让用户自己选择给哪个地址转账

	// 将Publickey转为Address
	//AAddress := crypto.PubkeyToAddress(PublicKey) // 钱包公钥对应的地址

	// 获取输入相关信息
	for {
		fmt.Println("> 请输入 转账目标账户：")
		num1, err := fmt.Scanln(BAccouont)
		if num1 != 1 || err != nil {
			fmt.Println("! 输入转账目标账户出现错误")
			return
		}
		// TODO: 开始输出所有该账户的地址，让用户选择使用哪些地址进行交易
		IsFind, AllBAddress := w.GetAccountAddress(BAccouont)
		if IsFind == false {
			fmt.Println("! 您输入的账户是无效账户，请您重新输入")
			return
		}
		fmt.Println("您输入的账户包含以下地址，请您输入需要转帐的地址标号，注意一次只能输入一个地址")
		for i, k := range AllBAddress {
			fmt.Printf("%d . %v \n", i+1, k)
		}

		num1, err = fmt.Scanln(inputnum)
		if num1 != 1 || err != nil {
			fmt.Println("! 输入转账目标账户序号出现错误")
			return
		}
		// 添加到目标地址数组
		// 将字符串转换为整数
		num, err := strconv.Atoi(inputnum)
		if err != nil {
			fmt.Println("! string -> int 转换失败")
			return
		}

		fmt.Println("> 请输入 转账目标金额：")
		num3, err := fmt.Scanln(BMoney)
		if num3 != 1 || err != nil {
			fmt.Println("! 输入转账目标金额出现错误")
			return
		}

		if num > len(AllBAddress) || num <= 0 {
			fmt.Println("! 输入地址标号超出范围限制")
			return
		}
		// 目标账户地址与转账金额赋值
		BAddressMoney[AllBAddress[num]] = BMoney

		fmt.Println("是否结束？Y/N")
		num4, err := fmt.Scanln(IsEnd)
		if num4 != 1 || err != nil {
			fmt.Println("! 输入是否结束出现错误")
			return
		}

		if IsEnd == "Y" {
			break
		}
	}

	// 修改为用户选择显示的地址，不需要验证地址是否有效
	//var isTrueAddress bool = false
	//for k, v := range wallet.AccountData {
	//	if k == BAccouont {
	//		// 将Publickey转为Address
	//		vaddress := crypto.PubkeyToAddress(v.Publickey) // 钱包公钥对应的地址
	//		// 目标地址是否存在，目前是单一地址，后续可能需要改成多地址
	//		if BAddress != vaddress {
	//			fmt.Println("! 地址不符合")
	//			return
	//		}
	//		isTrueAddress = true
	//		break
	//	}
	//}
	//if !isTrueAddress {
	//	fmt.Println("! 您输入的目标地址不存在")
	//	return
	//}

	//// 2 查询自己账户余额是否足够（单钱包）
	//if BMoney > Balance {
	//	fmt.Println("! 转账金额大于余额，转账失败")
	//	return
	//}

	// 查询自己账户余额是否足够（多钱包）
	// 计算总金额
	var AllMoney2 int = 0
	for _, v := range BAddressMoney {
		AllMoney2 += v
	}
	if AllMoney2 > AllMoney1 {
		fmt.Println("! 余额不足")
		return
	}
	// 按转账需要金额从高到底排序
	//// 对切片进行排序
	//sort.Slice(BAddressMoney, func(i, j int) bool {
	//	return BAddressMoney[i].Value > BAddressMoney[j].Value
	//})

	// 3 所有检查通过，开始发送交易，新建UTXO
	// 方法内包含了对UTXO out的签名
	// 新建钱包
	w1 := core.NewWallet(Account, w.Privatekey, w.Publickey) // w.Privatekey签名，w.Publickey验证签名

	// 合并可用的UTXO TODO:有点麻烦，看后续有没有其他的解决方法
	for _, v := range UTXOs {
		for k1, v1 := range v.Txouts {
			for _, k3 := range v1.Outputs {
				a := AllUTXOs[k1].Outputs
				a = append(a, k3)
				b := AllUTXOs[k1]
				b.Outputs = a
				AllUTXOs[k1] = b
			}
		}
	}
	TX, err := core.NewTransaction(w1, BAddressMoney, AllUTXOs)
	if err != nil {
		fmt.Println("! 构造新交易出现错误")
		return
	}
	// tx := core.NewUTXOTransaction(w, crypto.PubkeyToAddress(key2), money2, nil)
	// 新交易构造完成
	fmt.Println("> 交易构造完成")

	// 发送交易
	// 模拟发送交易
	Pool = append(Pool, TX)

	fmt.Println("> 交易已发送，欢迎您的使用")
}

// TransactionVerify 节点收到交易后的验证
func TransactionVerify(tx core.Transaction) bool {
	// 其实就是把上面的事情再做一遍，然后把交易添加到区块中
	// 暂时省略验证的步骤，区块添加入块
	return true
}

func NewAccount() {
	// 模拟服务器的账户密码数据，在这里我们新建三个用户账户数据
	a, _ := crypto.GenerateKey()
	b, _ := crypto.GenerateKey()
	c, _ := crypto.GenerateKey()
	wallet.AccountData["aaa"] = wallet.AccountECDSA{Password: "aaa", Publickey: a.PublicKey, PrivateKey: *a}
	wallet.AccountData["bbb"] = wallet.AccountECDSA{Password: "bbb", Publickey: b.PublicKey, PrivateKey: *b}
	wallet.AccountData["ccc"] = wallet.AccountECDSA{Password: "ccc", Publickey: c.PublicKey, PrivateKey: *c}
}

// NewAccount2 模拟多钱包账户数据
func NewAccount2() {
	// 模拟服务器的账户密码数据，在这里我们新建三个用户账户数据
	a := wallet.NewWallets("aaa", "aaa")
	b := wallet.NewWallets("bbb", "bbb")
	c := wallet.NewWallets("ccc", "ccc")
	a.CreateWallet(3)
	b.CreateWallet(2)
	c.CreateWallet(1)
	wallet.AccountData2 = append(wallet.AccountData2, *a)
	wallet.AccountData2 = append(wallet.AccountData2, *b)
	wallet.AccountData2 = append(wallet.AccountData2, *c)
}

// 测试更改

// 测试更改
