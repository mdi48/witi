package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const pacmanLocalDBPath = "/var/lib/pacman/local/"

type Package struct {
	Name          string
	Version       string
	InstallReason string
	Dependencies  []string
	RequiredBy    []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <package-name>")
		os.Exit(1)
	}

	pkgName := os.Args[1]
	fmt.Println("Fetching package information for:", pkgName)
	fmt.Println(getPackageInfo(pkgName)) // Testing (it works!)
}

// gets package info from pacman's local database
func getPackageInfo(pkgName string) (*Package, error) {
	// find the package dir (format: pkgname-version)
	pkgDir, err := findPackageDir(pkgName)
	if err != nil {
		return nil, err
	}

	// read the desc file in the package dir
	descFile := filepath.Join(pkgDir, "desc")
	pkg, err := parseDescFile(descFile)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

// finds the package director in /var/lib/pacman/local/
func findPackageDir(pkgName string) (string, error) {
	entries, err := os.ReadDir(pacmanLocalDBPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), pkgName+"-") {
			return filepath.Join(pacmanLocalDBPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("package directory not found for package: %s", pkgName)
}

// parses the desc file to extract package info
func parseDescFile(descFile string) (*Package, error) {
	file, err := os.Open(descFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pkg := &Package{
		// making InstallReason 0 by default to indicate explicitly installed (so as to align with pacman's behaviour). If 1 is found it changes
		InstallReason: "Explicitly installed",
	}
	scanner := bufio.NewScanner(file)
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Section headers are in the format "%SECTION%"
		if strings.HasPrefix(line, "%") && strings.HasSuffix(line, "%") {
			currentSection = strings.Trim(line, "%")
			continue
		}

		// skips empty lines
		if line == "" {
			continue
		}

		switch currentSection {
		case "NAME":
			pkg.Name = line
		case "VERSION":
			pkg.Version = line
		case "REASON":
			if line == "1" {
				pkg.InstallReason = "Installed as a dependency"
			}
		case "DEPENDS":
			pkg.Dependencies = append(pkg.Dependencies, line)
		}
	}

	return pkg, scanner.Err()
}

// IMPLEMENT HERE: function to find which packages require the given package (reverse dependencies)
