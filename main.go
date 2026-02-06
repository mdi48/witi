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

type InstallChain []string

type PackageCache struct {
	packages map[string]*Package
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <package-name>")
		os.Exit(1)
	}

	pkgName := os.Args[1]

	// load all packages once for efficiency
	cache, err := loadAllPackages()
	if err != nil {
		fmt.Printf("Error loading packages: %v\n", err)
		os.Exit(1)
	}

	pkg, err := getPackageInfo(pkgName, cache)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// find installation chains via dfs
	chains := findInstallationChains(pkgName, cache)

	displayPackageInfo(pkg, chains)
}

// loads all packages from pacman's local db into a cache to avoid repeated file reads
func loadAllPackages() (*PackageCache, error) {
	cache := &PackageCache{packages: make(map[string]*Package)}

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
			continue
		}
		cache.packages[pkg.Name] = pkg
	}

	return cache, nil
}

// gets package info from pacman's local database
func getPackageInfo(pkgName string, cache *PackageCache) (*Package, error) {
	pkg, exists := cache.packages[pkgName]
	if !exists {
		return nil, fmt.Errorf("package not found: %s", pkgName)
	}

	// get list of packages that require this package (reverse dependencies)
	pkg.RequiredBy = findRequiredBy(pkgName, cache)

	return pkg, nil
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

// searches all packages to find which ones depend on the given package
func findRequiredBy(pkgName string, cache *PackageCache) []string {
	var requiredBy []string

	for _, pkg := range cache.packages {
		for _, dep := range pkg.Dependencies {
			// dependencies might have version constraints (e.g. "libfoo>=1.0")
			// so we extract just the package name for comparison
			depName := cleanDependencyName(dep)

			if depName == pkgName {
				// use the NAME field from the parsed desc file instead of parsing directory name
				requiredBy = append(requiredBy, pkg.Name)
				break // no need to check other dependencies for this package
			}
		}
	}

	return requiredBy
}

// uses dfs to find all paths from explicitly installed packages to the target package
func findInstallationChains(targetPkg string, cache *PackageCache) []InstallChain {
	var chains []InstallChain

	// Build reverse dependency map for efficient lookup
	reverseDeps := buildReverseDependencyMap(cache)

	// Start DFS from the target package and work backwards to explicitly installed packages
	visited := make(map[string]bool)
	path := []string{targetPkg}
	dfsBackwards(targetPkg, path, visited, &chains, cache, reverseDeps)

	// Reverse each chain so it goes from root package -> target
	for i := range chains {
		reverseChain(chains[i])
	}

	return chains
}

// builds a map of package -> list of packages that depend on it
func buildReverseDependencyMap(cache *PackageCache) map[string][]string {
	reverseDeps := make(map[string][]string)

	for _, pkg := range cache.packages {
		for _, dep := range pkg.Dependencies {
			depName := cleanDependencyName(dep)
			reverseDeps[depName] = append(reverseDeps[depName], pkg.Name)
		}
	}

	return reverseDeps
}

// performs a backwards DFS from target package to explicitly installed packages
func dfsBackwards(currentPkg string, path []string, visited map[string]bool, chains *[]InstallChain, cache *PackageCache, reverseDeps map[string][]string) {
	// avoid cycles
	if visited[currentPkg] {
		return
	}
	visited[currentPkg] = true
	defer func() { visited[currentPkg] = false }()

	// check if current package is explicitly installed (we found a complete chain)
	pkg, exists := cache.packages[currentPkg]
	if exists && pkg.InstallReason == "Explicitly installed" {
		chain := make(InstallChain, len(path))
		copy(chain, path)
		*chains = append(*chains, chain)
		return
	}

	// explore packages that depend on the current package
	for _, dependentPkg := range reverseDeps[currentPkg] {
		newPath := make([]string, len(path), len(path)+1)
		copy(newPath, path)
		newPath = append(newPath, dependentPkg)
		dfsBackwards(dependentPkg, newPath, visited, chains, cache, reverseDeps)
	}
}

// reverses a chain in place
func reverseChain(chain []string) {
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
}

// returns all explicitly installed packages
func getExplicitlyInstalledPackages(cache *PackageCache) []string {
	var explicitPkgs []string

	for _, pkg := range cache.packages {
		if pkg.InstallReason == "Explicitly installed" {
			explicitPkgs = append(explicitPkgs, pkg.Name)
		}
	}

	return explicitPkgs
}

// remove version constraints and other modifiers from dependency names (e.g. "libfoo>=1.0" -> "libfoo")
// refactored this to avoid more redundancy (will need to do more later)
func cleanDependencyName(dep string) string {
	depName := strings.FieldsFunc(dep, func(r rune) bool {
		return r == ' ' || r == '<' || r == '>' || r == '='
	})[0]
	return depName
}

func displayPackageInfo(pkg *Package, chains []InstallChain) {
	fmt.Printf("\nPackage: %s (%s)\n", pkg.Name, pkg.Version)
	fmt.Printf("=========================================\n\n")

	if pkg.InstallReason == "Explicitly installed" {
		fmt.Printf("This package was EXPLICITLY installed by the user.\n")
	} else {
		fmt.Printf("This package was INSTALLED AS A DEPENDENCY for another package.\n")

		if len(chains) > 0 {
			fmt.Printf("\nInstallation chains leading to this package:\n")

			// Show up to 6 chains to avoid taking up too much terminal space (will add flags to control later)
			displayCount := len(chains)
			if displayCount > 6 {
				displayCount = 6
			}

			for i := 0; i < displayCount; i++ {
				chain := chains[i]
				fmt.Printf("Chain %d:\n", i+1)

				// display chain from root to target package
				for j := 0; j < len(chain); j++ {
					indent := strings.Repeat("  ", j)

					if j == 0 {
						fmt.Printf("%s%s (explicitly installed)\n", indent, chain[j])
					} else if j == len(chain)-1 {
						fmt.Printf("%s%s (target package)\n", indent, chain[j]) /// ill replace with an arrow thingy later like cool people use
					} else {
						fmt.Printf("%s%s\n", indent, chain[j])
					}
				}
				fmt.Println()
			}

			if len(chains) > 5 {
				fmt.Printf("... and %d more chain(s)\n\n", len(chains)-5)
			}
		} else {
			fmt.Printf("No installation chains found for this package (possible orphan package).\n")
		}
	}

	fmt.Printf("\n#########################################\n") // for ease of separating sections

	if len(pkg.Dependencies) > 0 {
		fmt.Printf("\nThis package DEPENDS ON the following packages:\n") // will control via flags later
		for _, dep := range pkg.Dependencies {
			fmt.Printf(" - %s\n", dep)
		}
	}

	fmt.Printf("\n#########################################\n")

	if len(pkg.RequiredBy) > 0 {
		fmt.Printf("\nThe following packages REQUIRE THIS PACKAGE:\n") // will control via flags later
		for _, req := range pkg.RequiredBy {
			fmt.Printf(" - %s\n", req)
		}
	}

	fmt.Println("\n=========================================")
}
