// build.go - Multi-platform build script for the watcher application
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Supported platforms for cross-compilation
var platforms = []struct {
	os   string
	arch string
}{
	{"windows", "amd64"},
	{"windows", "386"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "386"},
	{"linux", "arm"},
	{"linux", "arm64"},
}

func main() {
	// Parse command-line flags
	outputDir := flag.String("output", "dist", "Output directory for binaries")
	version := flag.String("version", "1.0.0", "Version number for the build")
	onlyCurrentPlatform := flag.Bool("current", false, "Build only for the current platform")
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Get the package name from go.mod
	packageName, err := getPackageName()
	if err != nil {
		fmt.Printf("Error getting package name: %v\n", err)
		os.Exit(1)
	}

	// Build for specified platforms
	if *onlyCurrentPlatform {
		// Build only for current platform
		if err := buildBinary(runtime.GOOS, runtime.GOARCH, *outputDir, packageName, *version); err != nil {
			fmt.Printf("Error building for current platform: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Build for all supported platforms
		for _, platform := range platforms {
			if err := buildBinary(platform.os, platform.arch, *outputDir, packageName, *version); err != nil {
				fmt.Printf("Error building for %s/%s: %v\n", platform.os, platform.arch, err)
				continue
			}
		}
	}

	fmt.Println("Build completed successfully!")
}

// getPackageName extracts the package name from go.mod
func getPackageName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty go.mod file")
	}

	// Extract module name from the first line
	// Format: module github.com/username/repo
	parts := strings.Fields(lines[0])
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid go.mod format")
	}

	modulePath := parts[1]
	// Get the last part of the module path as the binary name
	return filepath.Base(modulePath), nil
}

// buildBinary builds the application for a specific platform
func buildBinary(goos, goarch, outputDir, packageName, version string) error {
	// Set binary name based on platform
	binaryName := packageName
	if goos == "windows" {
		binaryName += ".exe"
	}

	// Create platform-specific output directory
	platformDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", goos, goarch))
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(platformDir, binaryName)
	fmt.Printf("Building for %s/%s: %s\n", goos, goarch, outputPath)

	// Set up environment for cross-compilation
	cmd := exec.Command("go", "build", "-o", outputPath, "-ldflags", fmt.Sprintf("-X main.Version=%s", version))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", goos),
		fmt.Sprintf("GOARCH=%s", goarch),
	)

	// Run the build command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\n%s", err, output)
	}

	fmt.Printf("Successfully built %s\n", outputPath)
	return nil
}