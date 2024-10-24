package kgc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
)

// 固定参数A和B
var fixedA, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)
var fixedB, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFD", 16)

// HAParams结构体存储用户信息和曲线参数
type HAParams struct {
	userID  string         // 用户ID
	entlenA int            // A的长度
	curve   elliptic.Curve // 椭圆曲线
	xPub    *big.Int       // 公钥X坐标
	yPub    *big.Int       // 公钥Y坐标
}

// 生成主密钥对
func GenerateMasterKeyPair() (*sm2.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := sm2.GenerateKey(rand.Reader) // 生成私钥
	if err != nil {
		return nil, nil, fmt.Errorf("生成主密钥失败: %v", err)
	}

	// 转换为ECDSA公钥
	ecdsaPublicKey := &ecdsa.PublicKey{
		Curve: sm2.P256(),
		X:     privateKey.PublicKey.X,
		Y:     privateKey.PublicKey.Y,
	}

	return privateKey, ecdsaPublicKey, nil
}

// 计算HA
func CalculateHA(params HAParams) []byte {
	entla := make([]byte, 2)                                  // 创建长度为2的字节切片
	binary.BigEndian.PutUint16(entla, uint16(params.entlenA)) // 存储A的长度
	ida := []byte(params.userID)                              // 转换用户ID为字节
	xG := params.curve.Params().Gx                            // 获取基点X坐标
	yG := params.curve.Params().Gy                            // 获取基点Y坐标
	a := fixedA.Bytes()                                       // 固定参数A的字节表示
	b := fixedB.Bytes()                                       // 固定参数B的字节表示

	// 拼接数据
	data := bytes.Join([][]byte{
		entla, ida, a, b,
		xG.Bytes(), yG.Bytes(),
		params.xPub.Bytes(), params.yPub.Bytes(),
	}, nil)

	// 打印输入数据
	fmt.Printf("HA输入数据:\n")
	fmt.Printf("  entla: %x\n", entla)              // 打印A的长度
	fmt.Printf("  ida: %x\n", ida)                  // 打印用户ID
	fmt.Printf("  a: %x\n", a)                      // 打印固定参数A
	fmt.Printf("  b: %x\n", b)                      // 打印固定参数B
	fmt.Printf("  xG: %x\n", xG.Bytes())            // 打印基点X坐标
	fmt.Printf("  yG: %x\n", yG.Bytes())            // 打印基点Y坐标
	fmt.Printf("  xPub: %x\n", params.xPub.Bytes()) // 打印公钥X坐标
	fmt.Printf("  yPub: %x\n", params.yPub.Bytes()) // 打印公钥Y坐标
	fmt.Printf("  合并数据: %x\n", data)                // 打印拼接后的数据

	hash := sm3.New()    // 创建SM3哈希
	hash.Write(data)     // 写入数据
	return hash.Sum(nil) // 返回哈希值
}

// 检查点是否在曲线上
func isPointOnCurve(curve elliptic.Curve, x, y *big.Int) bool {
	return curve.IsOnCurve(x, y)
}

// 计算L值
func CalculateL(waX, waY *big.Int, ha []byte) *big.Int {
	waCoords := append(waX.Bytes(), waY.Bytes()...) // 拼接WA的坐标
	dataForL := append(waCoords, ha...)             // 拼接计算L的输入数据
	hashL := sm3.New()                              // 创建SM3哈希
	hashL.Write(dataForL)                           // 写入数据
	l := new(big.Int).SetBytes(hashL.Sum(nil))      // 获取哈希结果
	l.Mod(l, sm2.P256().Params().N)                 // 对N取模

	// 打印计算结果
	fmt.Printf("waX: %s, waY: %s, ha: %x\n", waX.String(), waY.String(), ha) // 打印输入数据
	fmt.Printf("计算的l: %s\n", l.String())                                     // 打印计算出的L值
	fmt.Printf("CalculateL中的HA: %x\n", ha)                                   // 打印HA值

	return l
}

// 生成KGC部分密钥
func GenerateKGCPartialKey(userID string, entlenA int, kgcPrivateKey *sm2.PrivateKey, userPublicKey *ecdsa.PublicKey, ua *ecdsa.PublicKey, userEmail string, kgcPublicKey *ecdsa.PublicKey) (*ecdsa.PublicKey, *big.Int) {
	curve := sm2.P256()                          // 使用SM2曲线
	xPub, yPub := kgcPublicKey.X, kgcPublicKey.Y // 获取KGC公钥坐标

	// 设置HA参数
	params := HAParams{
		userID:  userID,
		entlenA: entlenA,
		curve:   curve,
		xPub:    xPub,
		yPub:    yPub,
	}
	ha := CalculateHA(params) // 计算HA

	w, err := sm2.GenerateKey(rand.Reader) // 生成随机密钥w
	if err != nil {
		fmt.Printf("生成KGC部分密钥失败: %v\n", err)
		return nil, nil
	}

	// 打印生成的随机密钥
	fmt.Printf("生成的随机密钥w: d: %s, 公钥: (%s, %s)\n", w.D.String(), w.PublicKey.X.String(), w.PublicKey.Y.String()) // 打印随机密钥信息

	// 计算WA
	waX, waY := curve.Add(userPublicKey.X, userPublicKey.Y, w.PublicKey.X, w.PublicKey.Y)
	if !isPointOnCurve(curve, waX, waY) { // 检查WA是否在曲线上
		fmt.Println("WA不在曲线上！")
		return nil, nil
	}

	l := CalculateL(waX, waY, ha)                                                                         // 计算L
	tA := new(big.Int).Mod(new(big.Int).Add(w.D, new(big.Int).Mul(l, kgcPrivateKey.D)), curve.Params().N) // 计算tA

	// 打印生成结果
	fmt.Printf("生成的WA: (%s, %s)\n", waX.String(), waY.String()) // 打印WA坐标
	fmt.Printf("生成的tA: %s\n", tA.String())                      // 打印tA值
	fmt.Printf("用户邮箱: %s\n", userEmail)                         // 打印用户邮箱

	return &ecdsa.PublicKey{Curve: curve, X: waX, Y: waY}, tA
}

// 生成用户密钥
func GenerateUserKey(entlenA int, kgcPublicKey *ecdsa.PublicKey, kgcPrivateKey *sm2.PrivateKey, userEmail string) (*big.Int, *ecdsa.PublicKey, *big.Int) {
	userID := userEmail // 使用用户邮箱作为用户ID

	dA_, err := rand.Int(rand.Reader, kgcPublicKey.Curve.Params().N) // 生成随机d'A
	if err != nil {
		fmt.Printf("生成随机d'A失败: %v\n", err)
		return nil, nil, nil
	}
	fmt.Printf("生成的d'A: %s\n", dA_) // 打印生成的d'A

	curve := kgcPublicKey.Curve                   // 获取曲线
	UAx, UAy := curve.ScalarBaseMult(dA_.Bytes()) // 计算用户公钥

	// 生成KGC部分密钥
	waPublicKey, tA := GenerateKGCPartialKey(userID, entlenA, kgcPrivateKey, kgcPublicKey, &ecdsa.PublicKey{Curve: curve, X: UAx, Y: UAy}, userEmail, kgcPublicKey)

	dA := new(big.Int).Mod(new(big.Int).Add(tA, dA_), curve.Params().N) // 计算dA

	if dA.Sign() == 0 { // 检查dA是否为0
		fmt.Println("dA为0，返回A1")
		return nil, nil, nil
	}

	return dA, waPublicKey, tA // 返回dA、WA公钥和tA
}

// 验证密钥对
func VerifyKeyPair(dA *big.Int, WA *ecdsa.PublicKey, userID string, entlenA int, kgcPublicKey *ecdsa.PublicKey, receivedWAPubKey *ecdsa.PublicKey, receivedTA *big.Int) bool {
	ha := CalculateHA(HAParams{
		userID:  userID,
		entlenA: entlenA,
		curve:   sm2.P256(),
		xPub:    kgcPublicKey.X, // 使用KGC公钥的X坐标
		yPub:    kgcPublicKey.Y, // 使用KGC公钥的Y坐标
	})
	fmt.Printf("在VerifyKeyPair中计算的HA: %x，用户ID: %s\n", ha, userID) // 打印HA和用户ID

	l := CalculateL(WA.X, WA.Y, ha) // 计算L

	PAx, PAy := WA.X, WA.Y                                                          // WA的坐标
	Ppub := &ecdsa.PublicKey{Curve: WA.Curve, X: kgcPublicKey.X, Y: kgcPublicKey.Y} // KGC公钥

	// 计算PA的坐标
	lPpubX, lPpubY := WA.Curve.ScalarMult(Ppub.X, Ppub.Y, l.Bytes())
	PAx, PAy = WA.Curve.Add(PAx, PAy, lPpubX, lPpubY)

	if !WA.Curve.IsOnCurve(PAx, PAy) { // 检查PA是否在曲线上
		fmt.Printf("PA坐标不在曲线上\n")
		return false
	}

	fmt.Printf("PA坐标: (%s, %s)\n", PAx.String(), PAy.String()) // 打印PA坐标

	PAPx, PAPy := WA.Curve.ScalarBaseMult(dA.Bytes())             // 计算P'A坐标
	fmt.Printf("P'A坐标: (%s, %s)\n", PAPx.String(), PAPy.String()) // 打印P'A坐标

	if !WA.Curve.IsOnCurve(PAPx, PAPy) { // 检查P'A是否在曲线上
		fmt.Printf("P'A坐标不在曲线上\n")
		return false
	}

	// 比较PA和P'A坐标
	if PAx.Cmp(PAPx) == 0 && PAy.Cmp(PAPy) == 0 {
		fmt.Println("验证成功") // 验证成功的输出
		return true
	}

	fmt.Println("验证失败") // 验证失败的输出
	return false
}
