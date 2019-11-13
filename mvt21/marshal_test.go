package mvt21_test

import (
	"testing"

	"github.com/everystreet/go-mvt/mvt21"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/everystreet/go-geojson"
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
	})
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

func TestMarshalMetadata(t *testing.T) {
	type check func(*testing.T, spec.Tile_Layer, error)

	var checks = func(cs ...check) []check { return cs }

	var (
		hasNoError = func() check {
			return func(t *testing.T, _ spec.Tile_Layer, err error) {
				require.NoError(t, err)
			}
		}

		hasError = func(contains string) check {
			return func(t *testing.T, _ spec.Tile_Layer, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), contains)
			}
		}

		hasMetadataLength = func(len int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Len(t, layer.Keys, 1)
				require.Len(t, layer.Values, 1)
			}
		}

		hasKeyStringValue = func(key, value string, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetStringValue())
			}
		}

		hasKeyFloat32Value = func(key string, value float32, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetFloatValue())
			}
		}

		hasKeyFloat64Value = func(key string, value float64, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetDoubleValue())
			}
		}

		hasKeyInt64Value = func(key string, value int64, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetIntValue())
			}
		}

		hasKeyUint64Value = func(key string, value uint64, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetUintValue())
			}
		}

		hasKeyBoolValue = func(key string, value bool, pos int) check {
			return func(t *testing.T, layer spec.Tile_Layer, _ error) {
				require.Equal(t, key, layer.Keys[0])
				require.Equal(t, value, layer.Values[0].GetBoolValue())
			}
		}
	)

	opts := func(os ...mvt21.MarshalOption) []mvt21.MarshalOption { return os }

	for _, tt := range []struct {
		Name     string
		Metadata geojson.PropertyList
		Options  []mvt21.MarshalOption
		Checks   []check
	}{
		{
			Name: "string value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: "value",
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyStringValue("key", "value", 0)),
		},
		{
			Name: "float32 value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: float32(3.142),
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyFloat32Value("key", 3.142, 0)),
		},
		{
			Name: "float64 value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: float64(3.142),
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyFloat64Value("key", 3.142, 0)),
		},
		{
			Name: "int value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: int(-95),
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyInt64Value("key", -95, 0)),
		},
		{
			Name: "uint value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: uint(95),
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyUint64Value("key", 95, 0)),
		},
		{
			Name: "bool value in layer",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: true,
			}},
			Checks: checks(hasNoError(), hasMetadataLength(1), hasKeyBoolValue("key", true, 0)),
		},
		{
			Name:    "additional string value",
			Options: opts(mvt21.WithStringValue("key", "value")),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyStringValue("key", "value", 0)),
		},
		{
			Name:    "additional float32 value",
			Options: opts(mvt21.WithFloat32Value("key", 3.14)),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyFloat32Value("key", 3.14, 0)),
		},
		{
			Name:    "additional float64 value",
			Options: opts(mvt21.WithFloat64Value("key", 3.14)),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyFloat64Value("key", 3.14, 0)),
		},
		{
			Name:    "additional int value",
			Options: opts(mvt21.WithIntValue("key", -95)),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyInt64Value("key", -95, 0)),
		},
		{
			Name:    "additional uint value",
			Options: opts(mvt21.WithUintValue("key", 95)),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyUint64Value("key", 95, 0)),
		},
		{
			Name:    "additional bool value",
			Options: opts(mvt21.WithBoolValue("key", true)),
			Checks:  checks(hasNoError(), hasMetadataLength(1), hasKeyBoolValue("key", true, 0)),
		},
		{
			Name: "key clash",
			Metadata: geojson.PropertyList{geojson.Property{
				Name:  "key",
				Value: int(-95),
			}},
			Options: opts(mvt21.WithUintValue("key", 95)),
			Checks:  checks(hasError("already exists")),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			data, marshalErr := mvt21.Marshal(mvt21.Layers{
				"my_layer": {
					Metadata: tt.Metadata,
				},
			}, tt.Options...)

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

func TestMarshalFeatureTags(t *testing.T) {
	t.Run("valid tags", func(t *testing.T) {
		data, err := mvt21.Marshal(mvt21.Layers{
			"my_layer": {
				Metadata: geojson.PropertyList{
					{
						Name:  "key1",
						Value: "value",
					},
					{
						Name:  "key2",
						Value: "value",
					},
					{
						Name:  "key3",
						Value: "value",
					},
				},
				Features: []mvt21.Feature{
					{
						Tags: []string{"key1", "key3"},
					},
				},
			},
		})
		require.NoError(t, err)

		var tile spec.Tile
		err = proto.Unmarshal(data, &tile)
		require.NoError(t, err)
		require.Len(t, tile.Layers, 1)

		require.Len(t, tile.Layers[0].Features, 1)
		require.Len(t, tile.Layers[0].Features[0].Tags, 2)

		var key1Pos, key3Pos int
		for i, key := range tile.Layers[0].Keys {
			switch key {
			case "key1":
				key1Pos = i
			case "key3":
				key3Pos = i
			}
		}

		require.Contains(t, tile.Layers[0].Features[0].Tags, uint32(key1Pos))
		require.Contains(t, tile.Layers[0].Features[0].Tags, uint32(key3Pos))
	})

	t.Run("invalid tag", func(t *testing.T) {
		_, err := mvt21.Marshal(mvt21.Layers{
			"my_layer": {
				Metadata: geojson.PropertyList{
					{
						Name:  "key1",
						Value: "value",
					},
				},
				Features: []mvt21.Feature{
					{
						Tags: []string{"key2"},
					},
				},
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not contain tag key")
	})
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
		})
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
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
}
