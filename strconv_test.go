package strconv2

import (
	"math"
	"strconv"
	"testing"
	"unsafe"
)

var sink string
var _sink string
var sinkUint64 uint64
var sinkInt64 int64
var sinkErr error

func _string(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func BenchmarkStrconv2FormatUint64(b *testing.B) {
	var buf [SAFETY_BUF_SIZE]byte
	for i := 0; i < b.N; i++ {
		n := FormatUint6410(buf[:], 1234567890123456789)
		_sink = _string(buf[:n])
	}
}

func BenchmarkStrconvFormatUint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_sink = strconv.FormatUint(1234567890123456789, 10)
	}
}

func BenchmarkStrconv2FormatInt(b *testing.B) {
	var buf [SAFETY_BUF_SIZE]byte
	for i := 0; i < b.N; i++ {
		n := FormatInt6410(buf[:], -1234567890123456789)
		_sink = _string(buf[:n])
	}
}

func BenchmarkStrconvInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_sink = strconv.FormatInt(-1234567890123456789, 10)
	}
}

func BenchmarkStrconv2FormatUint16(b *testing.B) {
	var buf [UINT16_MAX]byte
	for i := 0; i < b.N; i++ {
		n := FormatUint16(buf[:], 12345)
		_sink = _string(buf[:n])
	}
}

func BenchmarkStrconvFormatUint16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_sink = strconv.FormatUint(12345, 10)
	}
}

func BenchmarkStrconv2ParseUint64(b *testing.B) {
	s := "1234567890123456789"
	for i := 0; i < b.N; i++ {
		sinkUint64, sinkErr = ParseUint64(s)
	}
}
func BenchmarkStrconvParseUint64(b *testing.B) {
	s := "1234567890123456789"
	for i := 0; i < b.N; i++ {
		sinkUint64, sinkErr = strconv.ParseUint(s, 10, 64)
	}
}

func BenchmarkStrconv2ParseInt64(b *testing.B) {
	s := "-1234567890123456789"
	for i := 0; i < b.N; i++ {
		sinkInt64, sinkErr = ParseInt64(s)
	}
}

func BenchmarkStrconvParseInt64(b *testing.B) {
	s := "-1234567890123456789"
	for i := 0; i < b.N; i++ {
		sinkInt64, sinkErr = strconv.ParseInt(s, 10, 64)
	}
}

func TestDigits10(t *testing.T) {
	tests := []struct {
		value    uint64
		expected uint32
	}{
		{0, 1},
		{5, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1000, 4},
		{9999, 4},
		{10000, 5},
		{99999, 5},
		{100000, 6},
		{999999, 6},
		{1_000_000, 7},
		{9_999_999, 7},
		{10_000_000, 8},
		{99_999_999, 8},
		{100_000_000, 9},
		{999_999_999, 9},
		{1_000_000_000, 10},
		{9_999_999_999, 10},
		{10_000_000_000, 11},
		{99_999_999_999, 11},
		{100_000_000_000, 12},
		{999_999_999_999, 12},
		{1_000_000_000_000, 13},
		{math.MaxUint64, 20},
	}

	for _, tt := range tests {
		got := Digits10(tt.value)
		if got != tt.expected {
			t.Errorf("Digits10(%d) = %d, want %d", tt.value, got, tt.expected)
		}
	}
}

func TestFormatUint16(t *testing.T) {
	tests := []struct {
		value    uint16
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{99, "99"},
		{100, "100"},
		{12345, "12345"},
		{65535, "65535"},
	}

	for _, tt := range tests {
		buf := make([]byte, UINT16_MAX)
		n := FormatUint16(buf, tt.value)
		got := string(buf[:n])
		if got != tt.expected {
			t.Errorf("FormatUint16(%d) = %q, want %q", tt.value, got, tt.expected)
		}
	}
}

func TestParseUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		wantErr  bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"1234567890", 1234567890, false},
		{"18446744073709551615", math.MaxUint64, false},
		{"18446744073709551616", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseUint64(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseUint64(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ParseUint64(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"1234567890", 1234567890, false},
		{"-1234567890", -1234567890, false},
		{"9223372036854775807", math.MaxInt64, false},
		{"-9223372036854775808", math.MinInt64, false},
		{"9223372036854775808", 0, true},
		{"-9223372036854775809", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseInt64(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseInt64(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ParseInt64(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestUll2String(t *testing.T) {
	cases := []uint64{
		0, 1, 9, 10, 11, 99, 100, 9999,
		123456789, 1000000000000, 18446744073709551615,
	}

	for _, v := range cases {
		buf := make([]byte, 32)
		n := FormatUint6410(buf, v)
		got := string(buf[:n])
		want := strconv.FormatUint(v, 10)
		if got != want {
			t.Fatalf("Ull2String(%d) = %q, want %q", v, got, want)
		}
	}
}
