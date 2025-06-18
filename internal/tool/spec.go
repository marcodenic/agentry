package tool

import "github.com/marcodenic/agentry/internal/model"

// BuildSpecs converts a Registry into model.ToolSpec definitions.
func BuildSpecs(reg Registry) []model.ToolSpec {
	specs := make([]model.ToolSpec, 0, len(reg))
	for _, t := range reg {
		specs = append(specs, model.ToolSpec{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.JSONSchema(),
		})
	}
	return specs
}
