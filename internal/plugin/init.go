package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitTool scaffolds a builtin tool plugin in a new directory named after the tool.
func InitTool(name string) error {
	if name == "" {
		return fmt.Errorf("tool name required")
	}
	if err := os.MkdirAll(name, 0755); err != nil {
		return err
	}
	goFile := filepath.Join(name, name+".go")
	yamlFile := filepath.Join(name, name+".yaml")

	goSrc := fmt.Sprintf(`package %s

import (
    "context"

    "github.com/marcodenic/agentry/internal/tool"
)

func init() {
    tool.Register("%s", "%s tool", nil, Exec)
}

// Exec executes the %s tool.
func Exec(ctx context.Context, args map[string]any) (string, error) {
    return "", nil
}
`, name, name, name, name)

	yamlSrc := fmt.Sprintf("name: %s\ndescription: %s tool\ntype: builtin\n", name, name)

	if err := os.WriteFile(goFile, []byte(goSrc), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(yamlFile, []byte(yamlSrc), 0644); err != nil {
		return err
	}
	return nil
}
