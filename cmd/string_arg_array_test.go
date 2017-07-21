package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArgArray_String(t *testing.T) {
	saa := stringArgArray{
		values: []string{"beedrill", "pikachu"},
	}
	assert.Equal(t, "beedrill,pikachu", saa.String())
}

func TestStringArgArray_Set(t *testing.T) {
	saa := stringArgArray{
		values: []string{"vaporeon"},
	}
	saa.Set("drowzee")
	assert.Equal(t, "vaporeon,drowzee", saa.String())
	saa.Set("")
	assert.Equal(t, "vaporeon,drowzee", saa.String())
}

func TestStringArgArray_Type(t *testing.T) {
	saa := stringArgArray{
		values: []string{},
	}
	assert.Equal(t, "string", saa.Type())
}

func TestStringArgArray_Value(t *testing.T) {
	saa := stringArgArray{
		values: []string{},
	}
	assert.Nil(t, saa.Value())
	saa.Set("raichu")
	assert.Equal(t, "raichu", saa.Value()[0])
}
