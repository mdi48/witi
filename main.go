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
}

// IMPLEMENT HERE: function to get package info (makes call to findPackageDir)

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
