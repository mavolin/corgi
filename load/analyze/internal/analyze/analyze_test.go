package analyze

import (
	"slices"
	"strings"
	"testing"
)

func TestMethodsRunOnce(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	root := getExecutionOrder(analyzer)

	for name := range analyzer.methods {
		methodOrder := root.find(name)
		if methodOrder == nil {
			t.Errorf("%s: function is exported but not found in execution order", name)
		}
	}

	seen := make(map[string]bool)

	testMethodsRunOnce(t, root, seen)
}

func testMethodsRunOnce(t *testing.T, o *executionOrder, seen map[string]bool) {
	if strings.HasPrefix(o.name, "Complex") {
		return
	}

	if seen[o.name] {
		t.Errorf("%s: called multiple times in the execution order", o.name)
	}
	seen[o.name] = true

	for _, call := range o.calls {
		testMethodsRunOnce(t, call, seen)
	}
}

// ============================================================================
// Depends on Checks
// ======================================================================================

// TestCheckDependenciesFulfilled tests that all
// 'Depends on Checks'-dependencies are fulfilled, i.e. that the function being
// depended on is run before the function depending on it.
func TestCheckDependenciesFulfilled(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	root := getExecutionOrder(analyzer)

	for name, method := range analyzer.methods {
		for _, dep := range method.checkDependencies {
			if root.callsBefore(dep, name) {
				continue
			}
			t.Errorf("Function %s depends on %s, but %[2]s runs after %[1]s in execution order or not at all", name, dep)
		}
	}
}

func TestCheckDependenciesExist(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	for name, method := range analyzer.methods {
		for _, dep := range method.checkDependencies {
			if _, ok := analyzer.methods[dep]; !ok {
				t.Errorf("%s: depens upon %s, but %[2]s is not defined", name, dep)
			}
		}
	}
}

// ============================================================================
// Sets Fields
// ======================================================================================

func TestFieldsSetOnce(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	root := getExecutionOrder(analyzer)
	seen := make(map[string]bool)

	testFieldsSetOnce(t, root, seen)
}

func testFieldsSetOnce(t *testing.T, o *executionOrder, seen map[string]bool) {
	for _, field := range o.method.setsFields {
		key := o.name + "/" + field
		if seen[field] {
			t.Errorf("%s: field set multiple times by different methods", field)
		}
		seen[key] = true
	}

	for _, call := range o.calls {
		testFieldsSetOnce(t, call, seen)
	}
}

func TestSetFieldsExist(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	fields, err := getAllFields()
	if err != nil {
		t.Fatalf("Failed to get all fields: %v", err)
	}

	for name, info := range analyzer.methods {
		for _, field := range info.setsFields {
			if !slices.Contains(fields, field) {
				t.Errorf("Field %s is set by method %s but does not exist", field, name)
			}
		}
	}
}

func TestAllFieldsSet(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	fields, err := getAllFields()
	if err != nil {
		t.Fatalf("Failed to get all fields: %v", err)
	}

	for _, info := range analyzer.methods {
		for _, field := range info.setsFields {
			if i := slices.Index(fields, field); i >= 0 {
				fields[i] = ""
			}
		}
	}

	for _, field := range fields {
		if field == "" {
			continue
		}
		t.Errorf("%s: field not set by any method", field)
	}
}

// ============================================================================
// Depends on Fields
// ======================================================================================

// TestCheckDependenciesFields tests that all 'Depends on Fields'-dependencies
// are fulfilled, i.e. that the field being depended on is set before the
// function depending on it is called.
func TestFieldDependenciesFulfilled(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	root := getExecutionOrder(analyzer)

	for name, method := range analyzer.methods {
		for _, dep := range method.fieldDependencies {
			testFieldDependenciesFulfilled(t, dep, root, name)
		}
	}
}

func testFieldDependenciesFulfilled(t *testing.T, dep string, start *executionOrder, dependent string) (ok bool) {
	if start.name == dependent {
		t.Errorf("%s: depends on %s, but is not set before in execution order", dependent, dep)
	}

	for _, field := range start.method.setsFields {
		if field == dep {
			return true
		}
	}

	for _, call := range start.calls {
		if testFieldDependenciesFulfilled(t, dep, call, dependent) {
			return true
		}
	}

	return false
}

func TestDependedOnFieldsExist(t *testing.T) {
	analyzer, err := parseAnalyzerMethods()
	if err != nil {
		t.Fatalf("Failed to parse analyzer methods: %v", err)
	}

	fields, err := getAllFields()
	if err != nil {
		t.Fatalf("Failed to get all fields: %v", err)
	}

	for name, info := range analyzer.methods {
		for _, field := range info.fieldDependencies {
			if !slices.Contains(fields, field) {
				t.Errorf("%s: depends on %s but %[2]s does not exist", name, field)
			}
		}
	}
}
