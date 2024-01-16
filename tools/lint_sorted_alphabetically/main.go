package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

func main() {
	switch {
	case len(os.Args) == 1 || len(os.Args) > 2:
		displayUsage()
	case len(os.Args) == 2 && os.Args[1] == "run":
		lintFiles()
	case len(os.Args) == 2 && os.Args[1] == "test":
		runTests()
	default:
		fmt.Printf("Error: unknown argument: %s", os.Args[1])
		os.Exit(1)
	}
}

func displayUsage() {
	fmt.Println(`
This tool verifies that all Go struct definitions and usages contain the struct properties sorted alphabetically.

Available commands:
   run   Lints the source code files
   test  Verifies that this tool works
`[1:])
}

// shouldIgnorePath indicates whether the file with the given path should be ignored (not formatted).
func shouldIgnorePath(path string) bool {
	return false // strings.HasPrefix(path, "vendor/") || path == "src/config/configdomain/push_hook.go" || path == "src/config/configdomain/offline.go" || path == "src/cli/print/logger.go"
}

func lintFiles() {
	err := filepath.WalkDir(".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil || dirEntry.IsDir() || !isGoFile(path) || shouldIgnorePath(path) {
			return err
		}
		fmt.Print(".")
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		issues := lintFileContent(string(content), path)
		for _, issue := range issues {
			fmt.Printf("%s  %s", path, issue)
		}
		return nil
	})
	fmt.Println()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

type issue struct {
	line int
	msg  string
}

var structDefRE = *regexp.MustCompile(`(?ms)^type \w+ struct \{\n.*?\n\}`)

func lintFileContent(content, filepath string) []string {
	return []string{}
}

func findStructDefinitions(code string) []string {
	return structDefRE.FindAllString(code, -1)
}

func isGoFile(path string) bool {
	if strings.HasSuffix(path, "_test.go") {
		return false
	}
	return strings.HasSuffix(path, ".go")
}

/************************************************************************************
 * TESTS
 */

func testCorrectCode() {
	give := `
package test

var a = 1

type MyStruct struct {
	count int
	name string
}

var global = MyStruct{
	count: 1,
	name: "one",
}

calling(MyStruct{
	count: 1,
	name: "one",
})
`
	have := lintFileContent(give, "myfile.go")
	want := []string{}
	assertDeepEqual(want, have, "correct code")
}

func testUnsortedDeclaration() {
	give := `
package test

var a = 1

type MyStruct struct {
	name string
	count int
}
`
	have := lintFileContent(give, "myfile.go")
	want := []string{`myfile.go: unsorted fields in definition of struct "MyStruct"`}
	assertDeepEqual(want, have, "correct code")
}

func runTests() {
	testCorrectCode()
	testUnsortedDeclaration()
	fmt.Println()
}

func assertEqual[T comparable](want, have T, testName string) {
	fmt.Print(".")
	if have != want {
		fmt.Printf("\nTEST FAILURE in %q\n", testName)
		fmt.Println("\n\nWANT")
		fmt.Println("--------------------------------------------------------")
		fmt.Println(want)
		fmt.Println("\n\nHAVE")
		fmt.Println("--------------------------------------------------------")
		fmt.Println(have)
		os.Exit(1)
	}
}

func assertDeepEqual[T any](want, have T, testName string) {
	fmt.Print(".")
	if !reflect.DeepEqual(want, have) {
		fmt.Printf("\nTEST FAILURE in %q\n", testName)
		fmt.Println("\n\nWANT")
		fmt.Println("--------------------------------------------------------")
		fmt.Println(want)
		fmt.Println("\n\nHAVE")
		fmt.Println("--------------------------------------------------------")
		fmt.Println(have)
		os.Exit(1)
	}
}
