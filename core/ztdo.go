package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"path/filepath"

	"encoding/json"

	"github.com/OpenNHP/opennhp/utils"
	"github.com/google/uuid"
)

// Ztdo complete data structure definition
type Ztdo struct {
	Header    ZtdoHeader
	Payload   ZtdoPayload
	Signature ZtdoSignature
}

// ZtdoHeader represents the header structure of the data object
type ZtdoHeader struct {
	Typeid          [4]byte      // Type identifier of ZTDO
	Objectid        string       // ZTDO Data Object Identifier
	Version         [2]byte      // Version Number
	PayloadOffset   [2]byte      // The offset relative to the starting address is the starting address of IV.
	PayloadLength   [8]byte      // The total length of the Payload data item, including IV, ciphertext, and MAC
	SignatureLength [2]byte      // The length is the length of the data after the signature data structure ASN.1 encoding
	ECC_Bind_Mode   [1]byte      // Indicates whether the data used for the elliptic curve parameter and policy binding is a GMAC tag or an ECDSA signature
	SigConfig       [1]byte      // Indicates the data encryption algorithm and data signature algorithm of the Payload
	MetaInfo        ZtdoMetaInfo // Meta information
}

// ZtdoPayload represents the payload structure of the data object
type ZtdoPayload struct {
	IV        [16]byte // Initialization Vector (IV)
	Ciphertxt []byte   // Encrypted payload data
	MAC       [16]byte // Message Authentication Code (MAC)
}

// ZtdoSignature represents the signature structure of a data object
type ZtdoSignature struct {
	SignerId          []byte // Signer ID
	DomainId          []byte // Domain ID
	DeclaredPublicKey []byte // Public Key Declaration
	Sig               []byte //Signature Data
}
type ZtdoMetaInfo struct {
	FileName    string
	FileExtName string
}

// 加密Zto文件真实的内容
type ZtdoFile struct {
	Objectid    string `json:"objectid"`    //文件唯一id
	FileName    string `json:"fileName"`    //原始文件名称，如test.txt
	FileExt     string `json:"fileExt"`     //原始文件扩展名,如 .txt
	FileContent []byte `json:"fileContent"` //加密后的原始文件内容
	CreateDate  string `json:"createDate"`  //创建日期
}

// Encryption Function
func Encrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//Padding Plaintext
	plainText = pad(plainText, aes.BlockSize)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)
	return cipherText, nil
}

// Filling function
func pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// 解密
func Decrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 获取IV，IV是cipherText的前16字节
	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// 解密
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)

	// 去除填充
	decrypted = unpad(decrypted, aes.BlockSize)

	return decrypted, nil
}
func unpad(padded []byte, blockSize int) []byte {
	length := len(padded)
	unpadLen := int(padded[length-1])
	if unpadLen > blockSize || unpadLen > length {
		return nil
	}
	return padded[:length-unpadLen]
}

// Calculate the offset of the Payload
func CalculatePayloadOffset(ciphertxt []byte) [2]byte {
	payloadOffset := uint16(len(ciphertxt)) // Calculate the offset from the ciphertext
	var offset [2]byte
	offset[0] = byte(payloadOffset >> 8)
	offset[1] = byte(payloadOffset)
	return offset
}

// Calculate the total length of the Payload (including IV, ciphertext and MAC)
func CalculatePayloadLength(ciphertxt []byte, ivLength int, macLength int) [8]byte {
	// Calculate the total length: IV length + ciphertext length + MAC length
	payloadLength := uint64(ivLength + len(ciphertxt) + macLength)
	var length [8]byte
	for i := 0; i < 8; i++ {
		length[7-i] = byte(payloadLength >> (i * 8))
	}
	return length
}

// Generate ObjectId
func GenerateObjectId() (uuid.UUID, error) {
	objectid := uuid.New()
	return objectid, nil
}

// Convert Hex string to byte array
func ConvertHexStringToByteArray(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// Calculating TypeId
func CalculateTypeid(objectId []byte) ([4]byte, error) {
	var typeid [4]byte
	copy(typeid[:], objectId[:4])
	return typeid, nil
}

// Setting the ECC bonding mode
func SetECCBindMode(curveParam byte, useSignature bool) [1]byte {
	var mode [1]byte
	if useSignature {
		mode[0] = 1
	} else {
		mode[0] = 0
	}
	return mode
}

// Set up signing configuration
func SetSigConfig(symAlgMode byte, curveParam byte, hasSignature bool) [1]byte {
	var sigConfig [1]byte
	if hasSignature {
		sigConfig[0] = 1
	} else {
		sigConfig[0] = 0
	}
	return sigConfig
}

// Parsing source files
func WriteSourceFile(filename string, key []byte) (Ztdo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Ztdo{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	payloadData, err := io.ReadAll(file)
	if err != nil {
		return Ztdo{}, fmt.Errorf("error reading file: %v", err)
	}

	objectidResult, err := GenerateObjectId()
	if err != nil {
		return Ztdo{}, err
	}
	hexString := strings.ReplaceAll(objectidResult.String(), "-", "")
	objectId, err := ConvertHexStringToByteArray(hexString)
	if err != nil {
		return Ztdo{}, err
	}

	typeid, err := CalculateTypeid(objectId)
	if err != nil {
		return Ztdo{}, err
	}

	curveParam := byte(1)
	useSignature := true
	eccBindMode := SetECCBindMode(curveParam, useSignature)

	symAlgMode := byte(1)
	hasSignature := true
	sigConfig := SetSigConfig(symAlgMode, curveParam, hasSignature)

	ciphertxt, err := Encrypt(payloadData, key)
	if err != nil {
		return Ztdo{}, fmt.Errorf("failed to encrypt data: %v", err)
	}

	// The default IV length is 16 bytes, and the MAC length is 16 bytes
	ivLength := 16
	macLength := 16

	payloadOffset := CalculatePayloadOffset(ciphertxt)
	payloadLength := CalculatePayloadLength(ciphertxt, ivLength, macLength)

	//获取文件扩展名
	fileExt := filepath.Ext(filename)
	fileName := filepath.Base(filename)
	metaInfo := ZtdoMetaInfo{
		FileName:    fileName,
		FileExtName: fileExt,
	}
	fmt.Println("File Name:", fileName)
	fmt.Println("File Ext:", fileExt)
	ztdo := Ztdo{
		Header: ZtdoHeader{
			Typeid:          typeid,
			Objectid:        objectidResult.String(),
			Version:         [2]byte{1, 0},
			PayloadOffset:   payloadOffset,
			PayloadLength:   payloadLength,
			SignatureLength: [2]byte{},
			ECC_Bind_Mode:   eccBindMode,
			SigConfig:       sigConfig,
			MetaInfo:        metaInfo,
		},
		Payload: ZtdoPayload{
			IV:        [16]byte{},
			Ciphertxt: ciphertxt,
			MAC:       [16]byte{},
		},
		Signature: ZtdoSignature{
			SignerId:          []byte{},
			DomainId:          []byte{},
			DeclaredPublicKey: []byte{},
			Sig:               []byte{},
		},
	}

	return ztdo, nil
}

// Write Ztdo data to a file
func WriteZtdo(writer io.Writer, ztdo Ztdo) error {
	// if err := WriteZtdoHeader(writer, &ztdo.Header); err != nil {
	// 	return fmt.Errorf("failed to write ZtdoHeader: %w", err)
	// }
	// if err := WriteZtdoPayload(writer, &ztdo.Payload); err != nil {
	// 	return fmt.Errorf("failed to write ZtdoPayload: %w", err)
	// }
	// if err := WriteZtdoSignature(writer, &ztdo.Signature); err != nil {
	// 	return fmt.Errorf("failed to write ZtdoSignature: %w", err)
	// }
	date := utils.GetCurrentDate()
	ztdoFile := ZtdoFile{
		Objectid:    ztdo.Header.Objectid,
		FileName:    ztdo.Header.MetaInfo.FileName,
		FileExt:     ztdo.Header.MetaInfo.FileExtName,
		FileContent: ztdo.Payload.Ciphertxt,
		CreateDate:  date,
	}

	jsonData, err := json.Marshal(ztdoFile)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}
	// 将字节切片转换为字符串
	fileDataJsonString := string(jsonData)
	_, err = writer.Write([]byte(fileDataJsonString))
	return err
}

// Write ZtdoHeader data
func WriteZtdoHeader(writer io.Writer, header *ZtdoHeader) error {
	_, err := writer.Write(header.Typeid[:])
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(header.Objectid))
	return err
}

// Writing ZtdoPayload data
func WriteZtdoPayload(writer io.Writer, payload *ZtdoPayload) error {
	_, err := writer.Write(payload.Ciphertxt)
	return err
}

// Writing ZtdoSignature data
func WriteZtdoSignature(writer io.Writer, signature *ZtdoSignature) error {
	_, err := writer.Write(signature.Sig)
	return err
}
//读取Ztdo文件内容到ZtdoFile对象中
func ReadZtdoFile(filePath string) (ZtdoFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ZtdoFile{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return ZtdoFile{}, fmt.Errorf("error reading file: %v", err)
	}

	var ztdoFile ZtdoFile
	// 使用json.Unmarshal将JSON字符串解码到person实例中
	err = json.Unmarshal(fileContentByte, &ztdoFile)
	if err != nil {
		return ZtdoFile{}, fmt.Errorf("JSON解析错误: %s", err)
	}
	return ztdoFile, nil
}