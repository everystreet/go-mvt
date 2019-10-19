package geometry_test

import (
	"math"
	"testing"

	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
	"github.com/stretchr/testify/require"
)

func TestMakeParameterInteger(t *testing.T) {
	t.Run("min value", func(t *testing.T) {
		value := int32(math.Pow(2, 31)-1) * -1
		param, err := geometry.MakeParameterInteger(value)
		require.NoError(t, err)
		require.Equal(t, value, param.Value())
	})

	t.Run("max value", func(t *testing.T) {
		value := int32(math.Pow(2, 31) - 1)
		param, err := geometry.MakeParameterInteger(value)
		require.NoError(t, err)
		require.Equal(t, value, param.Value())
	})

	t.Run("min value exceeded", func(t *testing.T) {
		value := int32(math.Pow(2, 31)) * -1
		_, err := geometry.MakeParameterInteger(value)
		require.Error(t, err)
	})

	t.Run("max value exceeded", func(t *testing.T) {
		value := int32(math.Pow(2, 31))
		_, err := geometry.MakeParameterInteger(value)
		require.Error(t, err)
	})
}
