// Мультичекер запускается двумя командами
// go build -o lint cmd/staticlint/staticlint.go
// ./lint ./...
package main

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
}

// Checker for os.Exit in main functions
var OSExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for call os.Exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				if c, ok := node.(*ast.CallExpr); ok {
					if s, ok := c.Fun.(*ast.SelectorExpr); ok {
						lib := s.X.(*ast.Ident)
						if lib.Name == "os" && s.Sel.Name == "Exit" {
							if arg, ok := c.Args[0].(*ast.CallExpr); ok {
								if argAst, ok := arg.Fun.(*ast.SelectorExpr); ok {
									argLib := argAst.X.(*ast.Ident)
									if argLib.Name != "m" || argAst.Sel.Name != "Run" {
										pass.Reportf(c.Pos(), "os.Exit called")
										return false
									}
								} else {
									pass.Reportf(c.Pos(), "os.Exit called")
									return false
								}
							} else {
								pass.Reportf(c.Pos(), "os.Exit called")
								return false
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		OSExitCheckAnalyzer,
	}

	addedChecks := make(map[string]bool)
	pullOfChecks := staticcheck.Analyzers
	pullOfChecks = append(pullOfChecks, simple.Analyzers...)
	pullOfChecks = append(pullOfChecks, quickfix.Analyzers...)

	for _, v := range pullOfChecks {
		for _, check := range cfg.Staticcheck {
			if matched, _ := regexp.MatchString(check, v.Analyzer.Name); matched {
				if !addedChecks[v.Analyzer.Name] {
					mychecks = append(mychecks, v.Analyzer)
				}
				addedChecks[v.Analyzer.Name] = true
			}
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
