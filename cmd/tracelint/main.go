package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mshogin/archlint/internal/linter"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
