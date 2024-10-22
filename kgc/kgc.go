package kgc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"net/http"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
)

// 定义椭圆曲线参数的固定值
var fixedA, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)
var fixedB, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFD", 16)

// HAParams 结构体用于存储计算 HA 所需的参数
type HAParams struct {
	UserID  string         `json:"user_id"`  // 用户 ID
	EntlenA int            `json:"entlen_a"` // entlenA 的长度
	Curve   elliptic.Curve `json:"-"`        // 椭圆曲线
	XPub    *big.Int       `json:"x_pub"`    // 公钥的 x 坐标
	YPub    *big.Int       `json:"y_pub"`    // 公钥的 y 坐标
	UAx     *big.Int       `json:"u_ax"`     // UA x 坐标
	UAy     *big.Int       `json:"u_ay"`     // UA y 坐标
}

// 生成 KGC 的主密钥对
func GenerateMasterKeyPair() (*sm2.PrivateKey, *ecdsa.PublicKey, error) {
	ms, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("生成主密钥失败: %v", err)
	}
	curve := sm2.P256()
	PpubX, PpubY := curve.ScalarBaseMult(ms.D.Bytes())
	Ppub := &ecdsa.PublicKey{Curve: curve, X: PpubX, Y: PpubY}
	return ms, Ppub, nil
}

// 计算 HA 值
func CalculateHA(params HAParams) []byte {
	entla := make([]byte, 2)
	binary.BigEndian.PutUint16(entla, uint16(params.EntlenA))
	ida := []byte(params.UserID)
	xG := params.Curve.Params().Gx
	yG := params.Curve.Params().Gy
	a := fixedA.Bytes()
	b := fixedB.Bytes()

	data := bytes.Join([][]byte{
		entla, ida, a, b,
		xG.Bytes(), yG.Bytes(),
		params.XPub.Bytes(), params.YPub.Bytes(),
	}, nil)

	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// 固定种子初始化随机数生成器
var seededRand2 = mathRand.New(mathRand.NewSource(67890))

// 生成 kgc 随机数 W
func GetRandomW(n *big.Int) (*big.Int, error) {
	W := new(big.Int)
	nMinusOne := new(big.Int).Sub(n, big.NewInt(1))
	for {
		W.SetInt64(seededRand2.Int63n(nMinusOne.Int64()-1) + 1)

		if W.Cmp(big.NewInt(1)) >= 0 && W.Cmp(n) < 0 {
			break
		}
	}
	return W, nil
}

// CalculateWA 计算 WA = [W]G + UA
func CalculateWA(uaX, uaY *big.Int) (*big.Int, *big.Int, error) {
	n := sm2.P256().Params().N
	W, err := GetRandomW(n)
	if err != nil {
		return nil, nil, err
	}

	curve := sm2.P256()
	WAx, WAy := curve.ScalarBaseMult(W.Bytes())
	WAx, WAy = curve.Add(WAx, WAy, uaX, uaY)

	return WAx, WAy, nil
}

// 计算 L 值
func CalculateL(waX, waY *big.Int, ha []byte) *big.Int {
	waCoords := append(waX.Bytes(), waY.Bytes()...)
	dataForL := append(waCoords, ha...)

	hashL := sm3.New()
	hashL.Write(dataForL)
	hashValue := hashL.Sum(nil)

	l := new(big.Int).SetBytes(hashValue)
	l.Mod(l, sm2.P256().Params().N)

	return l
}

// 计算 tA
func CalculateTA(n *big.Int, ha []byte, waX, waY *big.Int) (*big.Int, error) {
	w, err := GetRandomW(n)
	if err != nil {
		return nil, err
	}

	ms, _, err := GenerateMasterKeyPair()
	if err != nil {
		return nil, err
	}

	l := CalculateL(waX, waY, ha)

	lms := new(big.Int).Mul(l, ms.D)
	tA := new(big.Int).Add(w, lms)
	tA.Mod(tA, n)

	return tA, nil
}

// HTTP 处理函数
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var params HAParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	params.Curve = sm2.P256() // 设置曲线

	ha := CalculateHA(params)

	// 计算 WA
	waX, waY, err := CalculateWA(params.UAx, params.UAy)
	if err != nil {
		http.Error(w, "计算 WA 失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ha": fmt.Sprintf("%x", ha),
		"wa": fmt.Sprintf("%x,%x", waX, waY),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 启动 HTTP 服务器
func StartServer() {
	http.HandleFunc("/calculate_ha", HandleRequest)
	http.ListenAndServe(":8080", nil)
}
