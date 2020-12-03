package lint

import (
	"github.com/osteele/liquid/render"
)

func Lint(root string) error {
	cfg := render.NewConfig()
	shopifyFilters(&cfg, AllChecks)
	newTheme(&cfg, root, AllChecks)
	return nil
}
