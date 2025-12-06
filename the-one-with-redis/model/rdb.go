package model

type OpCode byte

const (
	// Op codes
	EOF          = 0xFF
	SELECTDB     = 0xFE
	EXPIRETIME   = 0xFD
	EXPIRETIMEMS = 0xFC
	RESIZEDB     = 0xFB
	AUX          = 0xFA
	
	// Value types
	STRING              = 0
	LIST                = 1
	SET                 = 2
	SORTED_SET          = 3
	HASH                = 4
	ZIPMAP              = 9
	ZIPLIST             = 10
	INTSET              = 11
	SORTED_SET_ZIPLIST  = 12
	HASH_ZIPLIST        = 13
	QUICKLIST           = 14
)

type Database struct {
	ID        string
	KeyValues map[string]KeyValuePairs
}

type KeyValuePairs struct {
	KeyExpiryTimestamp int
	Value              Value
	TypeFlag           ValueType
}

type Value struct {
}

type ValueType byte

const ()
