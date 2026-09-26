package strconv2

import (
	"strconv"
	"testing"
)

func TestFormatUintExhaustiveVsStd(t *testing.T) {
	var buf [32]byte
	for v := uint64(0); v <= 2_000_000; v++ {
		n := FormatUint6410(buf[:], v)
		if string(buf[:n]) != strconv.FormatUint(v, 10) {
			t.Fatalf("v=%d got %q", v, buf[:n])
		}
	}
	edge := []uint64{999, 1000, 1001, 999999, 1000000, 1<<32 - 1, 1 << 32, 1e18, 9999999999999999999, 18446744073709551615}
	for _, v := range edge {
		n := FormatUint6410(buf[:], v)
		if string(buf[:n]) != strconv.FormatUint(v, 10) {
			t.Fatalf("edge v=%d got %q", v, buf[:n])
		}
	}
}

func TestFormatIntExhaustiveVsStd(t *testing.T) {
	var buf [32]byte
	for x := int64(-2_000_000); x <= 2_000_000; x++ {
		n := FormatInt6410(buf[:], x)
		if string(buf[:n]) != strconv.FormatInt(x, 10) {
			t.Fatalf("x=%d got %q", x, buf[:n])
		}
	}
}

func TestParseInt64VsStd(t *testing.T) {
	cases := []string{
		"0", "-0", "+0", "1", "+1", "-1", "123", "+123", "-123",
		"9223372036854775807", "+9223372036854775807", "-9223372036854775808",
		"9223372036854775808", "-9223372036854775809", "+9223372036854775808",
		"-0000000000000000000000009223372036854775808",
		"00000000000000000000000123",
		"", "+", "-", "12a3", "12-3", " 12", "12 ", "++1", "--1", "+-1",
	}
	for _, s := range cases {
		got, gerr := ParseInt64(s)
		want, werr := strconv.ParseInt(s, 10, 64)
		if (gerr != nil) != (werr != nil) {
			t.Errorf("ParseInt64(%q): err=%v, std err=%v", s, gerr, werr)
			continue
		}
		if gerr == nil && got != want {
			t.Errorf("ParseInt64(%q) = %d, want %d", s, got, want)
		}
	}
	for x := int64(-3_000_000); x <= 3_000_000; x++ {
		s := strconv.FormatInt(x, 10)
		got, err := ParseInt64(s)
		if err != nil || got != x {
			t.Fatalf("ParseInt64(%q) = %d, %v", s, got, err)
		}
	}
}

func TestItoaVsStd(t *testing.T) {
	for x := -100000; x <= 100000; x++ {
		if Itoa(x) != strconv.Itoa(x) {
			t.Fatalf("Itoa(%d) = %q, want %q", x, Itoa(x), strconv.Itoa(x))
		}
	}
	for _, x := range []int{0, 9, 10, 99, 100, -1, -99, -100, 1<<62, -(1<<62)} {
		if Itoa(x) != strconv.Itoa(x) {
			t.Fatalf("Itoa(%d) = %q, want %q", x, Itoa(x), strconv.Itoa(x))
		}
	}
}
