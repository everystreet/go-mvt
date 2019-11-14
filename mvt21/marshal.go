package mvt21

import (
	"fmt"

	"github.com/everystreet/go-geojson"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
)

// Marshal returns the mvt encoding of the supplied layers.
func Marshal(layers Layers) ([]byte, error) {
	tile := spec.Tile{
		Layers: make([]*spec.Tile_Layer, len(layers)),
	}

	var i int
	for name, data := range layers {
		layer, err := marshalLayer(data, string(name))
		if err != nil {
			return nil, err
		}

		tile.Layers[i] = layer
		i++
	}

	return proto.Marshal(&tile)
}

func marshalLayer(data Layer, name string) (*spec.Tile_Layer, error) {
	var version uint32 = 2
	layer := spec.Tile_Layer{
		Version: &version,
		Name:    &name,
		Extent:  &data.Extent,
	}

	if err := marshalFeatures(data.Features, &layer); err != nil {
		return nil, err
	}
	return &layer, nil
}

func marshalFeatures(features []Feature, layer *spec.Tile_Layer) error {
	layer.Features = make([]*spec.Tile_Feature, len(features))

	ids := make(map[uint64]struct{})
	keys := make(map[string]int)
	values := make(map[interface{}]int)

	for i, data := range features {
		feature := spec.Tile_Feature{}

		if id, ok := data.ID.Get(); ok {
			if _, ok = ids[id]; ok {
				return fmt.Errorf("layer with ID '%d' already exists", id)
			}
			feature.Id = &id
			ids[id] = struct{}{}
		}

		marshalTags(data.Tags, keys, values, &feature)

		layer.Features[i] = &feature
	}

	return marshalKeyValues(keys, values, layer)
}

func marshalTags(tags geojson.PropertyList, keys map[string]int, values map[interface{}]int, feature *spec.Tile_Feature) {
	feature.Tags = make([]uint32, len(tags)*2)
	for i, tag := range tags {
		if _, ok := keys[tag.Name]; !ok {
			keys[tag.Name] = len(keys)
		}

		if _, ok := values[tag.Value]; !ok {
			values[tag.Value] = len(values)
		}

		feature.Tags[i*2] = uint32(keys[tag.Name])
		feature.Tags[i*2+1] = uint32(values[tag.Value])
	}
}

func marshalKeyValues(keys map[string]int, values map[interface{}]int, layer *spec.Tile_Layer) error {
	layer.Keys = make([]string, len(keys))
	for key, pos := range keys {
		layer.Keys[pos] = key
	}

	layer.Values = make([]*spec.Tile_Value, len(values))
	for value, pos := range values {
		v, err := marshalKeyValue(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		layer.Values[pos] = v
	}
	return nil
}

func marshalKeyValue(value interface{}) (*spec.Tile_Value, error) {
	switch v := value.(type) {
	case int:
		value = int64(v)
	case int8:
		value = int64(v)
	case int16:
		value = int64(v)
	case int32:
		value = int64(v)
	case uint:
		value = uint64(v)
	case uint8:
		value = uint64(v)
	case uint16:
		value = uint64(v)
	case uint32:
		value = uint64(v)
	}

	switch v := value.(type) {
	case string:
		return &spec.Tile_Value{
			StringValue: &v,
		}, nil
	case float32:
		return &spec.Tile_Value{
			FloatValue: &v,
		}, nil
	case float64:
		return &spec.Tile_Value{
			DoubleValue: &v,
		}, nil
	case int64:
		return &spec.Tile_Value{
			IntValue: &v,
		}, nil
	case uint64:
		return &spec.Tile_Value{
			UintValue: &v,
		}, nil
	case bool:
		return &spec.Tile_Value{
			BoolValue: &v,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type '%t'", v)
	}
}
