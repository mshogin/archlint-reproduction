package tracer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Context represents a DocHub context.
type Context struct {
	Title        string     `yaml:"title"`
	Location     string     `yaml:"location,omitempty"`
	Presentation string     `yaml:"presentation,omitempty"`
	ExtraLinks   bool       `yaml:"extra-links,omitempty"`
	Components   []string   `yaml:"components"`
	UML          *UMLConfig `yaml:"uml,omitempty"`
}

// UMLConfig holds UML diagram configuration.
type UMLConfig struct {
	Before string `yaml:"$before,omitempty"`
	After  string `yaml:"$after,omitempty"`
	File   string `yaml:"file,omitempty"`
}

// Contexts is a map of context ID to Context.
type Contexts map[string]Context

// SequenceDiagram represents a PlantUML sequence diagram.
type SequenceDiagram struct {
	TestName     string
	Participants []string
	Calls        []SequenceCall
}

// SequenceCall represents a call in the sequence diagram.
type SequenceCall struct {
	From    string
	To      string
	Success bool
	Error   string
}

// GenerateContextsFromTraces generates contexts from trace files in a directory.
func GenerateContextsFromTraces(traceDir string) (Contexts, error) {
	contexts := make(Contexts)

	files, err := filepath.Glob(filepath.Join(traceDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob trace files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no trace files found in %s", traceDir)
	}

	for _, file := range files {
		trace, err := LoadTrace(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load trace %s: %w", file, err)
		}

		ctx, err := GenerateContextFromTrace(trace)
		if err != nil {
			return nil, fmt.Errorf("failed to generate context from %s: %w", file, err)
		}

		baseName := strings.TrimSuffix(filepath.Base(file), ".json")
		pumlFile := filepath.Join(traceDir, baseName+".puml")

		if err := GenerateSequenceDiagram(trace, pumlFile); err != nil {
			return nil, fmt.Errorf("failed to generate diagram for %s: %w", file, err)
		}

		ctx.UML = &UMLConfig{
			File: pumlFile,
		}

		contextID := "tests." + sanitizeContextID(trace.TestName)
		contexts[contextID] = ctx
	}

	return contexts, nil
}

// LoadTrace loads a trace from a JSON file.
func LoadTrace(filename string) (*Trace, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var trace Trace
	if err := json.Unmarshal(data, &trace); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &trace, nil
}

// GenerateContextFromTrace generates a context from a single trace.
func GenerateContextFromTrace(trace *Trace) (Context, error) {
	componentsMap := make(map[string]bool)

	for _, call := range trace.Calls {
		if call.Event == "enter" {
			hierarchicalID := toHierarchicalComponentID(call.Function)
			componentsMap[hierarchicalID] = true
		}
	}

	components := make([]string, 0, len(componentsMap))
	for id := range componentsMap {
		components = append(components, id)
	}
	sort.Strings(components)

	return Context{
		Title:        humanizeTestName(trace.TestName),
		Location:     "Tests/" + humanizeTestName(trace.TestName),
		Presentation: "plantuml",
		ExtraLinks:   false,
		Components:   components,
	}, nil
}

// GenerateSequenceDiagram generates a PlantUML sequence diagram from a trace.
func GenerateSequenceDiagram(trace *Trace, outputFile string) error {
	diagram := buildSequenceDiagram(trace)
	puml := generatePlantUML(diagram)

	if err := os.WriteFile(outputFile, []byte(puml), 0o600); err != nil {
		return fmt.Errorf("failed to write diagram: %w", err)
	}

	return nil
}

func buildSequenceDiagram(trace *Trace) *SequenceDiagram {
	diagram := &SequenceDiagram{
		TestName:     trace.TestName,
		Participants: []string{},
		Calls:        []SequenceCall{},
	}

	participantSet := make(map[string]bool)
	callStack := []string{}

	for _, call := range trace.Calls {
		shortName := shortName(call.Function)
		alias := sanitizeAlias(call.Function)

		if !participantSet[alias] {
			participantSet[alias] = true
			diagram.Participants = append(diagram.Participants, alias+"|"+shortName)
		}

		switch call.Event {
		case "enter":
			callStack = append(callStack, alias)
		case "exit_success", "exit_error":
			if len(callStack) > 1 {
				from := callStack[len(callStack)-2]
				to := callStack[len(callStack)-1]
				diagram.Calls = append(diagram.Calls, SequenceCall{
					From:    from,
					To:      to,
					Success: call.Event == "exit_success",
					Error:   call.Error,
				})
			}
			if len(callStack) > 0 {
				callStack = callStack[:len(callStack)-1]
			}
		}
	}

	return diagram
}

func generatePlantUML(diagram *SequenceDiagram) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")
	sb.WriteString(fmt.Sprintf("title %s\n\n", diagram.TestName))

	for _, p := range diagram.Participants {
		parts := strings.Split(p, "|")
		alias := parts[0]
		name := parts[1]
		sb.WriteString(fmt.Sprintf("participant \"%s\" as %s\n", name, alias))
	}
	sb.WriteString("\n")

	for _, call := range diagram.Calls {
		if call.Success {
			sb.WriteString(fmt.Sprintf("%s -> %s\n", call.From, call.To))
		} else {
			sb.WriteString(fmt.Sprintf("%s -> %s: <color:red>error</color>\n", call.From, call.To))
		}
	}

	sb.WriteString("\n@enduml\n")

	return sb.String()
}

// humanizeTestName converts TestProcessOrder to "Process Order".
func humanizeTestName(testName string) string {
	name := strings.TrimPrefix(testName, "Test")

	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}

	return strings.TrimSpace(result.String())
}

// sanitizeContextID converts TestProcessOrder to "test-process-order".
func sanitizeContextID(testName string) string {
	return strings.ToLower(camelToSnake(testName))
}

// sanitizeAlias converts a function name to a valid PlantUML alias.
func sanitizeAlias(name string) string {
	result := strings.ReplaceAll(name, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "-", "_")
	return result
}

// shortName extracts the short name from a fully qualified function name.
func shortName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

// toHierarchicalComponentID converts OrderService.ProcessOrder to order_service.process_order.
func toHierarchicalComponentID(functionName string) string {
	parts := strings.Split(functionName, ".")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = camelToSnake(part)
	}
	return strings.Join(result, ".")
}

// camelToSnake converts CamelCase to snake_case.
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// MatchComponentPattern matches a component ID against a pattern with wildcards.
// Patterns: exact match, single level (*), recursive (**).
func MatchComponentPattern(componentID, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return componentID == pattern
	}

	if strings.HasSuffix(pattern, ".**") {
		prefix := strings.TrimSuffix(pattern, ".**")
		return strings.HasPrefix(componentID, prefix+".")
	}

	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		if !strings.HasPrefix(componentID, prefix+".") {
			return false
		}
		rest := strings.TrimPrefix(componentID, prefix+".")
		return !strings.Contains(rest, ".")
	}

	regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", "[^.]*")
	regex, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		return false
	}

	return regex.MatchString(componentID)
}

// ExpandComponentPatterns expands patterns with wildcards into matching component IDs.
func ExpandComponentPatterns(patterns, allComponents []string) []string {
	result := make(map[string]bool)

	for _, pattern := range patterns {
		if !strings.Contains(pattern, "*") {
			result[pattern] = true
			continue
		}

		for _, comp := range allComponents {
			if MatchComponentPattern(comp, pattern) {
				result[comp] = true
			}
		}
	}

	expanded := make([]string, 0, len(result))
	for comp := range result {
		expanded = append(expanded, comp)
	}
	sort.Strings(expanded)

	return expanded
}
