package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"transfer/core"

	"github.com/ethereum/go-ethereum/common"
)

const walletFile = "wallet_%s.dat"

// Wallets stores a collection of wallets
// 总钱包包含了自己对应的账号和密码，使用总钱包的账号和密码解锁后就可以获得总钱包对应的所有子钱包的公私钥对。总钱包自己没有地址，也无法提供对应的转账功能
// 之后所有的用户都使用多钱包，如果只有一个子钱包地址，则多钱包结构体内只有一个map对应的key-value对
type Wallets struct {
	Account    string                     // 账户
	Password   string                     // 密码
	Publickey  ecdsa.PublicKey            // 公钥
	Privatekey ecdsa.PrivateKey           // 私钥
	Wallets    map[common.Address]*Wallet // 子钱包地址集合
}

// AccountData2 模拟服务器存储多钱包账户数据
var AccountData2 []Wallets

// NewWallets 新建多钱包
// Account是用户想要创建的账户，Password是用户想要创建账户的密码 TODO:暂时不增加格式检查
func NewWallets(Account string, Password string) *Wallets {
	wallets := Wallets{
		Account:  Account,
		Password: Password,
		Wallets:  nil,
	}
	wallets.Wallets = make(map[common.Address]*Wallet)

	return &wallets
}

// GetAccountAddress 根据账号返回当前账户下所有的子钱包地址
func (ws *Wallets) GetAccountAddress(account string) (IsFind bool, Alladdress []common.Address) {
	// 获取当前的账户数据 TODO:模拟服务器获取账户数据
	for _, w := range AccountData2 {
		if account != w.Account {
			continue
		} else {
			for a, _ := range w.Wallets {
				Alladdress = append(Alladdress, a) // 返回值中添加符合条件的地址
			}
			return true, Alladdress
		}
	}
	// 代码运行到这一步说明没有匹配到输入的目标账户
	return false, nil
}

// GetAccountWallets 根据账户返回该账户对应的钱包
func GetAccountWallets(account string) (IsFind bool, W Wallets) {
	// 获取当前的账户数据 TODO:模拟服务器获取账户数据
	for _, w := range AccountData2 {
		if account != w.Account {
			continue
		} else {
			// 找到目标账户对应的钱包
			return true, w
		}
	}
	// 代码运行到这一步说明没有匹配到输入的目标账户
	return false, Wallets{}
}

// CreateWallet adds a Wallet to Wallets
// amount表示需要创建几个子钱包
func (ws *Wallets) CreateWallet(amount int) (a []common.Address) {
	for i := 1; i <= amount; i++ {
		wallet := NewWallet()
		address := wallet.GetAddress()
		a = append(a, address)
		ws.Wallets[address] = wallet
	}
	return a
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []common.Address {
	var addresses []common.Address

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address common.Address) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// LoginWallets 多钱包登录函数
// 返回值中的string是Account
func LoginWallets() (bool, string, *Wallets, error) {
	var account string
	var password string
	// 1 输入钱包账号和密码
	fmt.Println("> 请输入您的钱包账号：")
	num1, err := fmt.Scanln(&account)
	if num1 != 1 || err != nil {
		return false, "", nil, err
	} else {
		fmt.Println("> 您输入的钱包账号是：", account)
	}
	fmt.Println("> 请输入您的钱包密码：")
	num2, err := fmt.Scanln(&password)
	if num2 != 1 || err != nil {
		return false, "", nil, err
	} else {
		fmt.Println("> 您输入的钱包密码是：", password)
	}

	// 登录验证
	verify, w := WalletsVerify(account, password)
	if verify == false {
		fmt.Println("登录验证失败")
		return false, "", nil, err
	}
	// 返回
	return true, account, w, nil
}

// WalletsVerify 登录验证
// 查询账号密码，返回值第一个值为true表示成功，第一个值为false表示失败
func WalletsVerify(a, p string) (bool, *Wallets) {
	// TODO: 这里应该接入服务器查询账号密码是否正确，为了模拟就不用了
	for _, v := range AccountData2 {
		if a == v.Account {
			if p == v.Password {
				fmt.Println("账号密码验证通过")
				// 验证通过后需要接入服务器查询对应的公私钥
				return true, &v
			}
		} else {
			continue
		}
	}
	fmt.Println("账号密码验证失败")
	return false, &Wallets{}
}

// WalletsBalance 正常交易
type WalletsBalance struct {
	Address common.Address            // 钱包地址
	Balance int                       // 余额
	Txouts  map[string]core.TXOutputs // 每个钱包可用的UTXO集合
}

// WalletsBalance2 正常交易
type WalletsBalance2 struct {
	Address common.Address             // 钱包地址
	Balance int                        // 余额
	Txouts  map[string]core.TXOutputs2 // 每个钱包可用的UTXO集合
}

// WalletsBalanceToLight 跨链交易ToLight
type WalletsBalanceToLight struct {
	Address common.Address                  // 钱包地址
	Balance int                             // 余额
	Txouts  map[string][]core.TXOutputsTran // 每个钱包可用的UTXO集合
}

// GetWalletsBalance 正常交易获取余额余额
func (ws Wallets) GetWalletsBalance() ([]WalletsBalance, error) {
	// 新建返回值
	wb := make([]WalletsBalance, 1)
	// 分开调用GetBalance方法
	for _, v := range ws.Wallets {
		balance, txouts, err := v.GetBalance()
		if err != nil {
			fmt.Println("! 获取多钱包余额时出现错误")
			return nil, err
		}
		wb = append(wb, WalletsBalance{
			Address: v.GetAddress(),
			Balance: balance,
			Txouts:  txouts,
		})
	}
	// 返回
	return wb, nil
}

// GetWalletsBalance2 正常交易获取余额余额
func (ws Wallets) GetWalletsBalance2() ([]WalletsBalance2, error) {
	// 新建返回值
	wb := make([]WalletsBalance2, 1)
	// 分开调用GetBalance方法
	for _, v := range ws.Wallets {
		balance, txouts, err := v.GetBalance2()
		if err != nil {
			fmt.Println("! 获取多钱包余额时出现错误")
			return nil, err
		}
		wb = append(wb, WalletsBalance2{
			Address: v.GetAddress(),
			Balance: balance,
			Txouts:  txouts,
		})
	}
	// 返回
	return wb, nil
}

// GetWalletsBalanceToLight 跨链交易ToLight获取余额余额
func (ws Wallets) GetWalletsBalanceToLight() ([]WalletsBalanceToLight, error) {
	// 新建返回值
	wb := make([]WalletsBalanceToLight, 1)
	// 分开调用GetBalance方法
	for _, v := range ws.Wallets {
		balance, txouts, err := v.GetBalanceToLight()
		if err != nil {
			fmt.Println("! 获取多钱包余额时出现错误")
			return nil, err
		}
		wb = append(wb, WalletsBalanceToLight{
			Address: v.GetAddress(),
			Balance: balance,
			Txouts:  txouts,
		})
	}
	// 返回
	return wb, nil
}

//func mergeMaps(a, b map[string]core.TXOutputs) map[string]core.TXOutputs {
//	result := make(map[string]core.TXOutputs)
//
//	// 将变量 a 的值复制到结果中
//	for key, value := range a {
//		result[key] = value
//	}
//
//	// 将变量 b 的值合并到结果中
//	for key, value := range b {
//		if _, ok := result[key]; ok {
//			// 如果结果中已存在该键，则将 value.Outputs 合并到结果中对应的值中
//			result[key].Outputs = append(result[key].Outputs, value.Outputs...)
//		} else {
//			// 如果结果中不存在该键，则直接赋值
//			result[key] = value
//		}
//	}
//
//	return result
//}

// GetPublickey 根据账户账号获取账户对应的公钥
func GetPublickey(Account string) (ecdsa.PublicKey, bool) {
	// 模拟从服务器获取账户数据
	for _, w := range AccountData2 {
		if w.Account == Account {
			return w.Publickey, true
		}
	}
	return ecdsa.PublicKey{}, false
}
