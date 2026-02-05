package main

import (
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
	fmt.Println(findPackageDir(pkgName)) // Testing (it works!)
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
	// Need to parse the desc file now
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

// IMPLEMENT HERE: function to parse the desc file so we can extract package info
