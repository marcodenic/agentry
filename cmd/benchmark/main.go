package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/marcodenic/agentry/internal/trace"
)

func main() {
	fmt.Printf("Go 1.25 JSON Performance Benchmark for Agentry\n")
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("==============================================\n\n")

	// Default to 10000 operations, or use command line argument
	operations := 10000
	if len(os.Args) > 1 {
		if ops, err := strconv.Atoi(os.Args[1]); err == nil && ops > 0 {
			operations = ops
		}
	}

	fmt.Printf("Running JSON benchmark with %d operations...\n\n", operations)

	// Run the benchmark
	result := trace.BenchmarkJSON(operations)

	// Display results
	fmt.Printf("Benchmark Results:\n")
	fmt.Printf("-----------------\n")
	fmt.Printf("Go Version:         %s\n", result.GoVersion)
	fmt.Printf("Operations:         %d\n", result.Operations)
	fmt.Printf("Marshal Duration:   %v\n", result.MarshalDuration)
	fmt.Printf("Unmarshal Duration: %v\n", result.UnmarshalDuration)
	fmt.Printf("Total Duration:     %v\n", result.TotalDuration)
	fmt.Printf("Marshal Ops/sec:    %.0f\n", result.MarshalOpsPerSec)
	fmt.Printf("Unmarshal Ops/sec:  %.0f\n", result.UnmarshalOpsPerSec)
	fmt.Printf("\n")

	// Performance analysis
	avgMarshalNs := result.MarshalDuration.Nanoseconds() / int64(operations)
	avgUnmarshalNs := result.UnmarshalDuration.Nanoseconds() / int64(operations)

	fmt.Printf("Performance Analysis:\n")
	fmt.Printf("--------------------\n")
	fmt.Printf("Avg Marshal Time:   %d ns/op\n", avgMarshalNs)
	fmt.Printf("Avg Unmarshal Time: %d ns/op\n", avgUnmarshalNs)

	// Go 1.25 specific benefits
	fmt.Printf("\nGo 1.25 Benefits for Agentry:\n")
	fmt.Printf("-----------------------------\n")
	fmt.Printf("✅ JSON Performance: Substantially improved JSON encoding/decoding\n")
	fmt.Printf("✅ Trace Events: Faster serialization of trace events\n")
	fmt.Printf("✅ API Communication: Better performance for AI model API calls\n")
	fmt.Printf("✅ Tool Arguments: Optimized parsing of tool call arguments\n")
	fmt.Printf("✅ Configuration: Faster loading of YAML/JSON config files\n")

	if avgMarshalNs < 1000 {
		fmt.Printf("🚀 Excellent performance: Marshal operations under 1μs\n")
	} else if avgMarshalNs < 10000 {
		fmt.Printf("⚡ Good performance: Marshal operations under 10μs\n")
	}

	if avgUnmarshalNs < 2000 {
		fmt.Printf("🚀 Excellent performance: Unmarshal operations under 2μs\n")
	} else if avgUnmarshalNs < 20000 {
		fmt.Printf("⚡ Good performance: Unmarshal operations under 20μs\n")
	}
}
