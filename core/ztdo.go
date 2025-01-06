package core

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	_ "github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/google/uuid"
)

// Ztdo 完整的数据结构定义
type Ztdo struct {
	Header   ZtdoHeader
	Payload  ZtdoPayload
	Signature ZtdoSignature
}

// ZtdoHeader 表示数据对象的头部结构
type ZtdoHeader struct {
	Typeid         [4]byte  // ZTDO的类型标识符
	Objectid       string // ZTDO数据对象标识
	Version        [2]byte
	PayloadOffset  [2]byte
	PayloadLength  [8]byte
	SignatureLength [2]byte
	ECC_Bind_Mode  [1]byte
	SigConfig      [1]byte
	Reserved1      []byte // KAS
	Reserved2      []byte // Policy
	MetaInfo       []byte // 元信息
}

// ZtdoPayload 表示数据对象的负载结构
type ZtdoPayload struct {
	IV       	[8]byte
	Ciphertxt []byte
	MAC      	[]byte
}

// ZtdoSignature 表示数据对象的签名结构
type ZtdoSignature struct {
	SignerId          []byte
	DomainId          []byte
	DeclaredPublicKey []byte
	Sig               []byte
}

// 定义椭圆曲线参数的常量
const (
	SecP256R1 = 0x00
	SecP384R1 = 0x01
	SecP521R1 = 0x02
	SecP256K1 = 0x03
	SM2       = 0x04
)

const (
	UseGMAC      = 0  // GMAC
	UseECDSA     = 1  // ECDSA 数字签名
	UseSM2      = 1 // SM2 数字签名
)

// 定义对称算法模式常量
const (
	AES256GCM64BitTag  = 0x00 // 0: AES-256-GCM + 64bit Tag
	AES256GCM96BitTag  = 0x01 // 1: AES-256-GCM + 96bit Tag
	AES256GCM104BitTag = 0x02 // 2: AES-256-GCM + 104bit Tag
	AES256GCM112BitTag = 0x03 // 3: AES-256-GCM + 112bit Tag
	AES256GCM120BitTag = 0x04 // 4: AES-256-GCM + 120bit Tag
	AES256GCM128BitTag = 0x05 // 5: AES-256-GCM + 128bit Tag
	SM4GCM64BitTag     = 0x06 // 6: SM4-GCM + 64bit Tag (扩展)
	SM4GCM128BitTag    = 0x07 // 7: SM4-GCM + 128bit Tag (扩展)
)



//--------------填充实际结构体字段方法函数--------------------------------------------------------
// 用于返回 Objectid 的结构体
type ObjectIdResult struct {
	GUID      string // GUID格式的字符串
	HexString string // 十六进制字符串
}

// 计算Objectid，uuid生成16字节随机数作为Objectid，并返回GUID格式的字符串和十六进制字符串
func GenerateObjectId() (ObjectIdResult, error) {
	// 生成一个新的 UUID
	objectId := uuid.New()

	// 将 UUID 转换为十六进制字符串
	hexString := hex.EncodeToString(objectId[:]) // 正确生成32字符的十六进制字符串

	// 将 UUID 转换为字符串（带有"-"的格式，实际上即是 GUID）
	guid := objectId.String() 

	return ObjectIdResult{
		GUID:      guid,      // 这是标准的 GUID 格式
		HexString: hexString,  // 这是一个有效的 32 字符十六进制字符串
	}, nil
}

// 计算 Typeid 的函数
func CalculateTypeid(objectId [16]byte) ([4]byte, error) {
	// 初始化一个 SM3 哈希对象
	h := sm3.New()
	// 写入数据
	_, err := h.Write(objectId[:])
	if err != nil {
		return [4]byte{}, fmt.Errorf("failed to write object ID to SM3: %v", err)
	}
	// 计算哈希值
	hash := h.Sum(nil)
	
	// 提取高位的4字节
	var typeId [4]byte
	copy(typeId[:], hash[:4]) // 从 SM3 哈希值中提取前 4 字节
	return typeId, nil
}

// ConvertHexStringToByteArray 将十六进制字符串转换为 [16]byte
func ConvertHexStringToByteArray(hexString string) ([16]byte, error) {
	if len(hexString) != 32 {
			return [16]byte{}, errors.New("hex string must be 32 characters long")
	}
	var byteArray [16]byte
	decoded, err := hex.DecodeString(hexString)
	if err != nil {
			return byteArray, err
	}
	copy(byteArray[:], decoded)
	return byteArray, nil
}

// 设置 ECC_Bind_Mode 的方法
func SetECCBindMode(curveParam byte, useSignature bool) [1]byte {
	var bindMode byte

	// 设置低三位为椭圆曲线参数
	bindMode = curveParam & 0x07 // 保留高五位，低三位为curveParam

	// 设置最高位为策略绑定模式
	if useSignature {
		bindMode |= 0x80 // 最高位设为1
	} else {
		bindMode &= 0x7F // 最高位设为0
	}

	return [1]byte{bindMode}
}

// 创建 SigConfig 的函数
func SetSigConfig(symAlgMode byte, curveParam byte, hasSignature bool) [1]byte {
	var sigConfig byte

	// 设置低四位为对称算法模式
	sigConfig |= symAlgMode & 0x0F // 较低 4 位

	// 设置中间 3 位为椭圆曲线参数
	sigConfig |= (curveParam & 0x07) << 4 // 中间 3 位（左移 4 位）

	// 设置最高位是否有签名
	if hasSignature {
		sigConfig |= 0x80 // 最高位
	}

	return [1]byte{sigConfig}
}


//-----------------------------------------------------------------------------


// 解析传入的文本文件
func ParseZtdoFromFile(filename string, key []byte) (Ztdo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Ztdo{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var payloadData []byte

	if scanner.Scan() {
		payloadData = []byte(scanner.Text()) // 读取到的负载，按行读取
	}

	if err := scanner.Err(); err != nil {
		return Ztdo{}, fmt.Errorf("error reading file: %v", err)
	}
	
//-----调用填充字段函数---------------------------------------
	// 生成 Objectid
	objectidResult, err := GenerateObjectId()
	if err != nil {
		return Ztdo{}, err
	}
	hexString := objectidResult.HexString
	objectId, err := ConvertHexStringToByteArray(hexString)
	if err != nil {
    	return Ztdo{}, err
	}


	// 计算 Typeid
	typeid, err := CalculateTypeid(objectId)
	if err != nil {
    return Ztdo{}, err
	}
	

	// 设置 SM2 参数并选择数字签名模式
	curveParam := byte(SM2) // 强制转换为 byte 类型
	useSignature := true     // 选择数字签名
	// 获取 ECC_Bind_Mode 值
	eccBindMode := SetECCBindMode(curveParam, useSignature)

	// 示例赋值：选择 AES-256-GCM + 128bit Tag，SM2 作为椭圆曲线参数，并假设有签名
	symAlgMode := byte(AES256GCM128BitTag) // 对称加密算法，确保转换为 byte
	hasSignature := true                     // 指示有签名

	// 获取 SigConfig 值
	sigConfig := SetSigConfig(symAlgMode, curveParam, hasSignature)


//-------------------------------------------------

	// 使用 Encrypt 函数对 payloadData 进行加密
	ciphertxt, err := Encrypt(payloadData, key)
	if err != nil {
		return Ztdo{}, fmt.Errorf("failed to encrypt data: %v", err)
	}

	// 创建 Ztdo 的实例，并将加密后的数据传入 Ciphertxt
	ztdo := Ztdo{
		Header: ZtdoHeader{
			Typeid:       		typeid, 
			Objectid:     		objectidResult.HexString,
			Version:      		[2]byte{1.0},
			PayloadOffset: 		[2]byte{}, // 应该设定为合适的值
			PayloadLength: 		[8]byte{},  // 应该设定为加密的数据长度
			SignatureLength:  [2]byte{}, // 应该设定为签名的长度
			ECC_Bind_Mode: 		eccBindMode,
			SigConfig:    		sigConfig,
			Reserved1:     		nil, // 读取这些应用特定的值
			Reserved2:    		nil, // 读取策略
			MetaInfo:      		nil, // 读取元信息
		},
		Payload: ZtdoPayload{
			IV:        	[8]byte{}, // 此处应填充实际 IV
			Ciphertxt: 	[]byte(ciphertxt), // 将加密后的数据传入 Ciphertxt
			MAC:      	[]byte{}, // 此处应填充实际的 MAC
		},
		Signature: ZtdoSignature{
			SignerId:          []byte{}, // 先占位，后续可能填充
			DomainId:          []byte{}, // 先占位，后续可能填充
			DeclaredPublicKey: []byte{}, // 先占位，后续可能填充
			Sig:               []byte{}, // 先占位，后续可能填充
		},
	}
	return ztdo, nil
}


// 写入 Ztdo 的函数
func WriteZtdo(writer io.Writer, ztdo Ztdo) error {
	// 写入 ZtdoHeader
	if err := WriteZtdoHeader(writer, &ztdo.Header); err != nil {
		return fmt.Errorf("failed to write ZtdoHeader: %w", err)
	}

	// 写入 ZtdoPayload
	if err := WriteZtdoPayload(writer, &ztdo.Payload); err != nil {
		return fmt.Errorf("failed to write ZtdoPayload: %w", err)
	}

	// 写入 ZtdoSignature
	if err := WriteZtdoSignature(writer, &ztdo.Signature); err != nil {
		return fmt.Errorf("failed to write ZtdoSignature: %w", err)
	}

	return nil
}

func WriteZtdoHeader(writer io.Writer, header *ZtdoHeader) error {
	// 写入 header 的字段
	_, err := writer.Write(header.Typeid[:])
	if err != nil {
		return err
	}
	// 在这里写入更多 header 字段...
	return nil
}

func WriteZtdoPayload(writer io.Writer, payload *ZtdoPayload) error {
	// 写入 payload 的密文等
	_, err := writer.Write(payload.Ciphertxt)
	if err != nil {
		return err
	}
	// 根据需要继续写入 IV 和 MAC
	return nil
}

func WriteZtdoSignature(writer io.Writer, signature *ZtdoSignature) error {
	// 写入 signature 的字段
	_, err := writer.Write(signature.SignerId) // 作为示例
	if err != nil {
		return err
	}
	// 继续处理其他字段...
	return nil
}

// Encrypt 加密明文
func Encrypt(plainText []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 填充明文
	plainText = pad(plainText, aes.BlockSize)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt 解密密文
func Decrypt(cipherText string, key []byte) ([]byte, error) {
	cipherTextDec, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherTextDec) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := cipherTextDec[:aes.BlockSize]
	cipherTextDec = cipherTextDec[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherTextDec, cipherTextDec)

	// 去掉填充
	return unpad(cipherTextDec), nil
}

// pad 和 unpad 函数
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
