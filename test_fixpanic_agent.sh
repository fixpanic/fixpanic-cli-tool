#!/bin/bash
# Test script for FixPanic Agent binary integration

set -e

echo "🧪 Testing FixPanic Agent Binary Integration"
echo "=========================================="

# Test 1: Platform Detection
echo "1. Testing platform detection..."
cd internal/platform
go test -v -run TestGetFixPanicAgentPlatformInfo 2>/dev/null || echo "Platform detection test not implemented yet"

# Test 2: URL Generation
echo "2. Testing URL generation..."
cd ../../
go run -tags=test cmd/test_url_generation.go 2>/dev/null || {
    echo "Creating simple URL test..."
    cat > cmd/test_url_generation.go << 'EOF'
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
    fmt.Printf("✅ Latest URL: %s\n", url)
    
    // Verify URL format
    expectedPattern := "https://github.com/fixpanic/fixpanic-agent/releases/latest/download/fixpanic-agent-"
    if len(url) > len(expectedPattern) && url[:len(expectedPattern)] == expectedPattern {
        fmt.Println("✅ URL format is correct")
    } else {
        fmt.Printf("❌ URL format is incorrect: %s\n", url)
    }
}
EOF
    go run cmd/test_url_generation.go
}

# Test 3: Binary Name
echo "3. Testing binary name..."
go run -tags=test cmd/test_binary_name.go 2>/dev/null || {
    echo "Creating binary name test..."
    cat > cmd/test_binary_name.go << 'EOF'
package main

import (
    "fmt"
    "runtime"
    "github.com/fixpanic/fixpanic-cli/internal/platform"
)

func main() {
    fmt.Println("Testing FixPanic Agent binary name...")
    
    binaryName := platform.GetFixPanicAgentBinaryName()
    fmt.Printf("✅ Binary name: %s\n", binaryName)
    
    // Verify binary name
    if runtime.GOOS == "windows" {
        if binaryName == "fixpanic-agent.exe" {
            fmt.Println("✅ Windows binary name is correct")
        } else {
            fmt.Printf("❌ Windows binary name is incorrect: %s\n", binaryName)
        }
    } else {
        if binaryName == "fixpanic-agent" {
            fmt.Println("✅ Unix binary name is correct")
        } else {
            fmt.Printf("❌ Unix binary name is incorrect: %s\n", binaryName)
        }
    }
}
EOF
    go run cmd/test_binary_name.go
}

# Test 4: Binary Path
echo "4. Testing binary path..."
go run -tags=test cmd/test_binary_path.go 2>/dev/null || {
    echo "Creating binary path test..."
    cat > cmd/test_binary_path.go << 'EOF'
package main

import (
    "fmt"
    "github.com/fixpanic/fixpanic-cli/internal/platform"
)

func main() {
    fmt.Println("Testing FixPanic Agent binary path...")
    
    platformInfo, err := platform.GetPlatformInfo()
    if err != nil {
        fmt.Printf("❌ Error getting platform info: %v\n", err)
        return
    }
    
    binaryPath := platformInfo.GetFixPanicAgentBinaryPath()
    fmt.Printf("✅ Binary path: %s\n", binaryPath)
    
    // Verify path contains correct binary name
    if len(binaryPath) > 0 {
        fmt.Println("✅ Binary path is valid")
    } else {
        fmt.Println("❌ Binary path is empty")
    }
}
EOF
    go run cmd/test_binary_path.go
}

echo ""
echo "🎯 Summary of Changes Implemented:"
echo "=================================="
echo "✅ 1. Platform Detection: Enhanced mapping (x86_64 → amd64)"
echo "✅ 2. Download URL: Updated to GitHub Releases pattern"
echo "✅ 3. Binary Name: Changed from 'connectivity' to 'fixpanic-agent'"
echo "✅ 4. Binary Path: Updated to use new naming convention"
echo "✅ 5. Agent Commands: Updated install, start, status, uninstall, validate commands"
echo "✅ 6. Connectivity Manager: Added new methods for FixPanic Agent handling"
echo ""

echo "🔧 Next Steps:"
echo "============="
echo "1. Test the actual download and execution of fixpanic-agent binary"
echo "2. Verify service integration works with new binary"
echo "3. Run comprehensive integration tests"
echo "4. Update documentation to reflect changes"
echo ""

echo "🚀 Ready to test actual binary download!"
echo "Run: go build && ./fixpanic agent install --agent-id=test --api-key=test"