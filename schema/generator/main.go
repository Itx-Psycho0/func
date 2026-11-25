package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	fn "knative.dev/func/pkg/functions"
)

// This helper application generates json schemas:
// - schema for func.yaml stored in schema/func_yaml-schema.json
func main() {
	err := generateFuncYamlSchema()
	if err != nil {
		panic(err)
	}
}

// generateFuncYamlSchema generates json schema for function configuration file - func.yaml.
// Generated schema is written into schema/func_yaml-schema.json file
func generateFuncYamlSchema() error {
	schema, err := funcYamlSchema()
	if err != nil {
		return err
	}

	// write schema to the file
	return os.WriteFile("schema/func_yaml-schema.json", schema, 0644)
}

// funcYamlSchema generates the json schema for func.yaml exactly as it is
// written to schema/func_yaml-schema.json, indentation and trailing newline
// included. It is separated from generateFuncYamlSchema so the schema can be
// verified without writing to the working tree.
func funcYamlSchema() ([]byte, error) {
	r := &jsonschema.Reflector{
		// func.yaml fields are named by their yaml tags; Function has no
		// json tags.
		FieldNameTag: "yaml",
		// Do not emit an $id derived from the Go package path: it is not
		// where the schema is published.
		Anonymous: true,
	}

	if err := r.AddGoComments("knative.dev/func", "./pkg/functions/"); err != nil {
		return nil, fmt.Errorf("cannot parse docstrings: %w", err)
	}

	schema, err := json.Marshal(r.Reflect(&fn.Function{}))
	if err != nil {
		return nil, err
	}

	// indent the generated json
	var indentedSchema bytes.Buffer
	if err = json.Indent(&indentedSchema, schema, "", "\t"); err != nil {
		return nil, err
	}

	if err = indentedSchema.WriteByte('\n'); err != nil {
		return nil, err
	}

	return indentedSchema.Bytes(), nil
}
