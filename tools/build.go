// build.go - Multi-platform build script for the watcher application
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
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
	createReleaseArchives := flag.Bool("release", false, "Create release-ready zip archives")
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
	var builtBinaries []string
	if *onlyCurrentPlatform {
		// Build only for current platform
		binaryPath, err := buildBinary(runtime.GOOS, runtime.GOARCH, *outputDir, packageName, *version)
		if err != nil {
			fmt.Printf("Error building for current platform: %v\n", err)
			os.Exit(1)
		}
		builtBinaries = append(builtBinaries, binaryPath)
	} else {
		// Build for all supported platforms
		for _, platform := range platforms {
			binaryPath, err := buildBinary(platform.os, platform.arch, *outputDir, packageName, *version)
			if err != nil {
				fmt.Printf("Error building for %s/%s: %v\n", platform.os, platform.arch, err)
				continue
			}
			builtBinaries = append(builtBinaries, binaryPath)
		}
	}

	fmt.Println("Build completed successfully!")

	// Create release archives if requested
	if *createReleaseArchives {
		releaseDir := filepath.Join(*outputDir, "release")
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			fmt.Printf("Error creating release directory: %v\n", err)
			os.Exit(1)
		}

		for _, binaryPath := range builtBinaries {
			if err := createReleaseArchive(binaryPath, releaseDir, packageName, *version); err != nil {
				fmt.Printf("Error creating release archive for %s: %v\n", binaryPath, err)
				continue
			}
		}

		fmt.Printf("\nRelease archives created in %s\n", releaseDir)
		fmt.Println("You can now upload these archives to GitHub releases:")
		fmt.Println("1. Go to your GitHub repository")
		fmt.Println("2. Click on 'Releases'")
		fmt.Println("3. Click on 'Draft a new release'")
		fmt.Println("4. Fill in the release details")
		fmt.Println("5. Upload the zip files from the release directory")
		fmt.Println("6. Click 'Publish release'")
	}
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
func buildBinary(goos, goarch, outputDir, packageName, version string) (string, error) {
	// Set binary name based on platform
	binaryName := packageName
	if goos == "windows" {
		binaryName += ".exe"
	}

	// Create platform-specific output directory
	platformDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", goos, goarch))
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return "", err
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
		return "", fmt.Errorf("build failed: %v\n%s", err, output)
	}

	fmt.Printf("Successfully built %s\n", outputPath)
	return outputPath, nil
}

// createReleaseArchive creates a zip archive for a binary
func createReleaseArchive(binaryPath, releaseDir, packageName, version string) error {
	// Get platform info from binary path
	platformDir := filepath.Dir(binaryPath)
	platform := filepath.Base(platformDir)
	
	// Create zip file name
	zipName := fmt.Sprintf("%s-%s-%s.zip", packageName, version, platform)
	zipPath := filepath.Join(releaseDir, zipName)
	
	fmt.Printf("Creating release archive: %s\n", zipPath)
	
	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error creating zip file: %v", err)
	}
	defer zipFile.Close()
	
	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	// Add binary to zip
	binaryFile, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("error opening binary file: %v", err)
	}
	defer binaryFile.Close()
	
	// Get file info
	binaryInfo, err := binaryFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting binary file info: %v", err)
	}
	
	// Create zip file header
	header, err := zip.FileInfoHeader(binaryInfo)
	if err != nil {
		return fmt.Errorf("error creating zip file header: %v", err)
	}
	
	// Set compression
	header.Method = zip.Deflate
	
	// Add file to zip
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error creating zip file writer: %v", err)
	}
	
	// Copy binary to zip
	if _, err := io.Copy(writer, binaryFile); err != nil {
		return fmt.Errorf("error copying binary to zip: %v", err)
	}
	
	// Add README.md to zip
	if err := addFileToZip(zipWriter, "README.md"); err != nil {
		fmt.Printf("Warning: Could not add README.md to zip: %v\n", err)
	}
	
	// Add config.json to zip
	if err := addFileToZip(zipWriter, "config.json"); err != nil {
		fmt.Printf("Warning: Could not add config.json to zip: %v\n", err)
	}
	
	// Add LICENSE to zip if it exists
	if _, err := os.Stat("LICENSE"); err == nil {
		if err := addFileToZip(zipWriter, "LICENSE"); err != nil {
			fmt.Printf("Warning: Could not add LICENSE to zip: %v\n", err)
		}
	}
	
	fmt.Printf("Successfully created %s\n", zipPath)
	return nil
}

// addFileToZip adds a file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	info, err := file.Stat()
	if err != nil {
		return err
	}
	
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	
	header.Method = zip.Deflate
	
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	
	_, err = io.Copy(writer, file)
	return err
}