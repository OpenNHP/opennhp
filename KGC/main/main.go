package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/OpenNHP/opennhp/KGC" // 导入 KGC 包
)

// KeyResponse 包含返回给 Agent 的部分密钥和椭圆曲线参数
type KeyResponse struct {
	PartialPrivateKey string `json:"partial_private_key"`
	PartialPublicKeyX string `json:"partial_public_key_x"`
	PartialPublicKeyY string `json:"partial_public_key_y"`
	Gx                string `json:"Gx"`
	Gy                string `json:"Gy"`
	N                 string `json:"N"`
}

// 处理 /generateKeys 请求，生成部分密钥
func handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	// 调用 KGC 模块生成部分私钥和部分公钥
	kgc, err := KGC.NewKGC()
	if err != nil {
		http.Error(w, "failed to initialize KGC", http.StatusInternalServerError)
		return
	}

	// 生成部分私钥
	D_u, err := kgc.GeneratePartialPrivateKey(email)
	if err != nil {
		http.Error(w, "failed to generate partial private key", http.StatusInternalServerError)
		return
	}

	// 生成部分公钥
	P_u, err := kgc.GeneratePartialPublicKey(email)
	if err != nil {
		http.Error(w, "failed to generate partial public key", http.StatusInternalServerError)
		return
	}

	// 椭圆曲线的基点 G
	curve := kgc.MasterPublicKey.Curve
	Gx, Gy := curve.Params().Gx, curve.Params().Gy
	N := curve.Params().N

	// 构建 JSON 响应
	response := KeyResponse{
		PartialPrivateKey: D_u.Text(16),
		PartialPublicKeyX: P_u.X.Text(16),
		PartialPublicKeyY: P_u.Y.Text(16),
		Gx:                Gx.Text(16),
		Gy:                Gy.Text(16),
		N:                 N.Text(16),
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 设置 /generateKeys 路由
	http.HandleFunc("/generateKeys", handleKeyRequest)

	// 启动 HTTP 服务
	fmt.Println("KGC HTTP server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 循环100次，每次生成部分私钥和部分公钥，并计算平均时间，单独测试.............................................
/* package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/OpenNHP/opennhp/KGC" // 导入 KGC 包
)

// KeyResponse 包含返回给 Agent 的部分密钥、椭圆曲线参数和时间信息
type KeyResponse struct {
	PartialPrivateKey    string `json:"partial_private_key"`
	PartialPublicKeyX    string `json:"partial_public_key_x"`
	PartialPublicKeyY    string `json:"partial_public_key_y"`
	Gx                   string `json:"Gx"`
	Gy                   string `json:"Gy"`
	N                    string `json:"N"`
	AvgPrivateKeyGenTime string `json:"avg_private_key_gen_time"`
	AvgPublicKeyGenTime  string `json:"avg_public_key_gen_time"`
	AvgTotalTime         string `json:"avg_total_time"`
}

// 处理 /generateKeys 请求，生成部分密钥
func handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	// 循环次数
	numIterations := 20

	// 累计时间
	var totalPrivateKeyGenTime time.Duration
	var totalPublicKeyGenTime time.Duration
	var totalTotalTime time.Duration

	// 调用 KGC 模块生成部分私钥和部分公钥
	kgc, err := KGC.NewKGC()
	if err != nil {
		http.Error(w, "failed to initialize KGC", http.StatusInternalServerError)
		return
	}

	var D_u *big.Int
	var P_uX, P_uY *big.Int

	for i := 0; i < numIterations; i++ {
		// 开始总计时
		totalStartTime := time.Now()

		// 开始计时：生成部分私钥
		privateKeyStartTime := time.Now()
		D_u_temp, err := kgc.GeneratePartialPrivateKey(email)
		if err != nil {
			http.Error(w, "failed to generate partial private key", http.StatusInternalServerError)
			return
		}
		privateKeyGenTime := time.Since(privateKeyStartTime)
		totalPrivateKeyGenTime += privateKeyGenTime

		// 开始计时：生成部分公钥
		publicKeyStartTime := time.Now()
		P_u_temp, err := kgc.GeneratePartialPublicKey(email)
		if err != nil {
			http.Error(w, "failed to generate partial public key", http.StatusInternalServerError)
			return
		}
		publicKeyGenTime := time.Since(publicKeyStartTime)
		totalPublicKeyGenTime += publicKeyGenTime

		// 总时间
		totalTime := time.Since(totalStartTime)
		totalTotalTime += totalTime

		// 保存最后一次生成的密钥
		D_u = D_u_temp
		P_uX = P_u_temp.X
		P_uY = P_u_temp.Y
	}

	// 计算平均时间
	avgPrivateKeyGenTime := totalPrivateKeyGenTime / time.Duration(numIterations)
	avgPublicKeyGenTime := totalPublicKeyGenTime / time.Duration(numIterations)
	avgTotalTime := totalTotalTime / time.Duration(numIterations)

	// 椭圆曲线的基点 G
	curve := kgc.MasterPublicKey.Curve
	Gx, Gy := curve.Params().Gx, curve.Params().Gy
	N := curve.Params().N

	// 构建 JSON 响应
	response := KeyResponse{
		PartialPrivateKey:    D_u.Text(16),
		PartialPublicKeyX:    P_uX.Text(16),
		PartialPublicKeyY:    P_uY.Text(16),
		Gx:                   Gx.Text(16),
		Gy:                   Gy.Text(16),
		N:                    N.Text(16),
		AvgPrivateKeyGenTime: avgPrivateKeyGenTime.String(),
		AvgPublicKeyGenTime:  avgPublicKeyGenTime.String(),
		AvgTotalTime:         avgTotalTime.String(),
	}

	// 打印平均时间信息到控制台
	log.Printf("平均生成部分私钥时间: %s", avgPrivateKeyGenTime)
	log.Printf("平均生成部分公钥时间: %s", avgPublicKeyGenTime)
	log.Printf("平均总时间: %s", avgTotalTime)

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 设置 /generateKeys 路由
	http.HandleFunc("/generateKeys", handleKeyRequest)

	// 启动 HTTP 服务
	fmt.Println("KGC HTTP server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
} */

// 一次测试，生成部分私钥和部分公钥，并计算时间，配合agent进行测试................................................
/* package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OpenNHP/opennhp/KGC" // 请根据您的项目结构调整导入路径
)

// KeyResponse 包含返回给 Agent 的部分密钥和椭圆曲线参数
type KeyResponse struct {
	PartialPrivateKey string `json:"partial_private_key"`
	PartialPublicKeyX string `json:"partial_public_key_x"`
	PartialPublicKeyY string `json:"partial_public_key_y"`
	Gx                string `json:"Gx"`
	Gy                string `json:"Gy"`
	N                 string `json:"N"`
}

// 处理 /generateKeys 请求，生成部分密钥
func handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	// 开始计时
	startTime := time.Now()

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	// 初始化 KGC 模块
	kgc, err := KGC.NewKGC()
	if err != nil {
		http.Error(w, "failed to initialize KGC", http.StatusInternalServerError)
		return
	}

	// 生成部分私钥
	D_u, err := kgc.GeneratePartialPrivateKey(email)
	if err != nil {
		http.Error(w, "failed to generate partial private key", http.StatusInternalServerError)
		return
	}

	// 生成部分公钥
	P_u, err := kgc.GeneratePartialPublicKey(email)
	if err != nil {
		http.Error(w, "failed to generate partial public key", http.StatusInternalServerError)
		return
	}

	// 获取椭圆曲线的基点 G 和阶 N
	curve := kgc.MasterPublicKey.Curve
	Gx, Gy := curve.Params().Gx, curve.Params().Gy
	N := curve.Params().N

	// 构建 JSON 响应
	response := KeyResponse{
		PartialPrivateKey: D_u.Text(16),
		PartialPublicKeyX: P_u.X.Text(16),
		PartialPublicKeyY: P_u.Y.Text(16),
		Gx:                Gx.Text(16),
		Gy:                Gy.Text(16),
		N:                 N.Text(16),
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	// 计算并记录处理时间
	elapsedTime := time.Since(startTime)
	log.Printf("处理请求（邮箱：%s）耗时: %s", email, elapsedTime)
}

func main() {
	// 设置 /generateKeys 路由
	http.HandleFunc("/generateKeys", handleKeyRequest)

	// 启动 HTTP 服务
	fmt.Println("KGC HTTP server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
} */
