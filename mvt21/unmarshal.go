package mvt21

import (
	"fmt"

	"github.com/everystreet/go-geojson"
	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
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

	if err := unmarshalFeatures(data.Features, &layer); err != nil {
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

func unmarshalFeatures(features []*spec.Tile_Feature, layer *Layer) error {
	layer.Features = make([]Feature, len(features))

	ids := make(map[uint64]struct{})
	for i, data := range features {
		feature := Feature{}

		if id := data.Id; id != nil {
			if _, ok := ids[*id]; ok {
				return fmt.Errorf("layer with ID '%d' already exists", id)
			}
			feature.ID = NewOptionalUint64(*id)
			ids[*id] = struct{}{}
		}

		feature.Tags = make([]string, len(data.Tags))
		for i, tag := range data.Tags {
			if int(tag) >= len(layer.Metadata) {
				return fmt.Errorf("tag key '%d' does not exist in layer", tag)
			}
			feature.Tags[i] = layer.Metadata[tag].Name
		}

		if err := unmarshalGeometry(*data, &feature); err != nil {
			return err
		}
		layer.Features[i] = feature
	}
	return nil
}

func unmarshalGeometry(data spec.Tile_Feature, feature *Feature) error {
	if data.Type == nil {
		return fmt.Errorf("missing geometry type")
	}

	switch *data.Type {
	case spec.Tile_UNKNOWN:
		geo := &UnknownGeometry{}
		feature.Geometry = geo
		return geometry.UnmarshalRaw(data.Geometry, &geo.RawShape)
	default:
		return fmt.Errorf("unknown geometry type '%v'", data.Type)
	}
}
