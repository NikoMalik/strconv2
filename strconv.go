package strconv2

import (
	"errors"
	"math"
	"unsafe"
)

const SAFETY_BUF_SIZE = 32
const FAST_BUF_SIZE = 24
const Uint16MaxDigits = 5

const cutoff = math.MaxUint64 / 10
const cutlim = math.MaxUint64 % 10

const cutoff_neg = uint64(math.MaxInt64) + 1
const cutoff_no_neg = uint64(math.MaxInt64)

const (
	cutoffNegDiv10 = cutoff_neg / 10
	cutoffNegMod10 = cutoff_neg % 10
	cutoffPosDiv10 = cutoff_no_neg / 10
	cutoffPosMod10 = cutoff_no_neg % 10
)

var (
	ErrOverflow         = errors.New("overflow")
	ErrInvalidCharacter = errors.New("invalid character")
	ErrInvalidString    = errors.New("invalid string")
	ErrEmptyString      = errors.New("empty string")
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
				return 5 + uint32(Bool2int(v >= 100_000))
			}
			return 7 + uint32(Bool2int(v >= 10_000_000))
		}
		if v < 10_000_000_000 {
			return 9 + uint32(Bool2int(v >= 1_000_000_000))
		}
		return 11 + uint32(Bool2int(v >= 100_000_000_000))
	}
	if v < 10_000_000_000_000_000 {
		if v < 100_000_000_000_000 {
			return 13 + uint32(Bool2int(v >= 10_000_000_000_000))
		}
		return 15 + uint32(Bool2int(v >= 1_000_000_000_000_000))
	}
	if v < 1_000_000_000_000_000_000 {
		return 17 + uint32(Bool2int(v >= 100_000_000_000_000_000))
	}
	return 19 + uint32(Bool2int(v >= 10_000_000_000_000_000_000))
}

func Itoa(v int) string {
	var buf [SAFETY_BUF_SIZE]byte
	n := FormatInt6410(buf[:], int64(v))
	return _string(buf[:n])
}

func _string(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func FormatUint6410(dst []byte, value uint64) int {
	dstlen := len(dst)

	length := Digits10(value)
	if int(length) > dstlen {
		if dstlen > 0 {
			dst[0] = 0
		}
		return 0
	}
	next := length - 1

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
	} else {
		value = uint64(svalue)
	}

	length := FormatUint6410(dst, value)
	if length == 0 {
		return 0
	}

	return length + negative
}

// parse8DigitsSWAR parses exactly 8 ASCII digit bytes from s at the given offset
func parse8DigitsSWAR(s string, offset int) (uint64, bool) {
	p := unsafe.Add(unsafe.Pointer(unsafe.StringData(s)), offset)
	loaded := *(*uint64)(p)

	// Validate: upper nibble of every byte must be 0x3 (ASCII '0'-'9' are 0x30-0x39)
	if (loaded & 0xF0F0F0F0F0F0F0F0) != 0x3030303030303030 {
		return 0, false
	}
	// Validate: lower nibble of every byte must be < 10
	low := loaded & 0x0F0F0F0F0F0F0F0F
	if (low+0x0606060606060606)&0xF0F0F0F0F0F0F0F0 != 0 {
		return 0, false
	}

	// Combine pairs of digits into 2-digit numbers in 16-bit slots
	// On little-endian: byte 0 (most significant digit) is in the least significant byte
	lo := low & 0x00FF00FF00FF00FF
	hi := (low >> 8) & 0x00FF00FF00FF00FF
	val := lo*10 + hi

	// Combine pairs of 2-digit numbers into 4-digit numbers in 32-bit slots
	lo = val & 0x0000FFFF0000FFFF
	hi = (val >> 16) & 0x0000FFFF0000FFFF
	val = lo*100 + hi

	// Combine into final 8-digit number
	lo = val & 0x00000000FFFFFFFF
	hi = val >> 32
	return lo*10000 + hi, true
}

func ParseUint64(s string) (uint64, error) {
	n := len(s)
	if n == 0 {
		return 0, ErrEmptyString
	}

	// Fast path: <= 19 digits can never overflow uint64
	// (max 19-digit number = 9_999_999_999_999_999_999 < MaxUint64)
	if n <= 19 {
		return parseUint64Fast(s)
	}

	// Slow path: per-digit overflow checking (handles 20+ digits and leading zeros)
	return parseUint64Slow(s)
}

func parseUint64Fast(s string) (uint64, error) {
	var v uint64
	i := 0
	n := len(s)

	for i+8 <= n {
		chunk, ok := parse8DigitsSWAR(s, i)
		if !ok {
			return 0, ErrInvalidCharacter
		}
		v = v*100_000_000 + chunk
		i += 8
	}

	for ; i < n; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidCharacter
		}
		v = v*10 + uint64(c-'0')
	}
	return v, nil
}

func parseUint64Slow(s string) (uint64, error) {
	var v uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidCharacter
		}
		d := uint64(c - '0')
		if v > cutoff || (v == cutoff && d > cutlim) {
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

	digitStr := s[start:]
	digitLen := len(digitStr)

	var v uint64
	var err error

	// Max absolute value for int64 is 19 digits (9223372036854775808).
	// Any 19-digit number fits in uint64, so SWAR is safe.
	if digitLen <= 19 {
		v, err = parseUint64Fast(digitStr)
	} else {
		// 20+ digits: only valid with leading zeros. Scalar handles overflow.
		v, err = parseUint64Slow(digitStr)
	}
	if err != nil {
		return 0, err
	}

	if negative {
		if v > cutoff_neg {
			return 0, ErrOverflow
		}
		if v == cutoff_neg {
			return math.MinInt64, nil
		}
		return -int64(v), nil
	}
	if v > cutoff_no_neg {
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
