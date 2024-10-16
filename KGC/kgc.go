package KGC

import (
	"crypto/rand"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
	"golang.org/x/crypto/sha3"
)

// KGC 结构体，包含系统主私钥和主公钥
type KGC struct {
	MasterPrivateKey *big.Int       // 系统主私钥 s
	MasterPublicKey  *sm2.PublicKey // 系统主公钥 P_pub
}

// NewKGC 初始化一个新的 KGC 实例
func NewKGC() (*KGC, error) {
	kgc := &KGC{}

	// 生成系统主私钥 s
	var err error
	n := sm2.P256Sm2().Params().N // 获取曲线的阶 N
	kgc.MasterPrivateKey, err = rand.Int(rand.Reader, n)
	if err != nil {
		return nil, err
	}

	// 生成系统主公钥 P_pub = s * G
	kgc.MasterPublicKey = new(sm2.PublicKey)
	kgc.MasterPublicKey.Curve = sm2.P256Sm2()
	kgc.MasterPublicKey.X, kgc.MasterPublicKey.Y = kgc.MasterPublicKey.Curve.ScalarBaseMult(kgc.MasterPrivateKey.Bytes())

	return kgc, nil
}

// GeneratePartialPrivateKey 生成用户的部分私钥 D_u
func (kgc *KGC) GeneratePartialPrivateKey(ID string) (*big.Int, error) {
	// 计算 h_u = H(ID)
	h_u := hashID(ID)

	// 计算 D_u = s + h_u mod n
	n := kgc.MasterPublicKey.Curve.Params().N
	D_u := new(big.Int).Add(kgc.MasterPrivateKey, h_u)
	D_u.Mod(D_u, n)

	return D_u, nil
}

// GeneratePartialPublicKey 计算用户的部分公钥 P_u
func (kgc *KGC) GeneratePartialPublicKey(ID string) (*sm2.PublicKey, error) {
	// 计算 h_u = H(ID)
	h_u := hashID(ID)

	// 计算 P_u = h_u * G
	curve := kgc.MasterPublicKey.Curve
	x, y := curve.ScalarBaseMult(h_u.Bytes())

	P_u := &sm2.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	return P_u, nil
}

// hashID 对用户身份 ID 进行哈希，返回一个大整数
func hashID(ID string) *big.Int {
	hash := sha3.New256()
	hash.Write([]byte(ID))
	digest := hash.Sum(nil)
	h_u := new(big.Int).SetBytes(digest)
	n := sm2.P256Sm2().Params().N
	h_u.Mod(h_u, n)
	return h_u
}
