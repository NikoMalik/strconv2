package strconv2

import (
	"errors"
	"math"
	"unsafe"
)

const SAFETY_BUF_SIZE = 32
const FAST_BUF_SIZE = 24
const UINT16_MAX = 5

var (
	ErrOverflow         = errors.New("overflow")
	ErrInvalidCharacter = errors.New("invalid character")
	ErrInvalidString    = errors.New("invalid string")
	ErrEmptyString      = errors.New("empty string")
	ErrInvalidBoolStr   = errors.New("invalid parse bool argument")
)

var digits = [...]byte{
	'0', '0', '0', '1', '0', '2', '0', '3', '0', '4', '0', '5', '0', '6', '0', '7', '0', '8', '0', '9',
	'1', '0', '1', '1', '1', '2', '1', '3', '1', '4', '1', '5', '1', '6', '1', '7', '1', '8', '1', '9',
	'2', '0', '2', '1', '2', '2', '2', '3', '2', '4', '2', '5', '2', '6', '2', '7', '2', '8', '2', '9',
	'3', '0', '3', '1', '3', '2', '3', '3', '3', '4', '3', '5', '3', '6', '3', '7', '3', '8', '3', '9',
	'4', '0', '4', '1', '4', '2', '4', '3', '4', '4', '4', '5', '4', '6', '4', '7', '4', '8', '4', '9',
	'5', '0', '5', '1', '5', '2', '5', '3', '5', '4', '5', '5', '5', '6', '5', '7', '5', '8', '5', '9',
	'6', '0', '6', '1', '6', '2', '6', '3', '6', '4', '6', '5', '6', '6', '6', '7', '6', '8', '6', '9',
	'7', '0', '7', '1', '7', '2', '7', '3', '7', '4', '7', '5', '7', '6', '7', '7', '7', '8', '7', '9',
	'8', '0', '8', '1', '8', '2', '8', '3', '8', '4', '8', '5', '8', '6', '8', '7', '8', '8', '8', '9',
	'9', '0', '9', '1', '9', '2', '9', '3', '9', '4', '9', '5', '9', '6', '9', '7', '9', '8', '9', '9',
}

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func Digits10(v uint64) uint32 {
	if v < 10 {
		return 1
	}
	if v < 100 {
		return 2
	}
	if v < 1000 {
		return 3
	}
	if v < 1_000_000_000_000 {
		if v < 100_000_000 {
			if v < 1_000_000 {
				if v < 10_000 {
					return 4
				}
				return 5 + boolToUint32(v >= 100_000)
			}
			return 7 + boolToUint32(v >= 10_000_000)
		}
		if v < 10_000_000_000 {
			return 9 + boolToUint32(v >= 1_000_000_000)
		}
		return 11 + boolToUint32(v >= 100_000_000_000)
	}
	return 12 + Digits10(v/1_000_000_000_000)
}

func FormatUint6410(dst []byte, value uint64) int {
	dstlen := len(dst)

	length := Digits10(value)
	if int(length) >= dstlen {
		if dstlen > 0 {
			dst[0] = 0
		}
		return 0
	}
	next := length - 1
	dst[next+1] = 0

	for value >= 100 {
		i := (value % 100) * 2
		value /= 100
		dst[next] = digits[i+1]
		dst[next-1] = digits[i]
		next -= 2
	}

	if value < 10 {
		dst[next] = '0' + byte(value)
	} else {
		i := value * 2
		dst[next] = digits[i+1]
		dst[next-1] = digits[i]
	}

	return int(length)
}

func FormatInt6410(dst []byte, svalue int64) int {
	dstlen := len(dst)
	negative := 0
	var value uint64

	if svalue < 0 {
		if svalue != math.MinInt64 {
			value = uint64(-svalue)
		} else {
			value = uint64(math.MaxInt64) + 1
		}
		if dstlen < 2 {
			if dstlen > 0 {
				dst[0] = 0
			}
			return 0
		}
		negative = 1
		dst[0] = '-'
		dst = dst[1:]
		dstlen--
	} else {
		value = uint64(svalue)
	}

	length := FormatUint6410(dst, value)
	if length == 0 {
		return 0
	}

	return length + negative
}

func ParseUint64(s string) (uint64, error) {
	var v uint64
	if len(s) == 0 {
		return 0, ErrEmptyString
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidCharacter
		}

		d := uint64(c - '0')
		if v > (math.MaxUint64-d)/10 {
			return 0, ErrOverflow
		}
		v = v*10 + d
	}
	return v, nil
}

func ParseInt64(s string) (int64, error) {
	if len(s) == 0 {
		return 0, ErrEmptyString
	}
	negative := false
	start := 0
	if s[0] == '-' {
		negative = true
		start = 1
		if len(s) == 1 {
			return 0, ErrInvalidString
		}
	}
	var v uint64
	for i := start; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidCharacter
		}
		d := uint64(c - '0')
		if v > (math.MaxInt64-uint64(d))/10 {
			if negative && v == (math.MaxInt64+1-d)/10 {
				v = v*10 + d
				break
			}
			return 0, ErrOverflow
		}
		v = v*10 + d
	}
	if negative {
		if v > math.MaxInt64+1 {
			return 0, ErrOverflow
		}
		return -int64(v), nil
	}
	if v > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(v), nil
}

func Bool2int(x bool) int {
	return int(*(*uint8)(unsafe.Pointer(&x)))
}

func FormatUint16(dst []byte, value uint16) int {
	dstlen := len(dst)
	if dstlen == 0 {
		return 0
	}
	if value == 0 {
		dst[0] = '0'
		return 1
	}
	length := 1
	v := value
	for v >= 10 {
		v /= 10
		length++
	}
	if int(length) > dstlen {
		dst[0] = 0
		return 0
	}
	next := length - 1
	for value >= 100 {
		i := int(value%100) * 2
		value /= 100
		dst[next] = digits[i+1]
		dst[next-1] = digits[i]
		next -= 2
	}
	if value < 10 {
		dst[next] = '0' + byte(value)
	} else {
		i := int(value) * 2
		dst[next] = digits[i+1]
		dst[next-1] = digits[i]
	}
	return length
}
