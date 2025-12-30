// Package linter provides static analysis for tracer calls.
package linter

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
)

// Analyzer is the tracerlint analyzer.
var Analyzer = &analysis.Analyzer{
	Name: "tracerlint",
	Doc:  "checks that all functions have tracer.Enter() and tracer.ExitSuccess/ExitError() before returns",
	Run:  run,
}

// Config represents the configuration file structure.
type Config struct {
	Tracerlint struct {
		ExcludePackages []string `yaml:"exclude_packages"`
	} `yaml:"tracerlint"`
}

var excludePackages []string

func init() {
	excludePackages = loadConfig()
}

func loadConfig() []string {
	configFile := findConfigFile()
	if configFile == "" {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil
	}

	return config.Tracerlint.ExcludePackages
}

func findConfigFile() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for dir := cwd; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, ".archlint.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		if isExcluded(filename) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if fn.Body == nil {
				return true
			}

			if hasSkipTracerComment(fn) {
				return true
			}

			funcName := getFunctionName(fn)

			if !hasTracerEnter(fn) {
				pass.Reportf(fn.Pos(), "function %s missing tracer.Enter() at the beginning", funcName)
			}

			checkReturns(pass, fn, funcName)

			return true
		})
	}

	return nil, nil
}

func isExcluded(filename string) bool {
	if strings.Contains(filename, "go-build") {
		return true
	}

	if strings.HasSuffix(filename, "_test.go") {
		return true
	}

	for _, pkg := range excludePackages {
		if strings.Contains(filename, pkg) {
			return true
		}
	}

	return false
}

func hasSkipTracerComment(fn *ast.FuncDecl) bool {
	if fn.Doc == nil {
		return false
	}

	for _, comment := range fn.Doc.List {
		if strings.Contains(comment.Text, "@skip-tracer") {
			return true
		}
	}

	return false
}

func getFunctionName(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType := fn.Recv.List[0].Type
		var typeName string

		switch t := recvType.(type) {
		case *ast.Ident:
			typeName = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				typeName = ident.Name
			}
		}

		return typeName + "." + fn.Name.Name
	}

	return fn.Name.Name
}

func hasTracerEnter(fn *ast.FuncDecl) bool {
	if len(fn.Body.List) == 0 {
		return false
	}

	firstStmt := fn.Body.List[0]

	return isTracerCall(firstStmt, "Enter")
}

func checkReturns(pass *analysis.Pass, fn *ast.FuncDecl, funcName string) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}

		retStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		prevStmt := findPreviousStatement(fn.Body, retStmt)
		if prevStmt == nil || !isTracerExitCall(prevStmt) {
			pass.Reportf(retStmt.Pos(),
				"return in %s must be preceded by tracer.ExitSuccess() or tracer.ExitError()",
				funcName)
		}

		return true
	})
}

func findPreviousStatement(body *ast.BlockStmt, target ast.Stmt) ast.Stmt {
	return findPrevInBlock(body.List, target)
}

func findPrevInBlock(stmts []ast.Stmt, target ast.Stmt) ast.Stmt {
	for i, stmt := range stmts {
		if stmt == target && i > 0 {
			return stmts[i-1]
		}

		switch s := stmt.(type) {
		case *ast.IfStmt:
			if s.Body != nil {
				if prev := findPrevInBlock(s.Body.List, target); prev != nil {
					return prev
				}
			}
			if s.Else != nil {
				if block, ok := s.Else.(*ast.BlockStmt); ok {
					if prev := findPrevInBlock(block.List, target); prev != nil {
						return prev
					}
				}
			}
		case *ast.ForStmt:
			if s.Body != nil {
				if prev := findPrevInBlock(s.Body.List, target); prev != nil {
					return prev
				}
			}
		case *ast.RangeStmt:
			if s.Body != nil {
				if prev := findPrevInBlock(s.Body.List, target); prev != nil {
					return prev
				}
			}
		case *ast.SwitchStmt:
			if s.Body != nil {
				for _, clause := range s.Body.List {
					if cc, ok := clause.(*ast.CaseClause); ok {
						if prev := findPrevInBlock(cc.Body, target); prev != nil {
							return prev
						}
					}
				}
			}
		case *ast.TypeSwitchStmt:
			if s.Body != nil {
				for _, clause := range s.Body.List {
					if cc, ok := clause.(*ast.CaseClause); ok {
						if prev := findPrevInBlock(cc.Body, target); prev != nil {
							return prev
						}
					}
				}
			}
		case *ast.SelectStmt:
			if s.Body != nil {
				for _, clause := range s.Body.List {
					if cc, ok := clause.(*ast.CommClause); ok {
						if prev := findPrevInBlock(cc.Body, target); prev != nil {
							return prev
						}
					}
				}
			}
		case *ast.BlockStmt:
			if prev := findPrevInBlock(s.List, target); prev != nil {
				return prev
			}
		}
	}

	return nil
}

func isTracerExitCall(stmt ast.Stmt) bool {
	return isTracerCall(stmt, "ExitSuccess") || isTracerCall(stmt, "ExitError") || isTracerCall(stmt, "Exit")
}

func isTracerCall(stmt ast.Stmt, methodName string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	callExpr, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "tracer" && selExpr.Sel.Name == methodName
}
