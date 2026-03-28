package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// Policy schema URL must match the embedded document's $id (used for $ref resolution).
const policySchemaURL = "https://optitrade.local/schemas/config-policy.json"

//go:embed policy.schema.json
var policySchemaJSON []byte

var (
	compiledPolicySchema *jsonschema.Schema
	compileOnce          sync.Once
	compileErr           error
)

func policySchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		var doc any
		if err := json.Unmarshal(policySchemaJSON, &doc); err != nil {
			compileErr = fmt.Errorf("internal policy schema JSON: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft2020)
		if err := c.AddResource(policySchemaURL, doc); err != nil {
			compileErr = fmt.Errorf("register policy schema: %w", err)
			return
		}
		compiledPolicySchema, compileErr = c.Compile(policySchemaURL)
	})
	return compiledPolicySchema, compileErr
}

// validatePolicyJSON validates untrusted JSON bytes against the embedded policy schema.
// Production uses this single path before unmarshalling into [Policy].
func validatePolicyJSON(data []byte) error {
	sch, err := policySchema()
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("policy file is not valid JSON: %w", err)
	}
	if err := sch.Validate(v); err != nil {
		return fmt.Errorf("policy violates schema: %w", err)
	}
	return nil
}
