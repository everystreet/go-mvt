package mvt21_test

import (
	"testing"

	"github.com/mercatormaps/go-geojson"

	"github.com/everystreet/go-mvt/mvt21"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalLayers(t *testing.T) {
	t.Run("multiple layers", func(t *testing.T) {
		data, err := proto.Marshal(&spec.Tile{
			Layers: []*spec.Tile_Layer{
				newLayer("layer1", 2, 4096),
				newLayer("layer2", 2, 2048),
			},
		})
		require.NoError(t, err)

		layers, err := mvt21.Unmarshal(data)
		require.NoError(t, err)
		require.Len(t, layers, 2)

		require.Contains(t, layers, mvt21.LayerName("layer1"))
		require.Equal(t, uint32(4096), layers["layer1"].Extent)

		require.Contains(t, layers, mvt21.LayerName("layer2"))
		require.Equal(t, uint32(2048), layers["layer2"].Extent)
	})

	t.Run("duplicate layer name", func(t *testing.T) {
		name, version := "layer1", uint32(2)
		data, err := proto.Marshal(&spec.Tile{
			Layers: []*spec.Tile_Layer{
				{
					Version: &version,
					Name:    &name,
				},
				{
					Version: &version,
					Name:    &name,
				},
			},
		})
		require.NoError(t, err)

		_, err = mvt21.Unmarshal(data)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
}

func TestUnmarshalMetadata(t *testing.T) {
	type check func(*testing.T, mvt21.Layer, error)

	var checks = func(cs ...check) []check { return cs }

	var (
		hasNoError = func() check {
			return func(t *testing.T, _ mvt21.Layer, err error) {
				require.NoError(t, err)
			}
		}

		hasError = func(contains string) check {
			return func(t *testing.T, _ mvt21.Layer, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), contains)
			}
		}

		hasMetadataLength = func(len int) check {
			return func(t *testing.T, layer mvt21.Layer, _ error) {
				require.Len(t, layer.Metadata, len)
			}
		}

		hasProperty = func(name string, value interface{}) check {
			return func(t *testing.T, layer mvt21.Layer, _ error) {
				require.Contains(t, layer.Metadata, geojson.Property{
					Name:  name,
					Value: value,
				})
			}
		}
	)

	for _, tt := range []struct {
		Name   string
		Keys   []string
		Values []*spec.Tile_Value
		Checks []check
	}{
		{
			Name:   "string value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newStringValue("value")},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", "value")),
		},
		{
			Name:   "float value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newFloatValue(3.142)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", float32(3.142))),
		},
		{
			Name:   "double value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newDoubleValue(3.142)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", float64(3.142))),
		},
		{
			Name:   "int value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newIntValue(-95)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", int64(-95))),
		},
		{
			Name:   "uint value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newUintValue(95)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", uint64(95))),
		},
		{
			Name:   "sint value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newSintValue(-95)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", int64(-95))),
		},
		{
			Name:   "bool value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newBoolValue(true)},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasProperty("key", true)),
		},
		{
			Name:   "key clash",
			Keys:   []string{"key", "key"},
			Values: []*spec.Tile_Value{newStringValue("value1"), newStringValue("value2")},
			Checks: checks(hasError("already exists")),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			name, version := "layer1", uint32(2)
			data, err := proto.Marshal(&spec.Tile{
				Layers: []*spec.Tile_Layer{
					{
						Version: &version,
						Name:    &name,
						Keys:    tt.Keys,
						Values:  tt.Values,
					},
				},
			})
			require.NoError(t, err)

			layers, unmarshalErr := mvt21.Unmarshal(data)

			var layer mvt21.Layer
			if unmarshalErr == nil {
				require.Len(t, layers, 1)
				require.Contains(t, layers, mvt21.LayerName(name))
				layer = layers[mvt21.LayerName(name)]
			}

			for _, ch := range tt.Checks {
				ch(t, layer, unmarshalErr)
			}
		})
	}
}

func newStringValue(value string) *spec.Tile_Value {
	return &spec.Tile_Value{
		StringValue: &value,
	}
}

func newFloatValue(value float32) *spec.Tile_Value {
	return &spec.Tile_Value{
		FloatValue: &value,
	}
}

func newDoubleValue(value float64) *spec.Tile_Value {
	return &spec.Tile_Value{
		DoubleValue: &value,
	}
}

func newIntValue(value int64) *spec.Tile_Value {
	return &spec.Tile_Value{
		IntValue: &value,
	}
}

func newUintValue(value uint64) *spec.Tile_Value {
	return &spec.Tile_Value{
		UintValue: &value,
	}
}

func newSintValue(value int64) *spec.Tile_Value {
	return &spec.Tile_Value{
		SintValue: &value,
	}
}

func newBoolValue(value bool) *spec.Tile_Value {
	return &spec.Tile_Value{
		BoolValue: &value,
	}
}

func newLayer(name string, version, extent uint32) *spec.Tile_Layer {
	return &spec.Tile_Layer{
		Version: &version,
		Name:    &name,
		Extent:  &extent,
	}
}
