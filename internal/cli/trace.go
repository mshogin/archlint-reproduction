package cli

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mshogin/archlint/pkg/tracer"
)

var (
	errTraceDirNotExist = errors.New("trace directory does not exist")
	errNoTraceFiles     = errors.New("no trace files found")
)

var traceOutputFile string

var traceCmd = &cobra.Command{
	Use:   "trace [trace directory]",
	Short: "Generate contexts from test traces",
	Long: `Analyzes JSON trace files and generates DocHub contexts.

Each trace represents one execution flow (test) and is converted to:
1. DocHub context with component list
2. PlantUML sequence diagram`,
	Args: cobra.ExactArgs(1),
	RunE: runTrace,
}

func init() {
	traceCmd.Flags().StringVarP(&traceOutputFile, "output", "o",
		"contexts.yaml", "Output YAML file for contexts")
	rootCmd.AddCommand(traceCmd)
}

func runTrace(cmd *cobra.Command, args []string) error {
	tracer.Enter("cli.runTrace")

	traceDir := args[0]

	if _, err := os.Stat(traceDir); os.IsNotExist(err) {
		tracer.ExitError("cli.runTrace", errTraceDirNotExist)
		return fmt.Errorf("%w: %s", errTraceDirNotExist, traceDir)
	}

	fmt.Printf("Generating contexts from traces: %s\n", traceDir)

	contexts, err := tracer.GenerateContextsFromTraces(traceDir)
	if err != nil {
		tracer.ExitError("cli.runTrace", err)
		return fmt.Errorf("failed to generate contexts: %w", err)
	}

	if err := saveContexts(contexts); err != nil {
		tracer.ExitError("cli.runTrace", err)
		return err
	}

	printContextsInfo(contexts)

	fmt.Printf("Contexts saved to %s\n", traceOutputFile)

	tracer.ExitSuccess("cli.runTrace")
	return nil
}

func saveContexts(contexts tracer.Contexts) error {
	tracer.Enter("cli.saveContexts")

	wrapper := struct {
		Contexts tracer.Contexts `yaml:"contexts"`
	}{
		Contexts: contexts,
	}

	file, err := os.Create(traceOutputFile)
	if err != nil {
		tracer.ExitError("cli.saveContexts", err)
		return fmt.Errorf("failed to create file: %w", err)
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

	if err := encoder.Encode(wrapper); err != nil {
		tracer.ExitError("cli.saveContexts", err)
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	tracer.ExitSuccess("cli.saveContexts")
	return nil
}

func printContextsInfo(contexts tracer.Contexts) {
	tracer.Enter("cli.printContextsInfo")

	fmt.Printf("Generated %d contexts:\n", len(contexts))
	for id, ctx := range contexts {
		fmt.Printf("  - %s: %d components\n", id, len(ctx.Components))
	}

	tracer.ExitSuccess("cli.printContextsInfo")
}
