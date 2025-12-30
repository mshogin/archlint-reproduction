// Package sample provides a sample calculator for integration testing.
package sample

import (
	"github.com/mshogin/archlint/pkg/tracer"
)

// Calculator is a simple calculator with memory.
type Calculator struct {
	memory int
}

// NewCalculator creates a new Calculator instance.
func NewCalculator() *Calculator {
	tracer.Enter("sample.NewCalculator")

	c := &Calculator{memory: 0}

	tracer.ExitSuccess("sample.NewCalculator")
	return c
}

// Add adds two numbers.
func Add(a, b int) int {
	tracer.Enter("sample.Add")

	result := a + b

	tracer.ExitSuccess("sample.Add")
	return result
}

// Multiply multiplies two numbers.
func Multiply(a, b int) int {
	tracer.Enter("sample.Multiply")

	result := a * b

	tracer.ExitSuccess("sample.Multiply")
	return result
}

// AddToMemory adds a value to memory.
func (c *Calculator) AddToMemory(value int) {
	tracer.Enter("sample.Calculator.AddToMemory")

	c.memory += value

	tracer.ExitSuccess("sample.Calculator.AddToMemory")
}

// GetMemory returns the current memory value.
func (c *Calculator) GetMemory() int {
	tracer.Enter("sample.Calculator.GetMemory")

	result := c.memory

	tracer.ExitSuccess("sample.Calculator.GetMemory")
	return result
}

// Calculate performs a calculation using Add and Multiply.
func (c *Calculator) Calculate(a, b int) int {
	tracer.Enter("sample.Calculator.Calculate")

	sum := Add(a, b)
	product := Multiply(a, b)
	result := sum + product

	c.AddToMemory(result)

	tracer.ExitSuccess("sample.Calculator.Calculate")
	return result
}
