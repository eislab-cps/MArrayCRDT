package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== MArrayCRDT Performance Benchmark ===")
	fmt.Println()
	fmt.Println("This will run MArrayCRDT simulations at multiple scales:")
	fmt.Println("• 1k, 5k, 10k, 20k, 30k, 40k, 50k operations")  
	fmt.Println("• Using the exact same editing trace from Kleppmann et al.'s paper")
	fmt.Println("• Measuring throughput and memory usage at each scale")
	fmt.Println()
	fmt.Println("Expected runtime: 10-15 minutes")
	fmt.Println("Results will be saved to ../simulation/marraycrdt_results.csv")
	fmt.Println()
	
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	
	// Run the comprehensive benchmark suite
	fmt.Println("🚀 Starting comprehensive benchmark suite...")
	if err := RunComprehensiveBenchmarks(); err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		return
	}
	
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("✅ Comprehensive comparison completed!")
	fmt.Println()
	fmt.Println("📊 Generated files:")
	fmt.Println("  • ../simulation/marraycrdt_results.csv - For plotting and analysis")
	fmt.Println("  • ../simulation/marraycrdt_comprehensive_benchmark.json - Detailed MArrayCRDT results")
	fmt.Println()
	fmt.Println("📈 The CSV contains performance data for:")
	fmt.Println("  • MArrayCRDT at 1k, 5k, 10k, 20k, 30k, 40k, 50k operations")
	fmt.Println("  • Detailed metrics: throughput, memory usage, operation counts")
	fmt.Println()
	fmt.Println("Ready for visualization and analysis! 🎯")
}