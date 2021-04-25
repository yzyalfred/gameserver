package log

import (
	"testing"
)

func Benchmark_LogInfo(b *testing.B)  {
	InitLog("../log", "debug", false, 0)
	for i := 0; i < b.N; i++ {
		Info("test abc: ", "yzy")
	}
}

func Benchmark_LogError(b *testing.B)  {
	InitLog("../log", "debug", false, 0)
	for i := 0; i < b.N; i++ {
		Error("test abc: ", "yzy")
	}
}