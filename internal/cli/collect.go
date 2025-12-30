package cli

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mshogin/archlint/internal/analyzer"
	"github.com/mshogin/archlint/internal/model"
	"github.com/mshogin/archlint/pkg/tracer"
)

var (
	errDirNotExist       = errors.New("directory does not exist")
	errUnsupportedLang   = errors.New("unsupported language")
	errFileCreate        = errors.New("failed to create file")
	errYAMLSerialization = errors.New("failed to serialize YAML")
)

var (
	collectOutputFile string
	collectLanguage   string
)

var collectCmd = &cobra.Command{
	Use:   "collect [directory]",
	Short: "Collect architecture from source code",
	Long: `Analyzes source code and builds an architecture graph in YAML format.

Example:
  archlint collect . -l go -o architecture.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runCollect,
}

func init() {
	collectCmd.Flags().StringVarP(&collectOutputFile, "output", "o",
		"architecture.yaml", "Output YAML file")
	collectCmd.Flags().StringVarP(&collectLanguage, "language", "l",
		"go", "Programming language (go)")
	rootCmd.AddCommand(collectCmd)
}

func runCollect(cmd *cobra.Command, args []string) error {
	tracer.Enter("cli.runCollect")

	codeDir := args[0]

	if _, err := os.Stat(codeDir); os.IsNotExist(err) {
		tracer.ExitError("cli.runCollect", errDirNotExist)
		return fmt.Errorf("%w: %s", errDirNotExist, codeDir)
	}

	fmt.Printf("Analyzing code: %s (language: %s)\n", codeDir, collectLanguage)

	graph, err := analyzeCode(codeDir)
	if err != nil {
		tracer.ExitError("cli.runCollect", err)
		return err
	}

	printStats(graph)

	if err := saveGraph(graph); err != nil {
		tracer.ExitError("cli.runCollect", err)
		return err
	}

	fmt.Printf("Graph saved to %s\n", collectOutputFile)

	tracer.ExitSuccess("cli.runCollect")
	return nil
}

func analyzeCode(codeDir string) (*model.Graph, error) {
	tracer.Enter("cli.analyzeCode")

	if collectLanguage != "go" {
		tracer.ExitError("cli.analyzeCode", errUnsupportedLang)
		return nil, fmt.Errorf("%w: %s", errUnsupportedLang, collectLanguage)
	}

	a := analyzer.NewGoAnalyzer()
	graph, err := a.Analyze(codeDir)
	if err != nil {
		tracer.ExitError("cli.analyzeCode", err)
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	tracer.ExitSuccess("cli.analyzeCode")
	return graph, nil
}

func printStats(graph *model.Graph) {
	tracer.Enter("cli.printStats")

	stats := make(map[string]int)
	for _, node := range graph.Nodes {
		stats[node.Entity]++
	}

	fmt.Printf("Found components: %d\n", len(graph.Nodes))
	for entity, count := range stats {
		fmt.Printf("  - %s: %d\n", entity, count)
	}
	fmt.Printf("Found links: %d\n", len(graph.Edges))

	tracer.ExitSuccess("cli.printStats")
}

func saveGraph(graph *model.Graph) error {
	tracer.Enter("cli.saveGraph")

	file, err := os.Create(collectOutputFile)
	if err != nil {
		tracer.ExitError("cli.saveGraph", err)
		return fmt.Errorf("%w: %v", errFileCreate, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("failed to close file: %v", cerr)
		}
	}()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer func() {
		if cerr := encoder.Close(); cerr != nil {
			log.Printf("failed to close encoder: %v", cerr)
		}
	}()

	if err := encoder.Encode(graph); err != nil {
		tracer.ExitError("cli.saveGraph", err)
		return fmt.Errorf("%w: %v", errYAMLSerialization, err)
	}

	tracer.ExitSuccess("cli.saveGraph")
	return nil
}
