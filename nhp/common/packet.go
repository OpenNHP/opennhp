package common

// header flags (bit 0 - bit 11)
const (
	NHP_FLAG_EXTENDEDLENGTH = 1 << iota
	NHP_FLAG_COMPRESS
	NHP_FLAG_CL_PKC
)

// cipher scheme combination (bit 12 - bit 15)
const (
	NHP_FLAG_SCHEME_CURVE = 0 << 12
	NHP_FLAG_SCHEME_GMSM  = 1 << 12
)

const (
	CIPHER_SCHEME_CURVE int = iota // 0 - Curve25519/Blake2s/AES256 (international standard)
	CIPHER_SCHEME_GMSM             // 1 - SM2/SM3/SM4 (Chinese national standard)
)
