package mvt_test

import (
	"testing"

	"github.com/everystreet/go-geojson/v2"
	mvt21 "github.com/everystreet/go-mvt"
	spec "github.com/everystreet/go-mvt/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestMarshalLayers(t *testing.T) {
	data, err := mvt21.Marshal(mvt21.Layers{
		"layer1": {
			Extent: 4096,
		},
		"layer2": {
			Extent: 2048,
		},
	}, nil)
	require.NoError(t, err)

	var tile spec.Tile
	err = proto.Unmarshal(data, &tile)
	require.NoError(t, err)
	require.Len(t, tile.Layers, 2)

	var layer1, layer2 *spec.Tile_Layer
	for _, l := range tile.Layers {
		switch l.GetName() {
		case "layer1":
			layer1 = l
		case "layer2":
			layer2 = l
		}
	}

	require.Equal(t, uint32(4096), layer1.GetExtent())
	require.Equal(t, uint32(2048), layer2.GetExtent())
}

func TestMarshalFeatureID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		data, err := mvt21.Marshal(mvt21.Layers{
			"my_layer": {
				Features: []mvt21.Feature{
					{
						ID: mvt21.NewOptionalUint64(67),
					},
				},
			},
		}, nil)
		require.NoError(t, err)

		var tile spec.Tile
		err = proto.Unmarshal(data, &tile)
		require.NoError(t, err)
		require.Len(t, tile.Layers, 1)

		require.Len(t, tile.Layers[0].Features, 1)
		require.Equal(t, 67, int(tile.Layers[0].Features[0].GetId()))
	})

	t.Run("duplicate ID", func(t *testing.T) {
		_, err := mvt21.Marshal(mvt21.Layers{
			"my_layer": {
				Features: []mvt21.Feature{
					{
						ID: mvt21.NewOptionalUint64(67),
					},
					{
						ID: mvt21.NewOptionalUint64(67),
					},
				},
			},
		}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
}

func TestMarshalFeatureTags(t *testing.T) {
	type check func(*testing.T, spec.Tile_Layer, error)

	var checks = func(cs ...check) []check { return cs }

	var (
		hasNoError = func() check {
			return func(t *testing.T, _ spec.Tile_Layer, err error) {
				require.NoError(t, err)
			}
		}

		hasKeysLength = func(n int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Len(t, layer.Keys, n)
			}
		}

		hasValuesLength = func(n int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Len(t, layer.Values, n)
			}
		}

		hasKey = func(key string) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Contains(t, layer.Keys, key)
			}
		}

		hasKeyStringValue = func(value string) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.StringValue != nil {
						require.Equal(t, value, val.GetStringValue())
						return
					}
				}
				t.FailNow()
			}
		}

		hasKeyFloat32Value = func(value float32) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.FloatValue != nil {
						require.Equal(t, value, val.GetFloatValue())
						return
					}
				}
				t.FailNow()
			}
		}

		hasKeyFloat64Value = func(value float64) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.DoubleValue != nil {
						require.Equal(t, value, val.GetDoubleValue())
						return
					}
				}
				t.FailNow()
			}
		}

		hasKeyInt64Value = func(value int64) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.IntValue != nil {
						require.Equal(t, value, val.GetIntValue())
						return
					}
				}
				t.FailNow()
			}
		}

		hasKeyUint64Value = func(value uint64) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.UintValue != nil {
						require.Equal(t, value, val.GetUintValue())
						return
					}
				}
				t.FailNow()
			}
		}

		hasKeyBoolValue = func(value bool) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				for _, val := range layer.Values {
					if val.BoolValue != nil {
						require.Equal(t, value, val.GetBoolValue())
						return
					}
				}
				t.FailNow()
			}
		}

		tagsAre = func(tags ...uint32) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, tags, layer.Features[0].Tags)
			}
		}
	)

	for _, tt := range []struct {
		Name   string
		Tags   []geojson.Property
		Checks []check
	}{
		{
			Name: "string value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: "value",
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyStringValue("value"), tagsAre(0, 0)),
		},
		{
			Name: "float32 value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: float32(3.142),
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyFloat32Value(3.142), tagsAre(0, 0)),
		},
		{
			Name: "float64 value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: float64(3.142),
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyFloat64Value(3.142), tagsAre(0, 0)),
		},
		{
			Name: "int value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: int(-95),
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyInt64Value(-95), tagsAre(0, 0)),
		},
		{
			Name: "uint value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: uint(95),
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyUint64Value(95), tagsAre(0, 0)),
		},
		{
			Name: "bool value",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: true,
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(1), hasKeyBoolValue(true), tagsAre(0, 0)),
		},
		{
			Name: "multiple values",
			Tags: []geojson.Property{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key",
					Value: int(-95),
				},
			},
			Checks: checks(hasNoError(), hasKeysLength(1), hasKey("key"),
				hasValuesLength(2), hasKeyStringValue("value"), hasKeyInt64Value(-95), tagsAre(0, 0, 0, 1)),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			data, marshalErr := mvt21.Marshal(mvt21.Layers{
				"my_layer": {
					Features: []mvt21.Feature{
						{
							Tags: tt.Tags,
						},
					},
				},
			}, nil)

			var layer spec.Tile_Layer
			if marshalErr == nil {
				var tile spec.Tile
				err := proto.Unmarshal(data, &tile)
				require.NoError(t, err)

				require.Len(t, tile.Layers, 1)
				layer = *tile.Layers[0]
			}

			for _, ch := range tt.Checks {
				ch(t, layer, marshalErr)
			}
		})
	}
}
