package functions

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Options struct {
	Resources *ResourcesOptions `yaml:"resources,omitempty"`
}

// On the min/max tags below: maximum is int32 max because both deployers narrow
// min/max to the int32 Kubernetes replica count, so ValidateScale rejects
// anything larger -- the schema is kept in step. This block is deliberately
// not a field doc comment (a blank line separates it from ScaleOptions) so it
// does not leak into the generated schema as a description.

type ScaleOptions struct {
	Min  *int64            `yaml:"min,omitempty" jsonschema:"minimum=0,maximum=2147483647"`
	Max  *int64            `yaml:"max,omitempty" jsonschema:"minimum=0,maximum=2147483647"`
	KEDA *KEDAScaleOptions `yaml:"keda,omitempty"`
	KPA  *KPAScaleOptions  `yaml:"kpa,omitempty"`
}

type KEDAScaleOptions struct {
	// The jsonschema tag is split on commas, so a comma inside the description
	// must be escaped as \, or it ends the description early.
	PollingInterval *int32 `yaml:"pollingInterval,omitempty" jsonschema:"description=How often KEDA checks the trigger in seconds (default 30). Applies only to kafka triggers; it has no effect on an http trigger (which scales from interceptor-reported metrics and has no polling concept).,minimum=1"`
	CooldownPeriod  *int32 `yaml:"cooldownPeriod,omitempty" jsonschema:"minimum=1"`

	// triggers has no omitempty: the schema generator derives "required" from
	// its absence, matching ValidateScale (a written scale.keda requires >=1
	// trigger; a nil scale.keda defaults to the http scaler) so scale: {keda: {}}
	// is rejected at schema time too. KEDAScaleOptions is only
	// ever serialized for deployer: keda, where triggers is always populated.
	// (Blank line above keeps this note out of the generated schema description.)

	Triggers []KEDATrigger `yaml:"triggers" jsonschema:"minItems=1"`
}

type KEDATrigger struct {
	Type                   string `yaml:"type" jsonschema:"enum=http,enum=kafka"`
	TargetValue            *int64 `yaml:"targetValue,omitempty" jsonschema:"minimum=1"`
	LagThreshold           *int64 `yaml:"lagThreshold,omitempty" jsonschema:"minimum=1"`
	ActivationLagThreshold *int64 `yaml:"activationLagThreshold,omitempty" jsonschema:"minimum=0"`
}

type KPAScaleOptions struct {
	Metric      *string  `yaml:"metric,omitempty" jsonschema:"enum=concurrency,enum=rps"`
	Target      *float64 `yaml:"target,omitempty" jsonschema:"minimum=0.01"`
	Utilization *float64 `yaml:"utilization,omitempty" jsonschema:"minimum=1,maximum=100"`
}

type ResourcesOptions struct {
	Requests *ResourcesRequestsOptions `yaml:"requests,omitempty"`
	Limits   *ResourcesLimitsOptions   `yaml:"limits,omitempty"`
}

type ResourcesLimitsOptions struct {
	CPU         *string `yaml:"cpu,omitempty" jsonschema:"pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"`
	Memory      *string `yaml:"memory,omitempty" jsonschema:"pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"`
	Concurrency *int64  `yaml:"concurrency,omitempty" jsonschema:"minimum=0"`
}

type ResourcesRequestsOptions struct {
	CPU    *string `yaml:"cpu,omitempty" jsonschema:"pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"`
	Memory *string `yaml:"memory,omitempty" jsonschema:"pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"`
}

// validateOptions checks that input Options are correctly set.
// Scale validation is handled separately by ValidateScale.
// Returns array of error messages, empty if no errors are found
func validateOptions(options Options) (errors []string) {

	// options.resource
	if options.Resources != nil {

		// options.resource.requests
		if options.Resources.Requests != nil {

			if options.Resources.Requests.CPU != nil {
				_, err := resource.ParseQuantity(*options.Resources.Requests.CPU)
				if err != nil {
					errors = append(errors, fmt.Sprintf("options field \"resources.requests.cpu\" has invalid value set: \"%s\"; \"%s\"",
						*options.Resources.Requests.CPU, err.Error()))
				}
			}

			if options.Resources.Requests.Memory != nil {
				_, err := resource.ParseQuantity(*options.Resources.Requests.Memory)
				if err != nil {
					errors = append(errors, fmt.Sprintf("options field \"resources.requests.memory\" has invalid value set: \"%s\"; \"%s\"",
						*options.Resources.Requests.Memory, err.Error()))
				}
			}
		}

		// options.resource.limits
		if options.Resources.Limits != nil {

			if options.Resources.Limits.CPU != nil {
				_, err := resource.ParseQuantity(*options.Resources.Limits.CPU)
				if err != nil {
					errors = append(errors, fmt.Sprintf("options field \"resources.limits.cpu\" has invalid value set: \"%s\"; \"%s\"",
						*options.Resources.Limits.CPU, err.Error()))
				}
			}

			if options.Resources.Limits.Memory != nil {
				_, err := resource.ParseQuantity(*options.Resources.Limits.Memory)
				if err != nil {
					errors = append(errors, fmt.Sprintf("options field \"resources.limits.memory\" has invalid value set: \"%s\"; \"%s\"",
						*options.Resources.Limits.Memory, err.Error()))
				}
			}

			if options.Resources.Limits.Concurrency != nil {
				if *options.Resources.Limits.Concurrency < 0 {
					errors = append(errors, fmt.Sprintf("options field \"resources.limits.concurrency\" has value set to \"%d\", but it must not be less than 0",
						*options.Resources.Limits.Concurrency))
				}
			}
		}
	}

	return
}
