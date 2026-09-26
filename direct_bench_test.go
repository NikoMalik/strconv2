package strconv2

import "testing"

func BenchmarkDirectBranchless_randSmall(b *testing.B) {
	ss := rndSmall
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range ss {
			sinkInt64, sinkErr = ParseInt64(s)
		}
	}
}

func BenchmarkDirectBranchy_randSmall(b *testing.B) {
	ss := rndSmall
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range ss {
			sinkInt64, sinkErr = parseInt64Branchy(s)
		}
	}
}

func BenchmarkDirectBranchless_randLarge(b *testing.B) {
	ss := rndLarge
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range ss {
			sinkInt64, sinkErr = ParseInt64(s)
		}
	}
}

func BenchmarkDirectBranchy_randLarge(b *testing.B) {
	ss := rndLarge
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range ss {
			sinkInt64, sinkErr = parseInt64Branchy(s)
		}
	}
}
