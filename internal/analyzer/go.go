// Package analyzer provides Go source code analysis capabilities.
package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/mshogin/archlint/internal/model"
	"github.com/mshogin/archlint/pkg/tracer"
)

// PackageInfo holds information about a Go package.
type PackageInfo struct {
	Name    string
	Path    string
	Dir     string
	Imports []string
}

// TypeInfo holds information about a type declaration.
type TypeInfo struct {
	Name       string
	Package    string
	Kind       string // struct, interface
	File       string
	Line       int
	Fields     []FieldInfo
	Embeds     []string
	Implements []string
}

// FieldInfo holds information about a struct field.
type FieldInfo struct {
	Name     string
	TypeName string
	TypePkg  string
}

// FunctionInfo holds information about a function.
type FunctionInfo struct {
	Name    string
	Package string
	File    string
	Line    int
	Calls   []CallInfo
}

// MethodInfo holds information about a method.
type MethodInfo struct {
	Name     string
	Receiver string
	Package  string
	File     string
	Line     int
	Calls    []CallInfo
}

// CallInfo holds information about a function/method call.
type CallInfo struct {
	Target   string
	IsMethod bool
	Receiver string
	Line     int
}

// GoAnalyzer analyzes Go source code and builds an architecture graph.
type GoAnalyzer struct {
	packages  map[string]*PackageInfo
	types     map[string]*TypeInfo
	functions map[string]*FunctionInfo
	methods   map[string]*MethodInfo
	nodes     []model.Node
	edges     []model.Edge
	baseDir   string
	modulePath string
}

// NewGoAnalyzer creates a new GoAnalyzer instance.
func NewGoAnalyzer() *GoAnalyzer {
	tracer.Enter("analyzer.NewGoAnalyzer")

	a := &GoAnalyzer{
		packages:  make(map[string]*PackageInfo),
		types:     make(map[string]*TypeInfo),
		functions: make(map[string]*FunctionInfo),
		methods:   make(map[string]*MethodInfo),
		nodes:     []model.Node{},
		edges:     []model.Edge{},
	}

	tracer.ExitSuccess("analyzer.NewGoAnalyzer")
	return a
}

// Analyze analyzes Go source code in the given directory and returns an architecture graph.
func (a *GoAnalyzer) Analyze(dir string) (*model.Graph, error) {
	tracer.Enter("analyzer.GoAnalyzer.Analyze")

	absDir, err := filepath.Abs(dir)
	if err != nil {
		tracer.ExitError("analyzer.GoAnalyzer.Analyze", err)
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	a.baseDir = absDir

	a.modulePath = a.detectModulePath()

	err = filepath.Walk(absDir, a.walkFunc)
	if err != nil {
		tracer.ExitError("analyzer.GoAnalyzer.Analyze", err)
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	a.buildGraph()

	graph := &model.Graph{
		Nodes: a.nodes,
		Edges: a.edges,
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.Analyze")
	return graph, nil
}

func (a *GoAnalyzer) detectModulePath() string {
	tracer.Enter("analyzer.GoAnalyzer.detectModulePath")

	goModPath := filepath.Join(a.baseDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		tracer.ExitSuccess("analyzer.GoAnalyzer.detectModulePath")
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			tracer.ExitSuccess("analyzer.GoAnalyzer.detectModulePath")
			return strings.TrimSpace(modulePath)
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.detectModulePath")
	return ""
}

func (a *GoAnalyzer) walkFunc(path string, info os.FileInfo, err error) error {
	tracer.Enter("analyzer.GoAnalyzer.walkFunc")

	if err != nil {
		tracer.ExitError("analyzer.GoAnalyzer.walkFunc", err)
		return err
	}

	if info.IsDir() {
		name := info.Name()
		if name == "vendor" || name == "node_modules" || name == ".git" || name == "bin" {
			tracer.ExitSuccess("analyzer.GoAnalyzer.walkFunc")
			return filepath.SkipDir
		}
		tracer.ExitSuccess("analyzer.GoAnalyzer.walkFunc")
		return nil
	}

	if !strings.HasSuffix(path, ".go") {
		tracer.ExitSuccess("analyzer.GoAnalyzer.walkFunc")
		return nil
	}

	if strings.HasSuffix(path, "_test.go") {
		tracer.ExitSuccess("analyzer.GoAnalyzer.walkFunc")
		return nil
	}

	parseErr := a.parseFile(path)
	if parseErr != nil {
		tracer.ExitError("analyzer.GoAnalyzer.walkFunc", parseErr)
		return parseErr
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.walkFunc")
	return nil
}

func (a *GoAnalyzer) parseFile(filename string) error {
	tracer.Enter("analyzer.GoAnalyzer.parseFile")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		tracer.ExitError("analyzer.GoAnalyzer.parseFile", err)
		return fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	dir := filepath.Dir(filename)
	relDir, _ := filepath.Rel(a.baseDir, dir)
	pkgPath := a.modulePath
	if relDir != "." && relDir != "" {
		pkgPath = a.modulePath + "/" + relDir
	}

	if _, exists := a.packages[pkgPath]; !exists {
		a.packages[pkgPath] = &PackageInfo{
			Name:    node.Name.Name,
			Path:    pkgPath,
			Dir:     dir,
			Imports: []string{},
		}
	}

	for _, imp := range node.Imports {
		impPath := strings.Trim(imp.Path.Value, "\"")
		if !a.isStdLib(impPath) {
			a.packages[pkgPath].Imports = append(a.packages[pkgPath].Imports, impPath)
		}
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			a.parseGenDecl(d, pkgPath, filename, fset)
		case *ast.FuncDecl:
			a.parseFuncDecl(d, pkgPath, filename, fset)
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.parseFile")
	return nil
}

func (a *GoAnalyzer) parseGenDecl(decl *ast.GenDecl, pkgPath, filename string, fset *token.FileSet) {
	tracer.Enter("analyzer.GoAnalyzer.parseGenDecl")

	if decl.Tok != token.TYPE {
		tracer.ExitSuccess("analyzer.GoAnalyzer.parseGenDecl")
		return
	}

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		typeID := pkgPath + "." + typeSpec.Name.Name
		pos := fset.Position(typeSpec.Pos())

		typeInfo := &TypeInfo{
			Name:    typeSpec.Name.Name,
			Package: pkgPath,
			File:    filename,
			Line:    pos.Line,
			Fields:  []FieldInfo{},
			Embeds:  []string{},
		}

		switch t := typeSpec.Type.(type) {
		case *ast.StructType:
			typeInfo.Kind = "struct"
			if t.Fields != nil {
				for _, field := range t.Fields.List {
					a.parseStructField(field, typeInfo, pkgPath)
				}
			}
		case *ast.InterfaceType:
			typeInfo.Kind = "interface"
		}

		a.types[typeID] = typeInfo
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.parseGenDecl")
}

func (a *GoAnalyzer) parseStructField(field *ast.Field, typeInfo *TypeInfo, pkgPath string) {
	tracer.Enter("analyzer.GoAnalyzer.parseStructField")

	typeName, typePkg := a.resolveTypeName(field.Type, pkgPath)

	if len(field.Names) == 0 {
		typeInfo.Embeds = append(typeInfo.Embeds, typeName)
		tracer.ExitSuccess("analyzer.GoAnalyzer.parseStructField")
		return
	}

	for _, name := range field.Names {
		typeInfo.Fields = append(typeInfo.Fields, FieldInfo{
			Name:     name.Name,
			TypeName: typeName,
			TypePkg:  typePkg,
		})
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.parseStructField")
}

func (a *GoAnalyzer) resolveTypeName(expr ast.Expr, currentPkg string) (string, string) {
	tracer.Enter("analyzer.GoAnalyzer.resolveTypeName")

	var typeName, typePkg string

	switch t := expr.(type) {
	case *ast.Ident:
		typeName = t.Name
		typePkg = currentPkg
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			typePkg = ident.Name
			typeName = t.Sel.Name
		}
	case *ast.StarExpr:
		typeName, typePkg = a.resolveTypeName(t.X, currentPkg)
	case *ast.ArrayType:
		typeName, typePkg = a.resolveTypeName(t.Elt, currentPkg)
	case *ast.MapType:
		typeName, typePkg = a.resolveTypeName(t.Value, currentPkg)
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.resolveTypeName")
	return typeName, typePkg
}

func (a *GoAnalyzer) parseFuncDecl(decl *ast.FuncDecl, pkgPath, filename string, fset *token.FileSet) {
	tracer.Enter("analyzer.GoAnalyzer.parseFuncDecl")

	pos := fset.Position(decl.Pos())

	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		receiver := a.getReceiverName(decl.Recv.List[0].Type)
		methodID := pkgPath + "." + receiver + "." + decl.Name.Name

		methodInfo := &MethodInfo{
			Name:     decl.Name.Name,
			Receiver: receiver,
			Package:  pkgPath,
			File:     filename,
			Line:     pos.Line,
			Calls:    []CallInfo{},
		}

		if decl.Body != nil {
			methodInfo.Calls = a.collectCalls(decl.Body, fset)
		}

		a.methods[methodID] = methodInfo
	} else {
		funcID := pkgPath + "." + decl.Name.Name

		funcInfo := &FunctionInfo{
			Name:    decl.Name.Name,
			Package: pkgPath,
			File:    filename,
			Line:    pos.Line,
			Calls:   []CallInfo{},
		}

		if decl.Body != nil {
			funcInfo.Calls = a.collectCalls(decl.Body, fset)
		}

		a.functions[funcID] = funcInfo
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.parseFuncDecl")
}

func (a *GoAnalyzer) getReceiverName(expr ast.Expr) string {
	tracer.Enter("analyzer.GoAnalyzer.getReceiverName")

	var name string

	switch t := expr.(type) {
	case *ast.Ident:
		name = t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			name = ident.Name
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.getReceiverName")
	return name
}

func (a *GoAnalyzer) collectCalls(body *ast.BlockStmt, fset *token.FileSet) []CallInfo {
	tracer.Enter("analyzer.GoAnalyzer.collectCalls")

	var calls []CallInfo

	ast.Inspect(body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		pos := fset.Position(callExpr.Pos())

		switch fun := callExpr.Fun.(type) {
		case *ast.Ident:
			if !a.isBuiltin(fun.Name) {
				calls = append(calls, CallInfo{
					Target:   fun.Name,
					IsMethod: false,
					Line:     pos.Line,
				})
			}
		case *ast.SelectorExpr:
			if ident, ok := fun.X.(*ast.Ident); ok {
				calls = append(calls, CallInfo{
					Target:   fun.Sel.Name,
					IsMethod: true,
					Receiver: ident.Name,
					Line:     pos.Line,
				})
			}
		}

		return true
	})

	tracer.ExitSuccess("analyzer.GoAnalyzer.collectCalls")
	return calls
}

func (a *GoAnalyzer) isBuiltin(name string) bool {
	builtins := map[string]bool{
		"make": true, "new": true, "len": true, "cap": true,
		"append": true, "copy": true, "delete": true, "close": true,
		"panic": true, "recover": true, "print": true, "println": true,
		"complex": true, "real": true, "imag": true,
	}
	return builtins[name]
}

func (a *GoAnalyzer) isStdLib(importPath string) bool {
	if !strings.Contains(importPath, ".") {
		return true
	}

	stdlibPrefixes := []string{
		"archive/", "bufio", "bytes", "compress/", "container/",
		"context", "crypto/", "database/", "debug/", "embed",
		"encoding/", "errors", "expvar", "flag", "fmt",
		"go/", "hash/", "html/", "image/", "index/",
		"io", "log", "math/", "mime/", "net/",
		"os", "path/", "plugin", "reflect", "regexp",
		"runtime/", "sort", "strconv", "strings", "sync",
		"syscall", "testing", "text/", "time", "unicode/",
		"unsafe",
	}

	for _, prefix := range stdlibPrefixes {
		if strings.HasPrefix(importPath, prefix) || importPath == strings.TrimSuffix(prefix, "/") {
			return true
		}
	}

	return false
}

func (a *GoAnalyzer) buildGraph() {
	tracer.Enter("analyzer.GoAnalyzer.buildGraph")

	a.buildPackageNodes()
	a.buildTypeNodes()
	a.buildFunctionNodes()
	a.buildMethodNodes()
	a.buildImportEdges()
	a.buildContainsEdges()
	a.buildCallEdges()
	a.buildTypeDependencyEdges()

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildGraph")
}

func (a *GoAnalyzer) buildPackageNodes() {
	tracer.Enter("analyzer.GoAnalyzer.buildPackageNodes")

	for path, pkg := range a.packages {
		a.nodes = append(a.nodes, model.Node{
			ID:     path,
			Title:  pkg.Name,
			Entity: "package",
		})
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildPackageNodes")
}

func (a *GoAnalyzer) buildTypeNodes() {
	tracer.Enter("analyzer.GoAnalyzer.buildTypeNodes")

	for id, typeInfo := range a.types {
		entity := "struct"
		if typeInfo.Kind == "interface" {
			entity = "interface"
		}

		a.nodes = append(a.nodes, model.Node{
			ID:     id,
			Title:  typeInfo.Name,
			Entity: entity,
		})
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildTypeNodes")
}

func (a *GoAnalyzer) buildFunctionNodes() {
	tracer.Enter("analyzer.GoAnalyzer.buildFunctionNodes")

	for id, funcInfo := range a.functions {
		a.nodes = append(a.nodes, model.Node{
			ID:     id,
			Title:  funcInfo.Name,
			Entity: "function",
		})
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildFunctionNodes")
}

func (a *GoAnalyzer) buildMethodNodes() {
	tracer.Enter("analyzer.GoAnalyzer.buildMethodNodes")

	for id, methodInfo := range a.methods {
		a.nodes = append(a.nodes, model.Node{
			ID:     id,
			Title:  methodInfo.Name,
			Entity: "method",
		})
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildMethodNodes")
}

func (a *GoAnalyzer) buildImportEdges() {
	tracer.Enter("analyzer.GoAnalyzer.buildImportEdges")

	for path, pkg := range a.packages {
		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, a.modulePath) {
				a.edges = append(a.edges, model.Edge{
					From: path,
					To:   imp,
					Type: "import",
				})
			}
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildImportEdges")
}

func (a *GoAnalyzer) buildContainsEdges() {
	tracer.Enter("analyzer.GoAnalyzer.buildContainsEdges")

	for id, typeInfo := range a.types {
		a.edges = append(a.edges, model.Edge{
			From: typeInfo.Package,
			To:   id,
			Type: "contains",
		})
	}

	for id, funcInfo := range a.functions {
		a.edges = append(a.edges, model.Edge{
			From: funcInfo.Package,
			To:   id,
			Type: "contains",
		})
	}

	for id, methodInfo := range a.methods {
		typeID := methodInfo.Package + "." + methodInfo.Receiver
		if _, exists := a.types[typeID]; exists {
			a.edges = append(a.edges, model.Edge{
				From: typeID,
				To:   id,
				Type: "contains",
			})
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildContainsEdges")
}

func (a *GoAnalyzer) buildCallEdges() {
	tracer.Enter("analyzer.GoAnalyzer.buildCallEdges")

	for id, funcInfo := range a.functions {
		for _, call := range funcInfo.Calls {
			target := a.resolveCallTarget(call, funcInfo.Package)
			if target != "" {
				a.edges = append(a.edges, model.Edge{
					From:   id,
					To:     target,
					Type:   "calls",
					Method: call.Target,
				})
			}
		}
	}

	for id, methodInfo := range a.methods {
		for _, call := range methodInfo.Calls {
			target := a.resolveCallTarget(call, methodInfo.Package)
			if target != "" {
				a.edges = append(a.edges, model.Edge{
					From:   id,
					To:     target,
					Type:   "calls",
					Method: call.Target,
				})
			}
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildCallEdges")
}

func (a *GoAnalyzer) resolveCallTarget(call CallInfo, currentPkg string) string {
	tracer.Enter("analyzer.GoAnalyzer.resolveCallTarget")

	if !call.IsMethod {
		funcID := currentPkg + "." + call.Target
		if _, exists := a.functions[funcID]; exists {
			tracer.ExitSuccess("analyzer.GoAnalyzer.resolveCallTarget")
			return funcID
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.resolveCallTarget")
	return ""
}

func (a *GoAnalyzer) buildTypeDependencyEdges() {
	tracer.Enter("analyzer.GoAnalyzer.buildTypeDependencyEdges")

	primitives := map[string]bool{
		"bool": true, "string": true, "int": true, "int8": true,
		"int16": true, "int32": true, "int64": true, "uint": true,
		"uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"uintptr": true, "byte": true, "rune": true, "float32": true,
		"float64": true, "complex64": true, "complex128": true, "error": true,
	}

	for id, typeInfo := range a.types {
		for _, embed := range typeInfo.Embeds {
			if primitives[embed] {
				continue
			}

			embedID := typeInfo.Package + "." + embed
			if _, exists := a.types[embedID]; exists {
				a.edges = append(a.edges, model.Edge{
					From: id,
					To:   embedID,
					Type: "embeds",
				})
			}
		}

		for _, field := range typeInfo.Fields {
			if primitives[field.TypeName] {
				continue
			}

			var depID string
			if field.TypePkg == typeInfo.Package {
				depID = typeInfo.Package + "." + field.TypeName
			}

			if depID != "" {
				if _, exists := a.types[depID]; exists {
					a.edges = append(a.edges, model.Edge{
						From: id,
						To:   depID,
						Type: "uses",
					})
				}
			}
		}
	}

	tracer.ExitSuccess("analyzer.GoAnalyzer.buildTypeDependencyEdges")
}
