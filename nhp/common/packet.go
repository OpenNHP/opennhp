package common

// header flags (bit 0 - bit 11)
const (
	NHP_FLAG_EXTENDEDLENGTH = 1 << iota
	NHP_FLAG_COMPRESS
	NHP_FLAG_CL_PKC
)

// cipher scheme combination (bit 11 - bit 15)
const (
	NHP_FLAG_SCHEME_GMSM = 0 << 12
)

const (
	CIPHER_SCHEME_GMSM int = iota
	CIPHER_SCHEME_CURVE
)
