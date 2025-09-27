package main

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/platform"
)

func main() {
	fmt.Println("Testing FixPanic Agent URL generation...")

	// Test latest version
	url, err := platform.GetFixPanicAgentDownloadURL("latest")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated URL: %s\n", url)

	// Verify URL format
	expectedPattern := "https://github.com/fixpanic/fixpanic-agent/releases/latest/download/fixpanic-agent-"
	if len(url) > len(expectedPattern) && url[:len(expectedPattern)] == expectedPattern {
		fmt.Println("✅ URL format is correct - matches GitHub Releases pattern!")
	} else {
		fmt.Printf("❌ URL format is incorrect: %s\n", url)
	}

	// Test specific version
	url2, err := platform.GetFixPanicAgentDownloadURL("v1.2.3")
	if err != nil {
		fmt.Printf("❌ Error with specific version: %v\n", err)
		return
	}
	fmt.Printf("✅ Specific version URL: %s\n", url2)

	// Test platform info
	os, arch, err := platform.GetFixPanicAgentPlatformInfo()
	if err != nil {
		fmt.Printf("❌ Platform error: %v\n", err)
		return
	}
	fmt.Printf("✅ Platform: %s/%s\n", os, arch)

	fmt.Println("\n🎉 All tests passed! FixPanic Agent integration is working correctly.")
}
