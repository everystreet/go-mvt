package mvt21_test

import (
	"testing"

	"github.com/everystreet/go-geojson"
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

func TestUnmarshalFeatureTags(t *testing.T) {
	type check func(*testing.T, mvt21.Feature, error)

	var checks = func(cs ...check) []check { return cs }

	var (
		hasNoError = func() check {
			return func(t *testing.T, _ mvt21.Feature, err error) {
				require.NoError(t, err)
			}
		}

		hasTagsLength = func(num int) check {
			return func(t *testing.T, feature mvt21.Feature, _ error) {
				require.Len(t, feature.Tags, num)
			}
		}

		hasTag = func(name string, value interface{}) check {
			return func(t *testing.T, feature mvt21.Feature, _ error) {
				require.Contains(t, feature.Tags, geojson.Property{
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
		Tags   []uint32
		Checks []check
	}{
		{
			Name:   "string value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newStringValue("value")},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", "value")),
		},
		{
			Name:   "float value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newFloatValue(3.142)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", float32(3.142))),
		},
		{
			Name:   "double value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newDoubleValue(3.142)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", float64(3.142))),
		},
		{
			Name:   "int value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newIntValue(-95)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", int64(-95))),
		},
		{
			Name:   "uint value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newUintValue(95)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", uint64(95))),
		},
		{
			Name:   "sint value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newSintValue(-95)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", int64(-95))),
		},
		{
			Name:   "bool value",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newBoolValue(true)},
			Tags:   []uint32{0, 0},
			Checks: checks(hasNoError(), hasTagsLength(1), hasTag("key", true)),
		},
		{
			Name:   "multiple values",
			Keys:   []string{"key"},
			Values: []*spec.Tile_Value{newStringValue("value"), newIntValue(-95)},
			Tags:   []uint32{0, 0, 0, 1},
			Checks: checks(hasNoError(), hasTagsLength(2), hasTag("key", "value"), hasTag("key", int64(-95))),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			name, version, typ := "layer1", uint32(2), spec.Tile_UNKNOWN
			data, err := proto.Marshal(&spec.Tile{
				Layers: []*spec.Tile_Layer{
					{
						Version: &version,
						Name:    &name,
						Keys:    tt.Keys,
						Values:  tt.Values,
						Features: []*spec.Tile_Feature{
							{
								Type: &typ,
								Tags: tt.Tags,
							},
						},
					},
				},
			})
			require.NoError(t, err)

			layers, unmarshalErr := mvt21.Unmarshal(data)

			var feature mvt21.Feature
			if unmarshalErr == nil {
				require.Len(t, layers, 1)
				require.Contains(t, layers, mvt21.LayerName(name))

				require.Len(t, layers[mvt21.LayerName(name)].Features, 1)
				feature = layers[mvt21.LayerName(name)].Features[0]
			}

			for _, ch := range tt.Checks {
				ch(t, feature, unmarshalErr)
			}
		})
	}
}

func TestUnmarshalFeatureID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		name, version, id, typ := "my_layer", uint32(2), uint64(67), spec.Tile_UNKNOWN
		data, err := proto.Marshal(&spec.Tile{
			Layers: []*spec.Tile_Layer{
				{
					Version: &version,
					Name:    &name,
					Features: []*spec.Tile_Feature{
						{
							Type: &typ,
							Id:   &id,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		layers, err := mvt21.Unmarshal(data)
		require.NoError(t, err)
		require.Len(t, layers, 1)

		require.Contains(t, layers, mvt21.LayerName("my_layer"))
		require.Len(t, layers["my_layer"].Features, 1)

		require.True(t, layers["my_layer"].Features[0].ID.IsSet())
		require.Equal(t, 67, int(layers["my_layer"].Features[0].ID.Value()))
	})

	t.Run("duplicate ID", func(t *testing.T) {
		name, version, id, typ := "my_layer", uint32(2), uint64(67), spec.Tile_UNKNOWN
		data, err := proto.Marshal(&spec.Tile{
			Layers: []*spec.Tile_Layer{
				{
					Version: &version,
					Name:    &name,
					Features: []*spec.Tile_Feature{
						{
							Type: &typ,
							Id:   &id,
						},
						{
							Type: &typ,
							Id:   &id,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		_, err = mvt21.Unmarshal(data)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
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
