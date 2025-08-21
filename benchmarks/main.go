package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== MArrayCRDT vs Automerge: Comprehensive Performance Comparison ===")
	fmt.Println()
	fmt.Println("This will run MArrayCRDT simulations at all the same scales as Automerge:")
	fmt.Println("• 1k, 5k, 10k, 20k, 30k, 40k, 50k operations")  
	fmt.Println("• Using the exact same editing trace from Kleppmann et al.'s paper")
	fmt.Println("• Comparing against real Automerge benchmark results")
	fmt.Println()
	fmt.Println("Expected runtime: 10-15 minutes")
	fmt.Println("Results will be saved to ../data/comprehensive_performance_comparison.csv")
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
	fmt.Println("  • ../data/comprehensive_performance_comparison.csv - For plotting and analysis")
	fmt.Println("  • marraycrdt_comprehensive_benchmark.json - Detailed MArrayCRDT results")
	fmt.Println()
	fmt.Println("📈 The CSV contains performance data for:")
	fmt.Println("  • MArrayCRDT at 1k, 5k, 10k, 20k, 30k, 40k, 50k operations")
	fmt.Println("  • Automerge at matching scales (from real benchmarks)")
	fmt.Println("  • JavaScript Array baseline (259k operations)")
	fmt.Println()
	fmt.Println("Ready for visualization and analysis! 🎯")
}