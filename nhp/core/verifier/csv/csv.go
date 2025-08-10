package csv

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"unsafe"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
)

var (
	SM2A  = "fffffffeffffffffffffffffffffffffffffffff00000000fffffffffffffffc"
	SM2B  = "28e9fa9e9d9f5e344d5a9e4bcf6509a7f39789f515ab8f92ddbcbd414d940e93"
	SM2GX = "32c4ae2c1f1981195f9904466a39c9948fe30bbff2660be1715a4589334c74c7"
	SM2GY = "bc3736a2f4f6779c59bdcee36b692153d0a9877cc62a474002df32e52139f0a0"
	ECKEY = SM2A + SM2B + SM2GX + SM2GY
	Verifier *Attestation = nil
)

type AttestationBody struct {
	UserPubKeyDigest [32]byte `json:"user_pubkey_digest"`
	VmId             [16]byte `json:"vm_id"`
	VmVersion        [16]byte `json:"vm_version"`
	ReportData       [64]byte `json:"report_data"`
	Mnonce           [16]byte `json:"mnonce"`
	Measure          [32]byte `json:"measure"`
	Policy           uint32   `json:"policy"`
}

type Signature struct {
	R [72]byte `json:"r"`
	S [72]byte `json:"s"`
}

type AttestationReport struct {
	AttestationBody AttestationBody `json:"body"`
	SigUsage        uint32          `json:"sig_usage"`
	SigAlgo         uint32          `json:"sig_algo"`
	Anonce          uint32          `json:"anonce"`
	Sig             Signature       `json:"sig"`
}

type CertificateData struct {
	Kid      [16]byte `json:"kid"`
	Sid      [16]byte `json:"sid"`
	Usage    uint32   `json:"usage"`
	Reserved [24]byte `json:"reserved"`
}

type PubKey struct {
	G uint32   `json:"g"`
	X [72]byte `json:"x"`
	Y [72]byte `json:"y"`
}

type CertificatePreamble struct {
	Version uint32          `json:"ver"`
	Data    CertificateData `json:"data"`
}

type CaCertificateBody struct {
	Preamble CertificatePreamble `json:"preamble"`
	PubKey   PubKey              `json:"pubkey"`
	UidSize  uint16              `json:"uid_size"`
	UserId   [254]byte           `json:"user_id"`
	Reserved [108]byte           `json:"reserved"`
}

type Version struct {
	Major uint8 `json:"major"`
	Minor uint8 `json:"minor"`
}

type CsvPubkey struct {
	Usage uint32 `json:"usage"`
	Algo  uint32 `json:"algo"`
	Key   PubKey `json:"key"`
}

type CsvCertificateData struct {
	Firmware  Version   `json:"firmware"`
	Reserved1 uint16    `json:"reserved1"`
	PubKey    CsvPubkey `json:"pubkey"`
	UidSize   uint16    `json:"uid_size"`
	UserId    [254]byte `json:"user_id"`
	Sid       [16]byte  `json:"sid"`
	Reserved2 [608]byte `json:"reserved2"`
}

type CsvCertificateBody struct {
	Version uint32             `json:"ver"`
	Data    CsvCertificateData `json:"data"`
}

type CsvSignature struct {
	Usage     uint32    `json:"usage"`
	Algo      uint32    `json:"algo"`
	Signature Signature `json:"signature"`
	Reserved  [368]byte `json:"_reserved"`
}

type CsvCertificate struct {
	Body      CsvCertificateBody `json:"body"`
	Signature [2]CsvSignature    `json:"sigs"`
}

type CaCertificate struct {
	Body      CaCertificateBody `json:"body"`
	Signature Signature         `json:"signature"`
	Reserved  [112]byte         `json:"_reserved"`
}

type CertificateChain struct {
	Hsk CaCertificate  `json:"hsk"`
	Cek CsvCertificate `json:"cek"`
	Pek CsvCertificate `json:"pek"`
}

type CsvEvidence struct {
	AttestationReport AttestationReport `json:"attestation_report"`
	CertificateChain  CertificateChain  `json:"cert_chain"`
	SerialNumber      []byte            `json:"serial_number"`
}

type Attestation struct {
	evidence *CsvEvidence
	hrk      []byte
	hskCek   map[string][]byte
}

func ReverseBytes(b []byte) []byte {
	reversed := make([]byte, len(b))
	for i := range b {
		reversed[len(b)-1-i] = b[i]
	}
	return reversed
}

func buildIDMsg(id []byte, idLen int, ecKeyHex string, pubkeyHex string) []byte {
	// 计算 (id_len * 8) >> 8 % 256 和 (id_len * 8) % 256
	idLenBits := idLen * 8
	firstByte := byte((idLenBits >> 8) % 256)
	secondByte := byte(idLenBits % 256)

	// 转换十六进制字符串为字节
	ecKeyBytes, _ := hex.DecodeString(ecKeyHex)
	pubkeyBytes, _ := hex.DecodeString(pubkeyHex)

	// 拼接所有字节切片
	var result []byte
	result = append(result, firstByte)
	result = append(result, secondByte)
	result = append(result, id...)
	result = append(result, ecKeyBytes...)
	result = append(result, pubkeyBytes...)

	return result
}

func Sm3Digest(hrkData []byte) ([]byte, error) {
	hash := sm3.New()
	if _, err := hash.Write(hrkData); err != nil {
		return nil, fmt.Errorf("failed to write HRK data to SM3: %v", err)
	}
	digest := hash.Sum(nil)
	return digest, nil
}

func Sm3Hmac(data []byte, key []byte) []byte {
	// Block size of SM3 is 64 bytes (as specified in GM/T 0004-2012)
	const blockSize = 64

	// Ensure key is not longer than block size by hashing if necessary
	if len(key) > blockSize {
		hash := sm3.Sum(key)
		key = hash[:]
	}

	// Pad key to block size with zeros
	paddedKey := make([]byte, blockSize)
	copy(paddedKey, key)

	// Create inner and outer padding
	innerPad := make([]byte, blockSize)
	outerPad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		innerPad[i] = paddedKey[i] ^ 0x36 // Inner padding: 0x36 repeated
		outerPad[i] = paddedKey[i] ^ 0x5C // Outer padding: 0x5C repeated
	}

	// Compute inner hash: SM3(innerPad || data)
	innerHash := sm3.New()
	innerHash.Write(innerPad)
	innerHash.Write(data)
	innerResult := innerHash.Sum(nil)

	// Compute outer hash: SM3(outerPad || innerResult)
	outerHash := sm3.New()
	outerHash.Write(outerPad)
	outerHash.Write(innerResult)
	return outerHash.Sum(nil)
}

// refer to 7.1 in GB/T32918.2—2016
func VerifySignature(pub *ecdsa.PublicKey, hash []byte, r, s *big.Int) bool {
	n := pub.Curve.Params().N

	if r.Sign() <= 0 || s.Sign() <= 0 || r.Cmp(n) >= 0 || s.Cmp(n) >= 0 {
		return false
	}

	sm2_N := pub.Params().N

	// t = (r + s) % n
	t := new(big.Int).Mod(new(big.Int).Add(r, s), sm2_N)

	e := new(big.Int).SetBytes(hash)

	// x1, y1 = [r]G
	// x2, y2 = [s]PubKey
	x1, y1 := pub.Curve.ScalarBaseMult(s.Bytes())
	x2, y2 := pub.Curve.ScalarMult(pub.X, pub.Y, t.Bytes())

	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		// x1, y1 = Double(x1, y1)
		x1, y1 = pub.Curve.Double(x1, y1)
	} else {
		// x1, y1 = x1 + x2, y1 + y2
		x1, y1 = pub.Curve.Add(x1, y1, x2, y2)
	}

	return r.Cmp(new(big.Int).Mod(new(big.Int).Add(x1, e), sm2_N)) == 0
}

func (a *Attestation) verifySm2SignatureWithId(qx, qy, r, s []byte, id []byte, msg []byte) error {
	if len(qx) != 32 || len(qy) != 32 {
		return fmt.Errorf("invalid public key length: got %d, want 32", len(qx))
	}

	if len(r) != 32 || len(s) != 32 {
		return fmt.Errorf("invalid signature length: got %d, want 32", len(r))
	}

	qx = ReverseBytes(qx)
	qy = ReverseBytes(qy)
	r = ReverseBytes(r)
	s = ReverseBytes(s)

	pubKeyBytes := append(qx, qy...)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	id_msg := buildIDMsg(id, len(id), ECKEY, pubKeyHex)

	za, err := Sm3Digest(id_msg)
	if err != nil {
		return err
	}

	msgAll := append(za, msg...)
	msgAllDigest, err := Sm3Digest(msgAll)
	if err != nil {
		return err
	}

	xBig := new(big.Int).SetBytes(qx)
	yBig := new(big.Int).SetBytes(qy)

	pubKey := &ecdsa.PublicKey{
		Curve: sm2.P256(),
		X:     xBig,
		Y:     yBig,
	}

	rBig := new(big.Int).SetBytes(r)
	sBig := new(big.Int).SetBytes(s)

	if VerifySignature(pubKey, msgAllDigest, rBig, sBig) {
		return nil
	} else {
		return fmt.Errorf("failed to verify signature")
	}
}

func (a *Attestation) verifyCertChain(chipId string) error {
	// Download HRK from Hygon's certificate server
	if a.hrk == nil {
		resp, err := http.Get("https://cert.hygon.cn/hrk")
		if err != nil {
			return fmt.Errorf("failed to download HRK: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code when download HRK: %d", resp.StatusCode)
		}

		// Read the response body (HRK content)
		hrkData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read HRK data: %v", err)
		}

		a.hrk = hrkData
	}

	digest, err := Sm3Digest(a.hrk)
	if err != nil {
		return err
	}

	expectedDigest, _ := hex.DecodeString("f5a46663059fdb4cdd06d097ed21782142923bb3430b3b938f23d54292094e3a")
	if !bytes.Equal(digest, expectedDigest) {
		return fmt.Errorf("HRK digest verification failed: got %x, want %x", digest, expectedDigest)
	}

	if err := a.verifyHygonCertInfo(a.hrk, 0x03, 0, a.hrk[0x04:0x14]); err != nil {
		return err
	}

	// verify hrk cert signature (self-signed)
	hrkIdLen := int(binary.LittleEndian.Uint16(a.hrk[0xd4:0xd6]))
	if err := a.verifySm2SignatureWithId(
		a.hrk[0x44:0x64], a.hrk[0x8c:0xac],
		a.hrk[0x240:0x260], a.hrk[0x288:0x2a8],
		a.hrk[0xd6:0xd6+hrkIdLen], a.hrk[:0x240],
	); err != nil {
		return err
	}

	if _, ok := a.hskCek[chipId]; !ok {
		resp, err := http.Get(fmt.Sprintf("https://cert.hygon.cn/hsk_cek?snumber=%s", chipId))
		if err != nil {
			return fmt.Errorf("failed to download hsk_cek: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code when download hsk_cek: %d", resp.StatusCode)
		}

		hskCekData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read hsk_cek data: %v", err)
		}

		a.hskCek[chipId] = hskCekData
	}

	hskData := a.hskCek[chipId][:0x340]
	cekData := a.hskCek[chipId][0x340:]

	// verify hsk cert info
	if err := a.verifyHygonCertInfo(hskData, 0x03, 0x13, a.hrk[0x04:0x14]); err != nil {
		return err
	}

	// verify hsk cert signature (self-signed)
	if err := a.verifySm2SignatureWithId(
		a.hrk[0x44:0x64], a.hrk[0x8c:0xac],
		a.hskCek[chipId][0x240:0x260], a.hskCek[chipId][0x288:0x2a8],
		a.hrk[0xd6:0xd6+hrkIdLen], a.hskCek[chipId][:0x240],
	); err != nil {
		return err
	}

	// verify csv cert info
	if err := a.verifyCSVCertInfo(cekData, 0x13, 0x04, 0x1004, hskData[0x04:0x14]); err != nil {
		return err
	}

	// verify csv cert signature (self-signed)
	hskIdLen := int(binary.LittleEndian.Uint16(hskData[0xd4:0xd6]))
	if err := a.verifySm2SignatureWithId(
		hskData[0x44:0x64], hskData[0x8c:0xac],
		cekData[0x41c:0x43c], cekData[0x464:0x484],
		hskData[0xd6:0xd6+hskIdLen], cekData[:0x414],
	); err != nil {
		return err
	}

	// verify pek cert info
	pek := a.evidence.CertificateChain.Pek
	pekDataLen := unsafe.Sizeof(pek)
	pekData := (*[1 << 20]byte)(unsafe.Pointer(&pek))[:pekDataLen]

	// fmt.Printf("pek: %x\n", pekData)
	// fmt.Printf("body len of pek: %d\n", unsafe.Sizeof(pek.Body))

	if err := a.verifyCSVCertInfo(pekData, 0x1004, 0x04, 0x1002, hskData[0x1a4:0x1b4]); err != nil {
		return err
	}

	// verify pek cert signature (self-signed)
	cekIdLen := int(binary.LittleEndian.Uint16(cekData[0xa4:0xa6]))
	if err := a.verifySm2SignatureWithId(
		cekData[0x14:0x34], cekData[0x5c:0x7c],
		pekData[0x41c:0x43c], pekData[0x464:0x484],
		cekData[0xa6:0xa6+cekIdLen], pekData[:0x414],
	); err != nil {
		return err
	}

	return nil
}

func (a *Attestation) verifyHygonCertInfo(hrk []byte, curveId, keyUsage int, keyId []byte) error {
	hygonKeyUsage := hrk[0x24:0x28]

	hygonKeyUsageInt := int(binary.LittleEndian.Uint32(hygonKeyUsage))
	if hygonKeyUsageInt != keyUsage {
		return fmt.Errorf("key usage mismatch: got %d, want %d", keyUsage, keyUsage)
	}

	hygonCurveId := hrk[0x40:0x44]
	hygonCurveIdInt := int(binary.LittleEndian.Uint32(hygonCurveId))
	if hygonCurveIdInt != curveId {
		return fmt.Errorf("curve id mismatch: got %d, want %d", curveId, curveId)
	}

	hygonCertifyingId := hrk[0x14:0x24]
	if !bytes.Equal(hygonCertifyingId, keyId) {
		return fmt.Errorf("certifying id mismatch: got %x, want %x", hygonCertifyingId, keyId)
	}

	return nil
}

func (a *Attestation) verifyCSVCertInfo(csvCert []byte, sigUsage int, sigAlgo int, keyUsage int, keyId []byte) error {
	csvKeyUsage := csvCert[0x08:0x0C]
	csvKeyUsageInt := int(binary.LittleEndian.Uint32(csvKeyUsage))
	if csvKeyUsageInt != keyUsage {
		return fmt.Errorf("key usage mismatch: got %d, want %d", csvKeyUsageInt, sigUsage)
	}

	csvSigUsage := csvCert[0x414:0x418]
	csvSigUsageInt := int(binary.LittleEndian.Uint32(csvSigUsage))
	if csvSigUsageInt != sigUsage {
		return fmt.Errorf("sig usage mismatch: got %d, want %d", csvSigUsageInt, sigAlgo)
	}

	csvSigAlgo := csvCert[0x418:0x41C]
	csvSigAlgoInt := int(binary.LittleEndian.Uint32(csvSigAlgo))
	if csvSigAlgoInt != sigAlgo {
		return fmt.Errorf("sig algo mismatch: got %d, want %d", csvSigAlgoInt, sigAlgo)
	}

	csvCertifyingId := csvCert[0x1a4:0x1b4]
	if !bytes.Equal(csvCertifyingId, keyId) {
		return fmt.Errorf("certifying id mismatch: got %x, want %x", csvCertifyingId, keyId)
	}

	return nil
}

func (a *Attestation) Verify() error {
	if err := a.verifyCertChain(a.GetSerialNumber()); err != nil {
		return err
	}

	pek := a.evidence.CertificateChain.Pek
	attestationReport := a.evidence.AttestationReport
	attestationReportDataLen := unsafe.Sizeof(attestationReport)
	attestationReportData := (*[1 << 20]byte)(unsafe.Pointer(&attestationReport))[:attestationReportDataLen]

	if err := a.verifySm2SignatureWithId(
		pek.Body.Data.PubKey.Key.X[:32], pek.Body.Data.PubKey.Key.Y[:32],
		attestationReport.Sig.R[:32],
		attestationReport.Sig.S[:32],
		pek.Body.Data.UserId[:pek.Body.Data.UidSize],
		attestationReportData[:unsafe.Sizeof(attestationReport.AttestationBody)],
	); err != nil {
		return err
	}

	return nil
}

func (a *Attestation) GetSerialNumber() string {
	return string(bytes.TrimRight(a.evidence.SerialNumber, "\x00"))
}

func (a *Attestation) performXORBy4BytesGroup(data []byte, anouce uint32) []byte {
	result := make([]byte, 0, len(data))

	// Iterate through data in 4-byte steps
	for i := 0; i < len(data); i += 4 {
		// Create a 4-byte buffer
		var group [4]byte

		// Copy up to 4 bytes (handles partial last group)
		copy(group[:], data[i:min(i+4, len(data))])

		// Convert 4-byte group to int32 (using little-endian byte order)
		groupInt := binary.LittleEndian.Uint32(group[:])

		// Perform XOR with the key
		processedInt := uint32(uint32(groupInt) ^ anouce)

		// Convert back to bytes
		var processedGroup [4]byte
		binary.LittleEndian.PutUint32(processedGroup[:], processedInt)

		// Add to result, truncating if it's the last partial group
		end := i + 4
		if end > len(data) {
			end = len(data)
			result = append(result, processedGroup[:end-i]...)
		} else {
			result = append(result, processedGroup[:]...)
		}
	}

	return result
}

func (a *Attestation) GetMeasure() string {
	measure := a.performXORBy4BytesGroup(a.evidence.AttestationReport.AttestationBody.Measure[:], a.evidence.AttestationReport.Anonce)

	return hex.EncodeToString(measure)
}

func NewAttestation(attestationJsonStr string) (*Attestation, error) {
	var attestation *Attestation
	var evidence *CsvEvidence

	if Verifier == nil {
		attestation = &Attestation{}
		Verifier = attestation
	} else {
		attestation = Verifier
	}

	if err := json.Unmarshal([]byte(attestationJsonStr), &evidence); err != nil {
		return nil, err
	}

	attestation.hskCek = make(map[string][]byte)

	attestation.evidence = evidence

	return attestation, nil
}

