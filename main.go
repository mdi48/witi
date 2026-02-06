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

	pkg, err := getPackageInfo(pkgName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	displayPackageInfo(pkg)
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

	// get list of packages that require this package (reverse dependencies)
	pkg.RequiredBy, err = findRequiredBy(pkgName)
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

// searches all packages to find which ones depend on the given package (will benefit from efficiency improvements in the future, but for now it works)
func findRequiredBy(pkgName string) ([]string, error) {
	var requiredBy []string

	entries, err := os.ReadDir(pacmanLocalDBPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		descFile := filepath.Join(pacmanLocalDBPath, entry.Name(), "desc")
		pkg, err := parseDescFile(descFile)
		if err != nil {
			continue // skip packages that we can't read
		}

		for _, dep := range pkg.Dependencies {
			// dependencies might have version constraints (e.g. "libfoo>=1.0")
			// so we extract just the package name for comparison
			depName := strings.FieldsFunc(dep, func(r rune) bool {
				return r == '>' || r == '<' || r == '='
			})[0]

			if depName == pkgName {
				// use the NAME field from the parsed desc file instead of parsing directory name
				requiredBy = append(requiredBy, pkg.Name)
				break // no need to check other dependencies for this package
			}
		}
	}

	return requiredBy, nil
}

func displayPackageInfo(pkg *Package) {
	fmt.Printf("\nPackage: %s (%s)\n", pkg.Name, pkg.Version)
	fmt.Printf("=========================================\n\n")

	if pkg.InstallReason == "Explicitly installed" {
		fmt.Printf("This package was EXPLICITLY installed by the user.\n")
	} else {
		fmt.Printf("This package was INSTALLED AS A DEPENDENCY for another package.\n")
	}

	fmt.Printf("\n#########################################\n") // for ease of separating dep vs req

	if len(pkg.Dependencies) > 0 {
		fmt.Printf("\nThis package DEPENDS ON the following packages:\n")
		for _, dep := range pkg.Dependencies {
			fmt.Printf(" - %s\n", dep)
		}
	}

	fmt.Printf("\n#########################################\n")

	if len(pkg.RequiredBy) > 0 {
		fmt.Printf("\nThe following packages REQUIRE THIS PACKAGE:\n")
		for _, req := range pkg.RequiredBy {
			fmt.Printf(" - %s\n", req)
		}
	} else {
		if pkg.InstallReason == "Installed as a dependency" {
			fmt.Printf("Not required by any other package (possible orphan package).\n")
		}
	}

	fmt.Println("\n=========================================")
}
