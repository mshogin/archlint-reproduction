# Experiment: Recreating a Project from Specifications Using Claude Code

**Experiment Date:** 2025-12-30

[Русская версия](README.ru.md)

---

## Experiment Description

### Goal

Test how accurately Claude Code can recreate an existing project using only specifications (without access to the original source code).

### Methodology

1. **Original project:** `archlint` - a tool for analyzing Go project architecture
2. **Specifications:** 10 detailed spec files in Markdown + PlantUML format (73 KB)
3. **Process:** Claude Code received an empty directory and specifications, then implemented the project from scratch
4. **Result:** `archlint-reproduction` - the recreated project

### Input Data

| Specification | Description | Size |
|---------------|-------------|------|
| [0001-init-project.md](specs/todo/0001-init-project.md) | Go module initialization | 3.6 KB |
| [0002-makefile.md](specs/todo/0002-makefile.md) | Build system | 3.3 KB |
| [0003-data-model.md](specs/todo/0003-data-model.md) | Graph/Node/Edge data model | 5.3 KB |
| [0004-go-analyzer.md](specs/todo/0004-go-analyzer.md) | Go code analyzer | 10.5 KB |
| [0005-cli-framework.md](specs/todo/0005-cli-framework.md) | CLI framework with Cobra | 4.9 KB |
| [0006-collect-command.md](specs/todo/0006-collect-command.md) | Architecture collection command | 7.3 KB |
| [0007-tracer-library.md](specs/todo/0007-tracer-library.md) | Tracing library | 11.2 KB |
| [0008-trace-command.md](specs/todo/0008-trace-command.md) | Context generation command | 11.4 KB |
| [0009-tracerlint.md](specs/todo/0009-tracerlint.md) | Linter for tracer calls | 7.5 KB |
| [0010-integration-tests.md](specs/todo/0010-integration-tests.md) | Integration tests | 8.3 KB |

### Execution Time

- Implementation of all 10 specifications: ~20 minutes
- Fixing compilation errors: ~5 minutes
- Passing tests: ~5 minutes
- **Total:** ~30 minutes

---

## Executive Summary

| Metric | Value |
|--------|-------|
| Cloning success rate | **85.5%** |
| Structural identity | **100%** |
| Semantic equivalence | **~75%** |
| Total mutations | **23** |
| Critical mutations | **3** |
| Medium mutations | **8** |
| Minor mutations | **12** |

**Verdict:** Cloning SUCCESSFUL - the project was reproduced with core functionality preserved, but with noticeable mutations in implementation details.

---

## 1. Code Statistics

### 1.1 Codebase Size

| Metric | Original | Reproduction | Difference |
|--------|----------|--------------|------------|
| Total Go code lines | 2,159 | 1,845 | -314 (-14.5%) |
| Number of .go files | 13 | 13 | 0 |
| Specifications implemented | 10 | 10 | 0 |

### 1.2 Size by File

| File | Original | Reproduction | Difference |
|------|----------|--------------|------------|
| cmd/archlint/main.go | 20 | 19 | -1 (-5%) |
| cmd/tracelint/main.go | 14 | 11 | -3 (-21%) |
| internal/analyzer/go.go | 862 | 694 | -168 (-19.5%) |
| internal/cli/collect.go | 159 | 144 | -15 (-9.4%) |
| internal/cli/root.go | 38 | 33 | -5 (-13.2%) |
| internal/cli/trace.go | 123 | 117 | -6 (-4.9%) |
| internal/linter/tracerlint.go | 362 | 292 | -70 (-19.3%) |
| internal/model/model.go | 23 | 25 | +2 (+8.7%) |
| pkg/tracer/trace.go | 166 | 180 | +14 (+8.4%) |
| pkg/tracer/context_generator.go | 392 | 330 | -62 (-15.8%) |

---

## 2. Structural Compliance

### 2.1 Directories - IDENTICAL (100%)

```
cmd/
  archlint/      [OK]
  tracelint/     [OK]
internal/
  analyzer/      [OK]
  cli/           [OK]
  linter/        [OK]
  model/         [OK]
pkg/
  tracer/        [OK]
specs/
  done/          [OK]
  inprogress/    [OK]
  todo/          [OK]
tests/
  testdata/      [OK]
```

### 2.2 Files by Specification

| Specification | Files | Status |
|---------------|-------|--------|
| 0001-init-project | go.mod, go.sum | OK |
| 0002-makefile | Makefile | OK (simplified) |
| 0003-data-model | internal/model/model.go | OK |
| 0004-go-analyzer | internal/analyzer/go.go | OK (refactored) |
| 0005-cli-framework | internal/cli/root.go | OK |
| 0006-collect-command | internal/cli/collect.go | OK |
| 0007-tracer-library | pkg/tracer/*.go | OK (extended) |
| 0008-trace-command | internal/cli/trace.go | OK |
| 0009-tracerlint | internal/linter/tracerlint.go | OK (simplified) |
| 0010-integration-tests | tests/*.go | OK |

---

## 3. Mutation Catalog

### 3.1 Critical Mutations (affect behavior)

#### M-CRIT-01: Changed buildSequenceDiagram() logic

**File:** `pkg/tracer/context_generator.go`

**Original:**
```go
if len(*callStack) > 0 {
    from := (*callStack)[len(*callStack)-1]
    diagram.Calls = append(diagram.Calls, SequenceCall{
        From:    from,
        To:      call.Function,
        Success: true,
    })
}
```

**Reproduction:**
```go
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
```

**Impact:** Changed sequence diagram building algorithm. Reproduction records calls only when stack depth > 1 and uses index-2 as source.

---

#### M-CRIT-02: isTracerExitCall() accepts Exit()

**File:** `internal/linter/tracerlint.go`

**Original:**
```go
return isTracerCall(exprStmt.X, "ExitSuccess") || isTracerCall(exprStmt.X, "ExitError")
```

**Reproduction:**
```go
return isTracerCall(stmt, "ExitSuccess") || isTracerCall(stmt, "ExitError") || isTracerCall(stmt, "Exit")
```

**Impact:** Linter in reproduction accepts deprecated `Exit()` as a valid call, original does not.

---

#### M-CRIT-03: GoAnalyzer with additional fields

**File:** `internal/analyzer/go.go`

**Original:**
```go
type GoAnalyzer struct {
    packages  map[string]*PackageInfo
    types     map[string]*TypeInfo
    functions map[string]*FunctionInfo
    methods   map[string]*MethodInfo
    nodes     []model.Node
    edges     []model.Edge
}
```

**Reproduction:**
```go
type GoAnalyzer struct {
    packages   map[string]*PackageInfo
    types      map[string]*TypeInfo
    functions  map[string]*FunctionInfo
    methods    map[string]*MethodInfo
    nodes      []model.Node
    edges      []model.Edge
    baseDir    string     // NEW
    modulePath string     // NEW
}
```

**Impact:** Reproduction determines module path from go.mod, original uses directory-based approach.

---

### 3.2 Medium Mutations (affect output/logs)

| ID | File | Description | Type |
|----|------|-------------|------|
| M-MED-01 | pkg/tracer/trace.go | Added package-level `Save()` function | Extension |
| M-MED-02 | internal/cli/*.go | Removed tracer calls from `init()` functions | Simplification |
| M-MED-03 | cmd/tracelint/main.go | Removed tracer calls completely | Simplification |
| M-MED-04 | internal/cli/root.go | Removed error output to stderr | Change |
| M-MED-05 | internal/cli/trace.go | Simplified `printContextsInfo()` output | Change |
| M-MED-06 | pkg/tracer/context_generator.go | Added component sorting | Extension |
| M-MED-07 | internal/analyzer/go.go | Methods renamed (parseTypeDecl -> parseGenDecl) | Refactoring |
| M-MED-08 | internal/linter/tracerlint.go | Simplified `isExcluded()` logic | Simplification |

### 3.3 Minor Mutations (cosmetic)

| ID | Description |
|----|-------------|
| M-MIN-01 | Comment language: Russian -> English |
| M-MIN-02 | Tracer paths with package prefix (`cli.Execute` vs `Execute`) |
| M-MIN-03 | Variable names (`excludedPackages` -> `excludePackages`) |
| M-MIN-04 | Function order in files changed |
| M-MIN-05 | Variable declaration order changed |
| M-MIN-06 | Error messages in English |
| M-MIN-07 | Removed emoji from output |
| M-MIN-08 | Makefile simplified |
| M-MIN-09 | README simplified |
| M-MIN-10 | .archlint.yaml empty (instead of 1091 lines of rules) |
| M-MIN-11 | Dependency versions updated |
| M-MIN-12 | Go version: 1.25.1 -> 1.25.4 |

---

## 4. Specification Analysis

### 4.1 Specification Compliance

| Specification | Compliance | Mutations |
|---------------|------------|-----------|
| 0001-init-project | 100% | Updated versions |
| 0002-makefile | 80% | Simplified, removed emoji |
| 0003-data-model | 100% | Only comment language |
| 0004-go-analyzer | 90% | M-CRIT-03, M-MED-07 |
| 0005-cli-framework | 95% | M-MED-04 |
| 0006-collect-command | 90% | M-MED-02 |
| 0007-tracer-library | 95% | M-MED-01 (extension) |
| 0008-trace-command | 85% | M-MED-02, M-MED-05 |
| 0009-tracerlint | 85% | M-CRIT-02, M-MED-08 |
| 0010-integration-tests | 95% | Minor changes |

### 4.2 Acceptance Criteria

Key acceptance criteria verification from specifications:

| Specification | Criterion | Status |
|---------------|-----------|--------|
| 0003 | Graph, Node, Edge types defined | PASS |
| 0003 | YAML serialization works | PASS |
| 0004 | Go package analysis works | PASS |
| 0004 | Dependency graph is built | PASS |
| 0006 | collect command generates architecture.yaml | PASS |
| 0007 | Tracer records Enter/Exit | PASS |
| 0007 | PlantUML generation works | PASS (modified format) |
| 0008 | trace command generates contexts.yaml | PASS |
| 0009 | Linter checks tracer calls | PASS (with changes) |
| 0010 | Integration tests pass | PASS |

---

## 5. Mutation Causes

### 5.1 Specification Interpretation

Claude Code interpreted specifications following their spirit, not their letter:
- Simplified code where it seemed reasonable
- Added functionality where deemed useful (Save())
- Changed algorithms while preserving the overall goal

### 5.2 Language Differences

- Original written with Russian comments and messages
- Claude Code naturally used English

### 5.3 Stylistic Preferences

- Different code organization style (function order, variables)
- Different naming approach (tracer paths)
- Different level of tracer instrumentation detail

---

## 6. Functional Testing

### 6.1 Compilation Tests

```
Original:      go build ./... - PASS
Reproduction:  go build ./... - PASS
```

### 6.2 Output Compatibility

| Artifact | Compatible | Note |
|----------|------------|------|
| architecture.yaml | Yes | Format identical |
| contexts.yaml | Partial | Different path format |
| *.puml files | Partial | Different participant format |
| Tracer JSON | Yes | Format identical |

---

## 7. Conclusions

### 7.1 Experiment Success

**Spec-driven development with Claude Code works** - from 10 specifications, a functioning project was successfully created that:
- Has identical directory structure
- Implements all required components
- Passes compilation and tests
- Performs core functions

### 7.2 Approach Limitations

1. **Implementation details vary** - the same specification can be implemented differently
2. **Algorithms are interpreted** - complex logic (buildSequenceDiagram) was implemented differently
3. **Code style differs** - organization, naming, comments
4. **Context is lost** - implicit decisions from the original are not transferred

### 7.3 Recommendations for Improving Specifications

For more accurate cloning, specifications should contain:

1. **Pseudocode for critical algorithms** - not just "what it does", but "how it does it"
2. **Input/output examples** - concrete test cases
3. **Tracer instrumentation requirements** - which functions should have tracer
4. **Code style requirements** - naming, organization
5. **Concrete acceptance tests** - test code, not just descriptions

---

## 8. Cloning Quality Metrics

| Aspect | Score | Comment |
|--------|-------|---------|
| Structure | 10/10 | Full compliance |
| Data types | 9/10 | Minor extensions |
| API/Interfaces | 8/10 | Preserved with changes |
| Algorithms | 7/10 | Critical mutations |
| Tracer integration | 7/10 | Simplified |
| Program output | 8/10 | Partially compatible |
| Tests | 9/10 | Pass |

**Overall score: 8.3/10**

---

## Reproducing the Experiment

### Prerequisites

- Claude Code (Claude Opus 4.5 was used)

### Commands

```bash
# Clone the repository
git clone https://github.com/mshogin/archlint-reproduction.git
cd archlint-reproduction

# Remove existing implementation (keep only specs)
rm -rf cmd internal pkg tests Makefile go.mod go.sum

# Run Claude Code
claude

# Give instruction:
# "Implement the project according to specifications in specs/todo/ in order of numbers"
```

### Expected Result

After execution, you should get a functioning Go project with:
- CLI commands `archlint collect` and `archlint trace`
- Linter `tracelint`
- Tracing library `pkg/tracer`
- Passing integration tests

---

## Conclusion

The experiment showed that **spec-driven development with Claude Code works**:
- Project successfully reproduced with a score of 8.3/10
- Structure and core functionality preserved at 100%
- Mutations occur in implementation details, especially in algorithms

**Main takeaway:** For accurate reproduction, specifications should contain not only "what to do", but also "how to do it" - algorithm pseudocode, input/output examples, executable acceptance tests

---

## Appendix A: File List

### Original
```
archlint/
  cmd/archlint/main.go
  cmd/tracelint/main.go
  internal/analyzer/go.go
  internal/cli/collect.go
  internal/cli/root.go
  internal/cli/trace.go
  internal/linter/tracerlint.go
  internal/model/model.go
  pkg/tracer/trace.go
  pkg/tracer/context_generator.go
  tests/fullcycle_test.go
  tests/testdata/sample/calculator.go
  tests/testdata/sample/calculator_traced_test.go
```

### Reproduction
```
archlint-reproduction/
  cmd/archlint/main.go
  cmd/tracelint/main.go
  internal/analyzer/go.go
  internal/cli/collect.go
  internal/cli/root.go
  internal/cli/trace.go
  internal/linter/tracerlint.go
  internal/model/model.go
  pkg/tracer/trace.go
  pkg/tracer/context_generator.go
  tests/fullcycle_test.go
  tests/testdata/sample/calculator.go
  tests/testdata/sample/calculator_traced_test.go
```

---

## Appendix B: Specifications

Specifications in both projects are identical (copied):

| File | Size |
|------|------|
| [0001-init-project.md](specs/todo/0001-init-project.md) | 3,675 bytes |
| [0002-makefile.md](specs/todo/0002-makefile.md) | 3,312 bytes |
| [0003-data-model.md](specs/todo/0003-data-model.md) | 5,270 bytes |
| [0004-go-analyzer.md](specs/todo/0004-go-analyzer.md) | 10,539 bytes |
| [0005-cli-framework.md](specs/todo/0005-cli-framework.md) | 4,878 bytes |
| [0006-collect-command.md](specs/todo/0006-collect-command.md) | 7,296 bytes |
| [0007-tracer-library.md](specs/todo/0007-tracer-library.md) | 11,179 bytes |
| [0008-trace-command.md](specs/todo/0008-trace-command.md) | 11,399 bytes |
| [0009-tracerlint.md](specs/todo/0009-tracerlint.md) | 7,505 bytes |
| [0010-integration-tests.md](specs/todo/0010-integration-tests.md) | 8,276 bytes |

**Total:** 73,329 bytes of specifications
