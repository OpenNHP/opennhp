package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"net/http"
	"os"

	"github.com/OpenNHP/opennhp/kgc"
	"github.com/emmansun/gmsm/sm2" // 替换为实际路径
)

// 固定种子初始化随机数生成器
var seededRand1 = mathRand.New(mathRand.NewSource(12345)) // 使用固定的种子

// 生成用户随机数 dA_
func GetRandomdA_(n *big.Int) (*big.Int, error) {
	dA_ := new(big.Int)
	nMinusOne := new(big.Int).Sub(n, big.NewInt(1))
	for {
		dA_.SetInt64(seededRand1.Int63n(nMinusOne.Int64()-1) + 1)

		if dA_.Cmp(big.NewInt(1)) >= 0 && dA_.Cmp(n) < 0 {
			break
		}
	}
	return dA_, nil
}

// CalculateUA 计算 UA = [dA_]G
func CalculateUA() (*big.Int, *big.Int, *big.Int, error) {
	n := sm2.P256().Params().N
	dA_, err := GetRandomdA_(n)
	if err != nil {
		return nil, nil, nil, err
	}

	curve := sm2.P256()
	UAx, UAy := curve.ScalarBaseMult(dA_.Bytes())

	return dA_, UAx, UAy, nil
}

// 计算 dA = (tA + dA_) mod n
func CalculateDA(n *big.Int, ha []byte, waX, waY *big.Int) (*big.Int, error) {
	tA, err := kgc.CalculateTA(n, ha, waX, waY)
	if err != nil {
		return nil, err
	}

	dA_, err := GetRandomdA_(n)
	if err != nil {
		return nil, err
	}

	dA := new(big.Int).Add(tA, dA_)
	dA.Mod(dA, n)

	fmt.Printf("计算的 dA: %s\n", dA.String())
	return dA, nil
}

// CalculatePA 计算 PA = WA + [l]Ppub
func CalculatePA() (*big.Int, *big.Int, error) {
	_, Ppub, err := kgc.GenerateMasterKeyPair()
	if err != nil {
		return nil, nil, err
	}

	waX, waY, err := kgc.CalculateWA()
	if err != nil {
		return nil, nil, err
	}

	l := kgc.CalculateL(waX, waY, nil)

	curve := sm2.P256()
	PAx, PAy := curve.ScalarBaseMult(l.Bytes())
	PAx, PAy = curve.Add(PAx, PAy, Ppub.X, Ppub.Y)

	return PAx, PAy, nil
}

// 计算 PA_ = [dA]G
func CalculatePA_() (*big.Int, *big.Int, error) {
	curve := sm2.P256()
	n := curve.Params().N

	var ha []byte
	waX, waY, err := kgc.CalculateWA()
	if err != nil {
		return nil, nil, err
	}

	dA, err := CalculateDA(n, ha, waX, waY)
	if err != nil {
		return nil, nil, err
	}

	PA_X, PA_Y := curve.ScalarBaseMult(dA.Bytes())

	return PA_X, PA_Y, nil
}

// 发送 HTTP 请求
func SendHTTPRequest(email string) {
	params := kgc.HAParams{
		UserID:  email,
		EntlenA: len(email),
		Curve:   sm2.P256(),
	}

	// 设置公钥的 x 和 y 坐标（示例）
	// 实际使用中需要替换为计算得到的公钥坐标
	params.XPub = big.NewInt(0) // 设置为实际值
	params.YPub = big.NewInt(0) // 设置为实际值

	jsonData, err := json.Marshal(params)
	if err != nil {
		fmt.Printf("JSON 编码失败: %v\n", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/calculate_ha", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("HTTP 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求失败，状态码: %d\n", resp.StatusCode)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("响应解析失败: %v\n", err)
		return
	}

	fmt.Printf("计算的 HA: %s\n", response["ha"])
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入邮箱: ")
	email, _ := reader.ReadString('\n')

	// 发送 HTTP 请求
	SendHTTPRequest(email)
}
