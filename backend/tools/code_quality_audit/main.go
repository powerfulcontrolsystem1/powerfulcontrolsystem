// Command code_quality_audit measures structural debt without third-party
// dependencies. It is a non-regression gate, not a replacement for review,
// tests, coverage or runtime evidence.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type functionMetric struct {
	File       string `json:"file"`
	Name       string `json:"name"`
	Start      int    `json:"start"`
	Lines      int    `json:"lines"`
	Statements int    `json:"statements"`
	Hash       string `json:"-"`
}

type duplicateGroup struct {
	Hash      string           `json:"hash"`
	Functions []functionMetric `json:"functions"`
}

type fileDebtMetric struct {
	File                  string `json:"file"`
	DBCallsWithoutContext int    `json:"db_calls_without_context"`
	ExplicitIgnoredResult int    `json:"explicit_ignored_results"`
}

type metrics struct {
	ProductionFiles          int              `json:"production_files"`
	ProductionFunctions      int              `json:"production_functions"`
	FunctionsOver100Lines    int              `json:"functions_over_100_lines"`
	FunctionsOver200Lines    int              `json:"functions_over_200_lines"`
	LargestFunctionLines     int              `json:"largest_function_lines"`
	ExactDuplicateBodyGroups int              `json:"exact_duplicate_body_groups"`
	DBCallsWithoutContext    int              `json:"db_calls_without_context"`
	ExplicitIgnoredResults   int              `json:"explicit_ignored_results"`
	LargestFunctions         []functionMetric `json:"largest_functions"`
	DuplicateGroups          []duplicateGroup `json:"duplicate_groups"`
	TopDebtFiles             []fileDebtMetric `json:"top_debt_files,omitempty"`
}

type baseline struct {
	Version int     `json:"version"`
	Max     metrics `json:"max"`
}

type report struct {
	OK          bool      `json:"ok"`
	Root        string    `json:"root"`
	Metrics     metrics   `json:"metrics"`
	Baseline    *baseline `json:"baseline,omitempty"`
	Regressions []string  `json:"regressions"`
}

func selectorName(call *ast.CallExpr) string {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		if identifier, ok := call.Fun.(*ast.Ident); ok {
			return identifier.Name
		}
		return ""
	}
	var receiver strings.Builder
	_ = printer.Fprint(&receiver, token.NewFileSet(), selector.X)
	if receiver.Len() == 0 {
		return selector.Sel.Name
	}
	return receiver.String() + "." + selector.Sel.Name
}

func isDBCallWithoutContext(name string) bool {
	// net/http URL.Query() is a parameter parser, not a database operation.
	// Keeping it out prevents request-heavy handlers from inflating SQL debt.
	if strings.HasSuffix(name, ".URL.Query") {
		return false
	}
	for _, suffix := range []string{".Query", ".QueryRow", ".Exec", ".Prepare", ".Begin"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func measure(root string) (metrics, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return metrics{}, err
	}
	fset := token.NewFileSet()
	result := metrics{}
	byHash := map[string][]functionMetric{}
	debtByFile := map[string]*fileDebtMetric{}
	allFunctions := make([]functionMetric, 0, 4096)

	err = filepath.WalkDir(absRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			relativeDir, relErr := filepath.Rel(absRoot, path)
			if relErr == nil && filepath.ToSlash(relativeDir) == "tools/code_quality_audit" {
				// The auditor must not inflate the baseline it is enforcing.
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		relative, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", relative, err)
		}
		result.ProductionFiles++
		fileDebt := &fileDebtMetric{File: relative}
		debtByFile[relative] = fileDebt

		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			start := fset.Position(function.Pos()).Line
			end := fset.Position(function.End()).Line
			statements := 0
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if _, ok := node.(ast.Stmt); ok {
					statements++
				}
				return true
			})
			var body strings.Builder
			_ = printer.Fprint(&body, fset, function.Body)
			sum := sha256.Sum256([]byte(body.String()))
			item := functionMetric{
				File: relative, Name: function.Name.Name, Start: start,
				Lines: end - start + 1, Statements: statements,
				Hash: hex.EncodeToString(sum[:8]),
			}
			allFunctions = append(allFunctions, item)
			result.ProductionFunctions++
			if item.Lines > 100 {
				result.FunctionsOver100Lines++
			}
			if item.Lines > 200 {
				result.FunctionsOver200Lines++
			}
			if item.Lines > result.LargestFunctionLines {
				result.LargestFunctionLines = item.Lines
			}
			if statements >= 6 {
				byHash[item.Hash] = append(byHash[item.Hash], item)
			}
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch current := node.(type) {
			case *ast.AssignStmt:
				if len(current.Lhs) == 1 {
					if identifier, ok := current.Lhs[0].(*ast.Ident); ok && identifier.Name == "_" {
						result.ExplicitIgnoredResults++
						fileDebt.ExplicitIgnoredResult++
					}
				}
			case *ast.CallExpr:
				if isDBCallWithoutContext(selectorName(current)) {
					result.DBCallsWithoutContext++
					fileDebt.DBCallsWithoutContext++
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return metrics{}, err
	}

	for _, item := range debtByFile {
		if item.DBCallsWithoutContext == 0 && item.ExplicitIgnoredResult == 0 {
			continue
		}
		result.TopDebtFiles = append(result.TopDebtFiles, *item)
	}
	sort.Slice(result.TopDebtFiles, func(i, j int) bool {
		left := result.TopDebtFiles[i].DBCallsWithoutContext + result.TopDebtFiles[i].ExplicitIgnoredResult
		right := result.TopDebtFiles[j].DBCallsWithoutContext + result.TopDebtFiles[j].ExplicitIgnoredResult
		if left != right {
			return left > right
		}
		return result.TopDebtFiles[i].File < result.TopDebtFiles[j].File
	})
	if len(result.TopDebtFiles) > 30 {
		result.TopDebtFiles = result.TopDebtFiles[:30]
	}

	sort.Slice(allFunctions, func(i, j int) bool {
		if allFunctions[i].Lines != allFunctions[j].Lines {
			return allFunctions[i].Lines > allFunctions[j].Lines
		}
		if allFunctions[i].File != allFunctions[j].File {
			return allFunctions[i].File < allFunctions[j].File
		}
		return allFunctions[i].Start < allFunctions[j].Start
	})
	if len(allFunctions) > 20 {
		allFunctions = allFunctions[:20]
	}
	for index := range allFunctions {
		allFunctions[index].Hash = ""
	}
	result.LargestFunctions = allFunctions

	for hash, functions := range byHash {
		files := map[string]struct{}{}
		for _, function := range functions {
			files[function.File] = struct{}{}
		}
		if len(functions) < 2 || len(files) < 2 {
			continue
		}
		for index := range functions {
			functions[index].Hash = ""
		}
		result.DuplicateGroups = append(result.DuplicateGroups, duplicateGroup{Hash: hash, Functions: functions})
	}
	sort.Slice(result.DuplicateGroups, func(i, j int) bool {
		if len(result.DuplicateGroups[i].Functions) != len(result.DuplicateGroups[j].Functions) {
			return len(result.DuplicateGroups[i].Functions) > len(result.DuplicateGroups[j].Functions)
		}
		return result.DuplicateGroups[i].Hash < result.DuplicateGroups[j].Hash
	})
	result.ExactDuplicateBodyGroups = len(result.DuplicateGroups)
	return result, nil
}

func compare(current metrics, limits metrics) []string {
	checks := []struct {
		name    string
		current int
		maximum int
	}{
		{"functions_over_100_lines", current.FunctionsOver100Lines, limits.FunctionsOver100Lines},
		{"functions_over_200_lines", current.FunctionsOver200Lines, limits.FunctionsOver200Lines},
		{"largest_function_lines", current.LargestFunctionLines, limits.LargestFunctionLines},
		{"exact_duplicate_body_groups", current.ExactDuplicateBodyGroups, limits.ExactDuplicateBodyGroups},
		{"db_calls_without_context", current.DBCallsWithoutContext, limits.DBCallsWithoutContext},
		{"explicit_ignored_results", current.ExplicitIgnoredResults, limits.ExplicitIgnoredResults},
	}
	regressions := make([]string, 0)
	for _, check := range checks {
		if check.current > check.maximum {
			regressions = append(regressions, fmt.Sprintf("%s: %d > %d", check.name, check.current, check.maximum))
		}
	}
	return regressions
}

func readBaseline(path string) (*baseline, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	// #nosec G304 -- path is an explicit local CLI argument used only by the
	// repository's trusted CI/preflight operator, never untrusted HTTP input.
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	var value baseline
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	if value.Version != 1 {
		return nil, fmt.Errorf("unsupported baseline version %d", value.Version)
	}
	return &value, nil
}

func main() {
	root := flag.String("root", ".", "Go source root")
	baselinePath := flag.String("baseline", "", "non-regression baseline JSON")
	outputPath := flag.String("out", "", "optional JSON report path")
	flag.Parse()

	current, err := measure(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	limits, err := readBaseline(*baselinePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result := report{OK: true, Root: *root, Metrics: current, Baseline: limits, Regressions: []string{}}
	if limits != nil {
		result.Regressions = compare(current, limits.Max)
		result.OK = len(result.Regressions) == 0
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded = append(encoded, '\n')
	if strings.TrimSpace(*outputPath) != "" {
		if err := os.MkdirAll(filepath.Dir(*outputPath), 0o750); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*outputPath, encoded, 0o600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Printf("code quality: ok=%t over100=%d over200=%d duplicates=%d db_without_context=%d ignored=%d\n",
		result.OK, current.FunctionsOver100Lines, current.FunctionsOver200Lines,
		current.ExactDuplicateBodyGroups, current.DBCallsWithoutContext, current.ExplicitIgnoredResults)
	if !result.OK {
		for _, regression := range result.Regressions {
			fmt.Fprintln(os.Stderr, regression)
		}
		os.Exit(2)
	}
}
