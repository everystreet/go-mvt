package geometry_test

import (
	"math"
	"testing"

	"github.com/everystreet/go-mvt/internal/geometry"
	"github.com/stretchr/testify/require"
)

func TestMakeCommandInteger(t *testing.T) {
	t.Run("min count", func(t *testing.T) {
		cmd, err := geometry.MakeCommandInteger(geometry.LineTo, 0)
		require.NoError(t, err)
		require.Equal(t, geometry.CommandInteger(geometry.LineTo), geometry.CommandInteger(cmd))
		require.Equal(t, geometry.LineTo, cmd.ID())
		require.Equal(t, 0, int(cmd.Count()))
	})

	t.Run("max count", func(t *testing.T) {
		count := uint32(math.Pow(2, 29) - 1)
		cmd, err := geometry.MakeCommandInteger(geometry.LineTo, count)
		require.NoError(t, err)
		require.Equal(t, geometry.LineTo, cmd.ID())
		require.Equal(t, int(count), int(cmd.Count()))
	})

	t.Run("max count exceeded", func(t *testing.T) {
		count := uint32(math.Pow(2, 29))
		_, err := geometry.MakeCommandInteger(geometry.LineTo, count)
		require.Error(t, err)
	})
}
