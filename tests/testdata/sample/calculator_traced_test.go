// Package sample provides traced tests for the calculator.
package sample

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mshogin/archlint/pkg/tracer"
)

// TestCalculatorTraced tests calculator with tracing enabled.
func TestCalculatorTraced(t *testing.T) {
	tracer.StartTrace("TestCalculatorTraced")

	calc := NewCalculator()

	result := calc.Calculate(3, 4)

	if result != 19 { // 3+4=7, 3*4=12, 7+12=19
		t.Errorf("Calculate(3, 4) = %d, want 19", result)
	}

	memory := calc.GetMemory()

	if memory != 19 {
		t.Errorf("GetMemory() = %d, want 19", memory)
	}

	outputDir := filepath.Join("..", "..", "output")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	tracePath := filepath.Join(outputDir, "calculator_trace.json")

	if err := tracer.Save(tracePath); err != nil {
		t.Fatalf("failed to save trace: %v", err)
	}

	tracer.StopTrace()

	if _, err := os.Stat(tracePath); os.IsNotExist(err) {
		t.Errorf("trace file not created: %s", tracePath)
	}
}

// TestAddFunction tests the Add function with tracing.
func TestAddFunction(t *testing.T) {
	tracer.StartTrace("TestAddFunction")

	result := Add(10, 20)

	if result != 30 {
		t.Errorf("Add(10, 20) = %d, want 30", result)
	}

	tracer.StopTrace()
}

// TestMultiplyFunction tests the Multiply function with tracing.
func TestMultiplyFunction(t *testing.T) {
	tracer.StartTrace("TestMultiplyFunction")

	result := Multiply(5, 6)

	if result != 30 {
		t.Errorf("Multiply(5, 6) = %d, want 30", result)
	}

	tracer.StopTrace()
}
