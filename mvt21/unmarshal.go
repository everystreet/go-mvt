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

		layer, err := unmarshal(*data)
		if err != nil {
			return nil, err
		}
		layers[name] = *layer
	}

	return layers, nil
}

func unmarshal(data spec.Tile_Layer) (*Layer, error) {
	if v := data.GetVersion(); v != 2 {
		return nil, fmt.Errorf("unsupported version '%d'", v)
	} else if ks, vs := len(data.Keys), len(data.Values); ks != vs {
		return nil, fmt.Errorf("number of keys and values unequal (%d != %d)", ks, vs)
	}

	metadata := make([]geojson.Property, len(data.Keys))
	keys := make(map[string]interface{}, len(data.Keys))

	for i, key := range data.Keys {
		if _, ok := keys[key]; ok {
			return nil, fmt.Errorf("key with name '%s' already exists", key)
		}
		keys[key] = nil

		switch {
		case data.Values[i].StringValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetStringValue()}
		case data.Values[i].FloatValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetFloatValue()}
		case data.Values[i].DoubleValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetDoubleValue()}
		case data.Values[i].IntValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetIntValue()}
		case data.Values[i].UintValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetUintValue()}
		case data.Values[i].SintValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetSintValue()}
		case data.Values[i].BoolValue != nil:
			metadata[i] = geojson.Property{Name: key, Value: data.Values[i].GetBoolValue()}
		default:
			return nil, fmt.Errorf("missing value for '%s'", key)
		}
	}

	return &Layer{
		Extent:   data.GetExtent(),
		Metadata: metadata,
	}, nil
}
