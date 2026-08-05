# strconv2

Fast, zero-allocation integer <-> string conversion for Go. A focused drop-in for
the hot parts of the standard `strconv`, using branch-lean digit tables and SWAR
parsing.

```
go get github.com/NikoMalik/strconv2
```

## API

### Formatting (int -> bytes)

```go
func FormatUint6410(dst []byte, value uint64) int
func FormatInt6410(dst []byte, svalue int64) int
func FormatUint16(dst []byte, value uint16) int
```

All write ASCII decimal digits into the caller-provided `dst` and return the number
of bytes written. Nothing is allocated. If `dst` is too small the function writes a
single `0` sentinel byte at `dst[0]` (when there is room) and returns `0`.

```go
var buf [32]byte
n := FormatInt6410(buf[:], -1234567890)
s := string(buf[:n]) // "-1234567890"
```

### Parsing (string -> int)

```go
func ParseUint64(s string) (uint64, error)
func ParseInt64(s string) (int64, error)
```

```go
v, err := ParseUint64("18446744073709551615") // MaxUint64
n, err := ParseInt64("-9223372036854775808")  // MinInt64
```

### Helpers

```go
func Digits10(v uint64) uint32 // decimal digit count of v, Digits10(0) == 1
func Bool2int(x bool) int      // branchless bool -> 0/1
```

## Implementation notes

- **Zero allocations.** Formatting writes into a caller buffer, parsing reads the
  input string directly. No intermediate slices or strings.
- **SWAR parsing.** `ParseUint64`/`ParseInt64` consume 8 digits per iteration by
  loading 8 bytes into a `uint64` and validating plus folding them with masked
  arithmetic (`parse8DigitsSWAR`), instead of one branch per byte.
- **Two-digit lookup table.** Formatting emits digits two at a time from a 200-byte
  table, halving the divisions and stores versus a per-digit loop.
- **Fast / slow parse split.** Inputs up to 19 digits cannot overflow `uint64`
  (max 19-digit value < `MaxUint64`), so they take a check-free fast path. Inputs of
  20+ digits fall back to a per-digit path with overflow checks. This also correctly
  handles long inputs that are valid only because of leading zeros.
- **Branch-lean digit count.** `Digits10` uses a nested comparison tree plus
  `Bool2int` instead of a loop.

## Behavior and limitations

Read this before swapping in for `strconv`.

- **No sign prefix on input.** Parsing does **not** accept a leading `+`, and
  `ParseUint64` does not accept a leading `-`.
  - `ParseUint64("+5")` -> `ErrInvalidCharacter`
  - `ParseUint64("-5")` -> `ErrInvalidCharacter`
  - `ParseInt64("+5")`  -> `ErrInvalidCharacter`
  - `ParseInt64("-5")`  -> `-5` (only `-` is handled, and only for signed parsing)

  This differs from the standard library, where both `ParseInt` and `ParseUint`
  accept a leading `+`. If your input can carry an explicit `+`, strip it first.
- **Base 10 only.** No base argument, no `0x` / `0b` / `0o` prefixes, no underscores.
- **Leading zeros are allowed.** `"007"` -> `7`, `"-000123"` -> `-123`.
- **Empty string** -> `ErrEmptyString`. A lone `"-"` -> `ErrInvalidString`.
- **Buffer sizing for formatting.** Provide enough room or you get `0` back:
  - `uint64`: up to 20 bytes
  - `int64`: up to 20 bytes (19 digits + sign)
  - `uint16`: up to 5 bytes (`Uint16MaxDigits`)
- **`unsafe` is used** in the SWAR path to load 8 bytes at once. Reads are bounded by
  the input length (`i+8 <= n`), so they never cross the string's backing array.

### Errors

| Error | When |
|-------|------|
| `ErrEmptyString` | input is `""` |
| `ErrInvalidString` | input is only a sign, e.g. `"-"` |
| `ErrInvalidCharacter` | a non-digit byte (including `+`/`-` where not allowed) |
| `ErrOverflow` | value does not fit the target type |

## Benchmarks

`go version go1.26`, amd64, 12 threads. `go test -bench=. -benchmem`.

| Operation | strconv2 | strconv | allocs (strconv2 / strconv) |
|-----------|----------|---------|-----------------------------|
| Format uint64      | 24.8 ns | 49.1 ns (`FormatUint`) | 0 / 1 |
| Format int64       | 27.8 ns | 49.9 ns (`FormatInt`)  | 0 / 1 |
| Format uint16      | 10.1 ns | 23.9 ns (`FormatUint`) | 0 / 1 |
| Parse uint64       | 17.7 ns | 51.3 ns (`ParseUint`)  | 0 / 0 |
| Parse int64        | 19.3 ns | 55.1 ns (`ParseInt`)   | 0 / 0 |

Numbers are indicative and vary by machine and input length.

## Testing

```
go test ./...
```

Formatting and parsing are fuzz-checked against the standard `strconv` across the
full value range, including `0`, `MaxUint64`, `MaxInt64`, and `MinInt64`.

## License

See [LICENSE](LICENSE).
