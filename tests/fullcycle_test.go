// Package tests provides integration tests for archlint.
package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mshogin/archlint/internal/analyzer"
	"github.com/mshogin/archlint/pkg/tracer"
)

// TestFullCycle tests the complete workflow: analyze -> trace -> contexts.
func TestFullCycle(t *testing.T) {
	outputDir := filepath.Join("output")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Step 1: Analyze sample code.
	t.Run("Analyze", func(t *testing.T) {
		goAnalyzer := analyzer.NewGoAnalyzer()
		sampleDir := filepath.Join("testdata", "sample")

		graph, err := goAnalyzer.Analyze(sampleDir)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if len(graph.Nodes) == 0 {
			t.Error("expected nodes in graph")
		}

		if len(graph.Edges) == 0 {
			t.Error("expected edges in graph")
		}

		t.Logf("Found %d components and %d links", len(graph.Nodes), len(graph.Edges))
	})

	// Step 2: Run traced execution.
	t.Run("TracedExecution", func(t *testing.T) {
		tracer.StartTrace("TestFullCycle_TracedExecution")

		// Simulate some function calls that would be traced.
		tracer.Enter("tests.simulatedWork")
		tracer.ExitSuccess("tests.simulatedWork")

		tracePath := filepath.Join(outputDir, "fullcycle_trace.json")

		if err := tracer.Save(tracePath); err != nil {
			t.Fatalf("failed to save trace: %v", err)
		}

		tracer.StopTrace()

		if _, err := os.Stat(tracePath); os.IsNotExist(err) {
			t.Errorf("trace file not created: %s", tracePath)
		}
	})

	// Step 3: Generate contexts from traces.
	t.Run("GenerateContexts", func(t *testing.T) {
		contexts, err := tracer.GenerateContextsFromTraces(outputDir)
		if err != nil {
			t.Fatalf("GenerateContextsFromTraces failed: %v", err)
		}

		if len(contexts) == 0 {
			t.Error("expected at least one context")
		}

		for id, ctx := range contexts {
			if id == "" {
				t.Error("context ID should not be empty")
			}

			if ctx.Title == "" {
				t.Error("context Title should not be empty")
			}

			t.Logf("Generated context: %s - %s", id, ctx.Title)
		}
	})
}

// TestAnalyzerOutput verifies analyzer produces expected structure.
func TestAnalyzerOutput(t *testing.T) {
	goAnalyzer := analyzer.NewGoAnalyzer()
	sampleDir := filepath.Join("testdata", "sample")

	graph, err := goAnalyzer.Analyze(sampleDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Check for Calculator component.
	foundCalculator := false

	for _, node := range graph.Nodes {
		if node.Title == "Calculator" {
			foundCalculator = true

			break
		}
	}

	if !foundCalculator {
		t.Error("expected Calculator component in graph")
	}

	// Check for function components.
	expectedFunctions := []string{"Add", "Multiply", "NewCalculator"}
	foundFunctions := make(map[string]bool)

	for _, node := range graph.Nodes {
		for _, fn := range expectedFunctions {
			if node.Title == fn {
				foundFunctions[fn] = true
			}
		}
	}

	for _, fn := range expectedFunctions {
		if !foundFunctions[fn] {
			t.Errorf("expected function %s in graph", fn)
		}
	}
}

// TestTracerRoundTrip verifies trace can be saved and loaded.
func TestTracerRoundTrip(t *testing.T) {
	outputDir := filepath.Join("output")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	tracer.StartTrace("TestTracerRoundTrip")

	tracer.Enter("tests.function1")
	tracer.Enter("tests.function2")
	tracer.ExitSuccess("tests.function2")
	tracer.ExitSuccess("tests.function1")

	tracePath := filepath.Join(outputDir, "roundtrip_trace.json")

	if err := tracer.Save(tracePath); err != nil {
		t.Fatalf("failed to save trace: %v", err)
	}

	tracer.StopTrace()

	// Load and verify.
	trace, err := tracer.LoadTrace(tracePath)
	if err != nil {
		t.Fatalf("failed to load trace: %v", err)
	}

	if trace.TestName != "TestTracerRoundTrip" {
		t.Errorf("TestName = %s, want TestTracerRoundTrip", trace.TestName)
	}

	if len(trace.Calls) != 4 { // 2 enters + 2 exits
		t.Errorf("expected 4 calls, got %d", len(trace.Calls))
	}
}
