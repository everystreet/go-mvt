package mvt21

import (
	"fmt"

	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/mercatormaps/go-geojson"
)

// Unmarshal parses the supplied mvt data and returns a set of layers.
func Unmarshal(data []byte) (Layers, error) {
	tile := spec.Tile{}
	if err := proto.Unmarshal(data, &tile); err != nil {
		return nil, err
	}

	layers := make(Layers, len(tile.Layers))
	for _, data := range tile.Layers {
		name := LayerName(data.GetName())
		if _, ok := layers[name]; ok {
			return nil, fmt.Errorf("layer with name '%s' already exists", name)
		}

		layer, err := unmarshalLayer(*data)
		if err != nil {
			return nil, err
		}
		layers[name] = *layer
	}

	return layers, nil
}

func unmarshalLayer(data spec.Tile_Layer) (*Layer, error) {
	if v := data.GetVersion(); v != 2 {
		return nil, fmt.Errorf("unsupported version '%d'", v)
	}

	layer := Layer{
		Extent: data.GetExtent(),
	}

	if err := unmarshalKeyValues(data.Keys, data.Values, &layer); err != nil {
		return nil, err
	}
	return &layer, nil
}

func unmarshalKeyValues(keys []string, values []*spec.Tile_Value, layer *Layer) error {
	if ks, vs := len(keys), len(values); ks != vs {
		return fmt.Errorf("number of keys and values unequal (%d != %d)", ks, vs)
	}

	metadata := make([]geojson.Property, len(keys))
	uniqueKeys := make(map[string]interface{}, len(keys))

	for i, key := range keys {
		if _, ok := uniqueKeys[key]; ok {
			return fmt.Errorf("key with name '%s' already exists", key)
		}
		uniqueKeys[key] = nil

		switch {
		case values[i].StringValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetStringValue()}
		case values[i].FloatValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetFloatValue()}
		case values[i].DoubleValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetDoubleValue()}
		case values[i].IntValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetIntValue()}
		case values[i].UintValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetUintValue()}
		case values[i].SintValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetSintValue()}
		case values[i].BoolValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: values[i].GetBoolValue()}
		default:
			return fmt.Errorf("missing value for '%s'", key)
		}
	}

	layer.Metadata = metadata
	return nil
}
