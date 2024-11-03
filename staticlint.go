package main

import (
	"encoding/json"
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
