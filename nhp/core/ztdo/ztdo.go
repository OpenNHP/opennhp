package ztdo

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"reflect"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/google/uuid"
)

const (
	MagicNumberSize      = 4
	ObjectIDSize         = 16
	VersionSize          = 2
	NhpServerLenSize     = 1
	NhpServerMaxSize     = 255
	CipherConfigSize     = 1
	MetadataLenSize      = 2
	MetadataChunkMaxSize = 32767
	PayloadLengthSize    = 3
	IVSize               = 12
	SIGNATURELenSize     = 32
	LENGTHFOR            = "lengthFor"
	LENGTHCONTINUE       = "lengthContinue"
	SUBTRACTFROM         = "subtractFrom"
	DATACHUNKSIZE        = 16777187 // this is calculated by 2 ** 24 - 1 - IVSize (12 bytes) - MaxTagSize (16 bytes)
)

var lengthMap map[string]uint32 // map to store result after parsing the lengthFor tag
var lengthContinueIndicator bool = false
var littleEndian bool = true

// Endianness hides the endianness handling to make it easier to change the endianness of ztdo
type Endianness struct {
	littleEndian bool
}

func (e *Endianness) PutUint32(b []byte, v uint32) {
	if littleEndian {
		binary.LittleEndian.PutUint32(b, v)
	} else {
		binary.BigEndian.PutUint32(b, v)
	}
}

func (e *Endianness) Uint32(b []byte) uint32 {
	if littleEndian {
		return binary.LittleEndian.Uint32(b)
	} else {
		return binary.BigEndian.Uint32(b)
	}
}

var endianness = Endianness{littleEndian: littleEndian}

type ZtdoMetadata struct {
	MetadataLen [MetadataLenSize]byte `lengthFor:"Metadata" lengthContinue:"true"`
	// Metadata with variable length from 2 to 65508 bytes
	Metadata []byte
}

type ZtdoHeader struct {
	MagicNumber  [MagicNumberSize]byte
	ObjectID     [ObjectIDSize]byte
	Version      [VersionSize]byte
	NhpServerLen [NhpServerLenSize]byte `lengthFor:"NhpServer"`
	// NhpServer with variable length from 0 to 255 bytes
	NhpServer    []byte
	CipherConfig [CipherConfigSize]byte
	Metadata     []ZtdoMetadata
}

type ZtdoContent struct {
	Iv         [IVSize]byte `subtractFrom:"CipherText"`
	CipherText []byte
}

type ZtdoPayload struct {
	Length  [PayloadLengthSize]byte `lengthFor:"CipherText"`
	Content ZtdoContent
}

type ZtdoSignature struct {
	Signature [SIGNATURELenSize]byte
}

type Ztdo struct {
	header ZtdoHeader
	// No ZtdoPayload here, because it has variable length
	signature ZtdoSignature
}

func (header *ZtdoHeader) SetObjectID() {
	// generate uuid
	uuid := uuid.New()
	header.ObjectID = [ObjectIDSize]byte(uuid)

	// User first 4 bytes to set magic number
	header.MagicNumber = [MagicNumberSize]byte(uuid[0:4])
}

func (header *ZtdoHeader) GetObjectID() string {
	return uuid.UUID(header.ObjectID).String()
}

func (header *ZtdoHeader) SetVersion() {
	header.Version = [VersionSize]byte{0x00, 0x01}
}

func (header *ZtdoHeader) SetNhpServer(nhpServer string) error {
	if len(nhpServer) > NhpServerMaxSize {
		return fmt.Errorf("nhp server length is too long")
	}

	header.NhpServerLen[0] = byte(len(nhpServer))
	header.NhpServer = []byte(nhpServer)

	return nil
}

// SetMetadata supports variable length of metadata
func (header *ZtdoHeader) SetMetadata(metadata string) error {
	header.Metadata = []ZtdoMetadata{}

	totalLen := len(metadata)
	chunkNum := totalLen / MetadataChunkMaxSize

	for i := 0; i < chunkNum; i++ {
		chunk := metadata[i*MetadataChunkMaxSize : (i+1)*MetadataChunkMaxSize]
		chunkEncodedLen, err := encodeMetadataLength(len(chunk), true)
		if err != nil {
			return err
		}
		header.Metadata = append(header.Metadata, ZtdoMetadata{
			Metadata:    []byte(chunk),
			MetadataLen: chunkEncodedLen,
		})
	}

	chunk := metadata[chunkNum*MetadataChunkMaxSize:]
	chunkEncodedLen, _ := encodeMetadataLength(len(chunk), false)
	header.Metadata = append(header.Metadata, ZtdoMetadata{
		Metadata:    []byte(chunk),
		MetadataLen: chunkEncodedLen,
	})

	return nil
}

func (header *ZtdoHeader) GetMetadata() []byte {
	var metadata []byte

	for _, metadataChunk := range header.Metadata {
		metadata = append(metadata, metadataChunk.Metadata...)
	}

	return metadata
}

func (header *ZtdoHeader) SetCipherConfig(hasSignature bool, mode SymmetricCipherMode, eccMode DataKeyPairECCMode) {
	if hasSignature {
		header.CipherConfig[0] |= 0x80
	} else {
		header.CipherConfig[0] |= 0
	}

	header.CipherConfig[0] |= byte(mode)
	header.CipherConfig[0] |= byte(eccMode) << 4
}

func (header *ZtdoHeader) HasSignature() bool {
	return header.CipherConfig[0]&0x80 != 0
}

func (header *ZtdoHeader) GetCipherMode() SymmetricCipherMode {
	return SymmetricCipherMode(header.CipherConfig[0] & 0x0F)
}

func (header *ZtdoHeader) GetECCMode() DataKeyPairECCMode {
	return DataKeyPairECCMode((header.CipherConfig[0] & 0x7F) >> 4)
}

func (payload *ZtdoPayload) SetIV() {
	rand.Read(payload.Content.Iv[:])
}

func (payload *ZtdoPayload) SetCipherText(mode SymmetricCipherMode, key, plaintext []byte, ad []byte) error {
	var err error

	payload.Content.CipherText, err = mode.Encrypt(key, payload.Content.Iv[:], plaintext, ad)
	if err != nil {
		return err
	}

	return nil
}

func (payload *ZtdoPayload) GetPlainText(mode SymmetricCipherMode, key []byte, ad []byte) ([]byte, error) {
	return mode.Decrypt(key, payload.Content.Iv[:], payload.Content.CipherText, ad)
}

func (payload *ZtdoPayload) SetLength() {
	payloadLen := IVSize + len(payload.Content.CipherText)

	tmp := make([]byte, 4)
	endianness.PutUint32(tmp, uint32(payloadLen))

	copy(payload.Length[:], tmp[:PayloadLengthSize])
}

func (payload *ZtdoPayload) GetLength() uint32 {
	tmp := make([]byte, 4)

	copy(tmp[:3], payload.Length[:])

	return endianness.Uint32(tmp[:])
}

func NewZtdoHeader() *ZtdoHeader {
	header := &ZtdoHeader{}

	header.SetObjectID()
	header.SetVersion()
	// header.SetNhpServer("")

	return header
}

func NewZtdoPayload() *ZtdoPayload {
	payload := &ZtdoPayload{}

	payload.SetIV()

	return payload
}

func NewZtdoSignature() *ZtdoSignature {
	signature := &ZtdoSignature{}

	return signature
}

func (signature *ZtdoSignature) mixHash(buf *bytes.Buffer) {
	hashSig := core.NewHash(core.HASH_SHA256)

	hashSig.Write(signature.Signature[:])
	hashSig.Write(buf.Bytes())
	copy(signature.Signature[:], hashSig.Sum(nil))
}

func (signature *ZtdoSignature) sign(key []byte) {
	newHash := func() hash.Hash {
		return core.NewHash(core.HASH_SHA256)
	}

	hmac := hmac.New(newHash, key)
	hmac.Write(signature.Signature[:])
	copy(signature.Signature[:], hmac.Sum(nil))

	hmac.Reset()
}

func (signature *ZtdoSignature) verify(in *ZtdoSignature) bool {
	return bytes.Equal(signature.Signature[:], in.Signature[:])
}

func NewZtdo() *Ztdo {
	return &Ztdo{
		header:    *NewZtdoHeader(),
		signature: *NewZtdoSignature(),
	}
}

func (ztdo *Ztdo) Generate(mode DataKeyPairECCMode) (privateKey []byte) {
	ecdh := core.NewECDH(mode.ToEccType())
	return ecdh.PrivateKey()
}

func (ztdo *Ztdo) SetNhpServer(nhpServer string) error {
	return ztdo.header.SetNhpServer(nhpServer)
}

func (ztdo *Ztdo) SetCipherConfig(hasSignature bool, mode SymmetricCipherMode, eccMode DataKeyPairECCMode) {
	ztdo.header.SetCipherConfig(hasSignature, mode, eccMode)
}

func (ztdo *Ztdo) SetMetadata(metadata string) error {
	return ztdo.header.SetMetadata(metadata)
}

func (ztdo *Ztdo) GetObjectID() string {
	return ztdo.header.GetObjectID()
}

func (ztdo *Ztdo) GetCipherMode() SymmetricCipherMode {
	return ztdo.header.GetCipherMode()
}

func (ztdo *Ztdo) GetECCMode() DataKeyPairECCMode {
	return ztdo.header.GetECCMode()
}

func (ztdo *Ztdo) EncryptZtdoFile(plaintextPath, ciphertextPath string, gcmKey []byte, ad []byte) error {
	plaintextFile, err := os.OpenFile(plaintextPath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer plaintextFile.Close()

	ciphertextFile, err := os.Create(ciphertextPath)
	if err != nil {
		return err
	}
	defer ciphertextFile.Close()

	// If metadata is empty, set it to an empty string
	if len(ztdo.header.GetMetadata()) == 0 {
		ztdo.SetMetadata("")
	}

	// Write header
	headerBuf := toBuffer(ztdo.header)
	ztdo.signature.mixHash(headerBuf)
	ciphertextFile.Write(headerBuf.Bytes())

	// Write payloads
	for {
		chunkSize := getSecureRandomChunkSize()
		buf := make([]byte, chunkSize)
		n, err := plaintextFile.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			if n == 0 { // protect
				break
			}
		}

		payload := NewZtdoPayload()
		payload.SetCipherText(ztdo.header.GetCipherMode(), gcmKey, buf[:n], ad)
		payload.SetLength()
		payloadBuf := toBuffer(payload)
		ztdo.signature.mixHash(payloadBuf)
		ciphertextFile.Write(payloadBuf.Bytes())
	}

	// update signature
	ztdo.signature.sign(gcmKey)
	ciphertextFile.Write(toBuffer(ztdo.signature).Bytes())

	return nil
}

func (ztdo *Ztdo) ParseHeader(ciphertextPath string) error {
	ciphertextFile, err := os.OpenFile(ciphertextPath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer ciphertextFile.Close()

	// always initialize header
	ztdo.header = *NewZtdoHeader()

	// read header
	if err := toStructure(ciphertextFile, &ztdo.header); err != nil {
		return err
	}

	return nil
}

func (ztdo *Ztdo) DecryptZtdoFile(ciphertextPath, plaintextPath string, gcmKey []byte, ad []byte) error {
	ciphertextFile, err := os.OpenFile(ciphertextPath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer ciphertextFile.Close()

	ciphertextFileInfo, err := ciphertextFile.Stat()
	if err != nil {
		return err
	}

	remainingCiphertextFileSize := ciphertextFileInfo.Size()

	plaintextFile, err := os.Create(plaintextPath)
	if err != nil {
		return err
	}
	defer plaintextFile.Close()

	recalcSig := NewZtdoSignature()

	// always initialize header
	ztdo.header = *NewZtdoHeader()

	// read header
	if err := toStructure(ciphertextFile, &ztdo.header); err != nil {
		return err
	}
	headerBuf := toBuffer(ztdo.header)
	recalcSig.mixHash(headerBuf)
	remainingCiphertextFileSize -= int64(headerBuf.Len())

	// read payload
	for {
		payload := NewZtdoPayload()
		if err := toStructure(ciphertextFile, payload); err != nil {
			return err
		}

		payloadBuf := toBuffer(payload)
		recalcSig.mixHash(payloadBuf)
		remainingCiphertextFileSize -= int64(payloadBuf.Len())

		plaintext, err := payload.GetPlainText(ztdo.header.GetCipherMode(), gcmKey, ad)
		if err != nil {
			return err
		}
		plaintextFile.Write(plaintext)

		if ztdo.header.HasSignature() {
			if remainingCiphertextFileSize == SIGNATURELenSize {
				break
			}
			if remainingCiphertextFileSize == 0 {
				return fmt.Errorf("invalid ztdo file")
			}
		} else {
			if remainingCiphertextFileSize == 0 {
				break
			}
		}
	}

	// update recalculated signature
	recalcSig.sign(gcmKey)

	if ztdo.header.HasSignature() {
		if err := toStructure(ciphertextFile, &ztdo.signature); err != nil {
			return err
		}

		if !ztdo.signature.verify(recalcSig) {
			return fmt.Errorf("signature verification failed")
		}
	}

	return nil
}

// marshal searilizs Go struct into bytes buffer
func marshal(buf *bytes.Buffer, data any) error {
	rData := reflect.ValueOf(data)
	if rData.Kind() == reflect.Ptr {
		rData = rData.Elem()
	}

	if rData.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct")
	}

	for i := range rData.NumField() {
		field := rData.Field(i)
		if field.Type().Kind() == reflect.Struct {
			if err := marshal(buf, field.Interface()); err != nil {
				return err
			}
		} else {
			var bytes []byte
			if field.Type().Kind() == reflect.Array { // here assume it's an array of byte.
				bytes = make([]byte, field.Len())
				reflect.Copy(reflect.ValueOf(bytes), field)
			} else if field.Type().Kind() == reflect.Slice {
				if field.Type().Elem().Kind() == reflect.Struct {
					for j := range field.Len() {
						if err := marshal(buf, field.Index(j).Interface()); err != nil {
							return err
						}
					}
				} else if field.Type().Elem().Kind() == reflect.Uint8 {
					bytes = field.Interface().([]byte)
				} else {
					return fmt.Errorf("unsupported field type: %v in slice", field.Type().Elem().Kind())
				}
			} else {
				return fmt.Errorf("unsupported field type: %v", field.Type().Kind())
			}

			if _, err := buf.Write(bytes); err != nil {
				return err
			}
		}
	}

	return nil
}

// unmarshal recursively deserializes binary data from a file into a Go struct
func unmarshal(f *os.File, data any) error {
	rValues := reflect.ValueOf(data)
	rTypes := reflect.TypeOf(data)
	if rValues.Kind() == reflect.Ptr {
		rTypes = rTypes.Elem()
		rValues = rValues.Elem()
	}

	if rValues.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct")
	}

	for i := range rTypes.NumField() {
		field := rTypes.Field(i)
		value := rValues.Field(i)
		if field.Type.Kind() == reflect.Struct {
			if err := unmarshal(f, value.Addr().Interface()); err != nil {
				return err
			}
		} else {
			if field.Type.Kind() == reflect.Array {
				length := value.Len()
				bytes := make([]byte, length)
				_, err := f.Read(bytes)
				if err != nil {
					return err
				}

				setBytes(value, bytes)

				// parse lengthFor and lengthContinue tag
				// lengthFor tag means that this field represents the length of the field which is specified in the lengthFor tag
				// lengthContinue tag is usually used for slices which don't have fixed length.
				// lengthContinue tag means that the MSB (most significant bit) of this field indicates that there is still more data to be read
				lengthFor := field.Tag.Get(LENGTHFOR)
				lengthContinue := field.Tag.Get(LENGTHCONTINUE)

				if lengthContinue != "" {
					lengthContinueIndicator = preprocessContinuation(bytes)
				}

				if lengthFor != "" {
					expandedBytes := [4]byte{}
					copy(expandedBytes[:], bytes)
					lengthMap[lengthFor] = endianness.Uint32(expandedBytes[:])
				}

				subtractFrom := field.Tag.Get(SUBTRACTFROM)
				if subtractFrom != "" {
					lengthMap[subtractFrom] -= uint32(length)
				}
			} else if field.Type.Kind() == reflect.Slice {
				if field.Type.Elem().Kind() == reflect.Struct {
					for {
						ins := reflect.New(field.Type.Elem()).Elem()
						newValue := reflect.Append(value, ins)
						value.Set(newValue)
						err := unmarshal(f, value.Index(value.Len()-1).Addr().Interface())
						if err != nil {
							return err
						}
						if !lengthContinueIndicator { // more data to be read to construct element in current slice
							break
						}
					}
				} else if field.Type.Elem().Kind() == reflect.Uint8 { // if the elment in slice is a byte, there MUST be lengthFor tag to be parsed before.
					length := lengthMap[field.Name]
					bytes := make([]byte, length)
					_, err := f.Read(bytes)
					if err != nil {
						return err
					}
					setBytes(value, bytes)
				} else {
					return fmt.Errorf("unsupported element type: %v in slice", field.Type.Elem().Kind())
				}
			} else {
				return fmt.Errorf("unsupported field type: %v", field.Type.Kind())
			}
		}
	}

	return nil
}

// toBuffer provides unified way to serialize a Go struct into bytes buffer
func toBuffer(data any) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	marshal(buf, data)
	return buf
}

// toStructure provides unified way to deserialize bytes from a file into a Go struct
func toStructure(f *os.File, data any) error {
	lengthMap = make(map[string]uint32)
	rValues := reflect.ValueOf(data)
	if rValues.Kind() != reflect.Ptr {
		return fmt.Errorf("data must be a pointer")
	}
	return unmarshal(f, data)
}

// setBytes provides unified way to set bytes to a slice or array
func setBytes(rvalue reflect.Value, dst []byte) {
	if rvalue.Kind() == reflect.Slice {
		rvalue.SetBytes(dst)
	} else if rvalue.Kind() == reflect.Array {
		len := rvalue.Len()
		elType := rvalue.Type().Elem()

		arrayType := reflect.ArrayOf(len, elType)
		newArray := reflect.New(arrayType).Elem()

		for i := range len {
			newArray.Index(i).SetUint(uint64(dst[i]))
		}

		rvalue.Set(newArray)

	} else {
		panic("not support")
	}
}

// encodeMetadataLength encodes a length continuation bit into the MSB of metadata length
func encodeMetadataLength(length int, continuation bool) ([2]byte, error) {
	var result [2]byte

	if length > MetadataChunkMaxSize {
		return result, fmt.Errorf("length %d exceeds maximum encoded value (%d)", length, MetadataChunkMaxSize)
	}

	if length < 0 {
		return result, fmt.Errorf("length %d is negative", length)
	}

	result[0] = byte((length >> 8) & 0x7F)
	if continuation {
		result[0] |= 0x80
	}
	result[1] = byte(length & 0xFF)

	if littleEndian {
		result[0], result[1] = result[1], result[0]
	}

	return result, nil
}

// preprocessContinuation decodes a length continuation bit from the MSB of metadata length
func preprocessContinuation(encoded []byte) (continuation bool) {
	if littleEndian {
		continuation = encoded[len(encoded)-1]&0x80 != 0
		encoded[len(encoded)-1] &= 0x7F
	} else {
		continuation = encoded[0]&0x80 != 0
		encoded[0] &= 0x7F
	}

	return
}

// getSecureRandomChunkSize generates a random chunk size within a specified range
func getSecureRandomChunkSize() int {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		panic(err)
	}

	// Get random float between 0 and 1
	r := float64(binary.BigEndian.Uint64(buf[:])) / math.MaxUint64

	// Scale to desired range (e.g., 80% to 100% of base size)
	min := float64(DATACHUNKSIZE) * 0.8
	max := float64(DATACHUNKSIZE) * 1.0
	return int(min + r*(max-min))
}
