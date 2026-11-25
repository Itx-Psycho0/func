package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestMain runs the tests from the repository root, which is the working
// directory the generator is run from: it reads the docstrings under
// ./pkg/functions/.
func TestMain(m *testing.M) {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Chdir(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// repoRoot walks up from the working directory to the module root.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above %s", dir)
		}
		dir = parent
	}
}

// TestNumericBounds checks that the numeric bounds declared in jsonschema
// struct tags reach the generated schema unchanged.
//
// The previous generator (github.com/alecthomas/jsonschema) parsed minimum as
// an int and omitted a zero minimum, so KPAScaleOptions.target published
// "minimum": 0 instead of 0.01 and every minimum=0 had to be declared through
// jsonschema_extras. The verbatim JSON token is compared rather than a decoded
// number, so an integral bound cannot pass for a fractional one.
func TestNumericBounds(t *testing.T) {
	schema, err := funcYamlSchema()
	if err != nil {
		t.Fatalf("cannot generate schema: %s", err)
	}

	var doc struct {
		Defs map[string]struct {
			Properties map[string]map[string]json.RawMessage `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("cannot parse schema: %s", err)
	}

	tests := []struct {
		definition string
		property   string
		keyword    string
		want       string
	}{
		// fractional minimum, enforced as >= 0.01 by validateKPAScale
		{"KPAScaleOptions", "target", "minimum", "0.01"},
		{"KPAScaleOptions", "utilization", "minimum", "1"},
		{"KPAScaleOptions", "utilization", "maximum", "100"},
		// zero minimums must not be dropped
		{"ScaleOptions", "min", "minimum", "0"},
		{"ScaleOptions", "min", "maximum", "2147483647"},
		{"ScaleOptions", "max", "minimum", "0"},
		{"ScaleOptions", "max", "maximum", "2147483647"},
		{"KEDATrigger", "activationLagThreshold", "minimum", "0"},
		{"ResourcesLimitsOptions", "concurrency", "minimum", "0"},
		{"DeliverySpec", "retry", "minimum", "0"},
		// minimum declared next to a description in the same tag
		{"KEDAScaleOptions", "pollingInterval", "minimum", "1"},
		{"KEDAScaleOptions", "cooldownPeriod", "minimum", "1"},
		{"KEDATrigger", "targetValue", "minimum", "1"},
		{"KEDATrigger", "lagThreshold", "minimum", "1"},
	}
	for _, test := range tests {
		got := string(doc.Defs[test.definition].Properties[test.property][test.keyword])
		if got != test.want {
			t.Errorf("%s.%s %s = %s, want %s", test.definition, test.property, test.keyword, got, test.want)
		}
	}

	// A boolean exclusiveMinimum (draft-04) or a numeric one (2020-12) would
	// make target > 0.01 required, while validateKPAScale accepts exactly 0.01.
	if got, ok := doc.Defs["KPAScaleOptions"].Properties["target"]["exclusiveMinimum"]; ok {
		t.Errorf("KPAScaleOptions.target exclusiveMinimum = %s, want absent: validateKPAScale accepts target >= 0.01", got)
	}
}

// TestDescriptionsAreComplete checks that a description declared in a
// jsonschema tag is not cut short by the tag's other keywords.
func TestDescriptionsAreComplete(t *testing.T) {
	schema, err := funcYamlSchema()
	if err != nil {
		t.Fatalf("cannot generate schema: %s", err)
	}

	var doc struct {
		Defs map[string]struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("cannot parse schema: %s", err)
	}

	tests := []struct {
		definition, property, want string
	}{
		{"KEDAScaleOptions", "pollingInterval", "How often KEDA checks the trigger in seconds (default 30). Applies only to kafka triggers; it has no effect on an http trigger (which scales from interceptor-reported metrics and has no polling concept)."},
		{"DeliverySpec", "retry", "Retries after the first delivery attempt. Omit for unlimited; 0 means a single attempt with no retry."},
	}
	for _, test := range tests {
		got := doc.Defs[test.definition].Properties[test.property].Description
		if got != test.want {
			t.Errorf("%s.%s description = %q, want %q", test.definition, test.property, got, test.want)
		}
	}
}
