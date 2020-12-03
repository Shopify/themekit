package lint

import (
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
)

func TestPathToProject(t *testing.T) {
	engine := liquid.NewEngine()
	template := `{{ 'customer.order.title' | t: name: order.name }}`
	bindings := map[string]interface{}{}
	engine.RegisterFilter("t", func(value interface{}) interface{} {
		return value
	})
	_, err := engine.ParseAndRenderString(template, bindings)
	assert.Nil(t, err)
}
