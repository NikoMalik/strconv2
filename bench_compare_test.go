package strconv2

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// Deprecated: old 2-digit-per-iteration formatter, kept for benchmarking against
// the 3-digit FormatUint6410
func formatUint2Digit(dst []byte, value uint64) int {
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

// Deprecated: branchy sign handling, kept for benchmarking against the branchless
// ParseInt64, on direct calls branchless is tied on small ints and ~5% faster on
// large ones
func parseInt64Branchy(s string) (int64, error) {
	if len(s) == 0 {
		return 0, ErrEmptyString
	}
	first := s[0]
	negative := first == '-'
	start := 0
	if negative || first == '+' {
		start = 1
	}
	if start >= len(s) {
		return 0, ErrInvalidString
	}
	digitStr := s[start:]
	var v uint64
	var err error
	if len(digitStr) <= 19 {
		v, err = parseUint64Fast(digitStr)
	} else {
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

func rangeVals(lo, hi uint64) []uint64 {
	v := make([]uint64, 1024)
	step := (hi - lo) / 1024
	if step == 0 {
		step = 1
	}
	x := lo
	for i := range v {
		v[i] = x
		x += step
	}
	return v
}

var digitSink int

func benchFormat(b *testing.B, f func([]byte, uint64) int, vals []uint64) {
	var buf [32]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range vals {
			digitSink += f(buf[:], v)
		}
	}
}

var (
	vSmall = rangeVals(0, 999)
	vMid   = rangeVals(1000, 999999)
	vLarge = rangeVals(1e6, 1e12)
	vHuge  = rangeVals(1e12, math.MaxUint64)
)

func BenchmarkFmt2Digit_small(b *testing.B) { benchFormat(b, formatUint2Digit, vSmall) }
func BenchmarkFmt3Digit_small(b *testing.B) { benchFormat(b, FormatUint6410, vSmall) }
func BenchmarkFmt2Digit_mid(b *testing.B)   { benchFormat(b, formatUint2Digit, vMid) }
func BenchmarkFmt3Digit_mid(b *testing.B)   { benchFormat(b, FormatUint6410, vMid) }
func BenchmarkFmt2Digit_large(b *testing.B) { benchFormat(b, formatUint2Digit, vLarge) }
func BenchmarkFmt3Digit_large(b *testing.B) { benchFormat(b, FormatUint6410, vLarge) }
func BenchmarkFmt2Digit_huge(b *testing.B)  { benchFormat(b, formatUint2Digit, vHuge) }
func BenchmarkFmt3Digit_huge(b *testing.B)  { benchFormat(b, FormatUint6410, vHuge) }

func randomSignStrs(seed int64, magLo, magHi int64) []string {
	r := rand.New(rand.NewSource(seed))
	s := make([]string, 1024)
	span := uint64(magHi - magLo)
	for i := range s {
		mag := magLo + int64(r.Uint64()%(span+1))
		if r.Uint64()&1 == 0 {
			mag = -mag
		}
		s[i] = strconv.FormatInt(mag, 10)
	}
	return s
}

var (
	rndSmall = randomSignStrs(1, 0, 999)
	rndLarge = randomSignStrs(2, 0, 999999999999)
)
