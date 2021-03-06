package mvt

import (
	"fmt"

	"github.com/everystreet/go-geojson/v2"
	"github.com/everystreet/go-mvt/internal/geometry"
	spec "github.com/everystreet/go-mvt/internal/spec"
	"github.com/golang/protobuf/proto"
)

// Unproject a projected coordinate to a geographic CRS.
type Unproject geometry.Unproject

// Unmarshal parses the supplied mvt data and returns a set of layers.
func Unmarshal(data []byte, unproject Unproject) (Layers, error) {
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

		layer, err := unmarshalLayer(*data, geometry.Unproject(unproject))
		if err != nil {
			return nil, err
		}
		layers[name] = *layer
	}

	return layers, nil
}

func unmarshalLayer(data spec.Tile_Layer, unproject geometry.Unproject) (*Layer, error) {
	if v := data.GetVersion(); v != 2 {
		return nil, fmt.Errorf("unsupported version '%d'", v)
	}

	layer := Layer{
		Extent: data.GetExtent(),
	}

	if err := unmarshalFeatures(data, unproject, &layer); err != nil {
		return nil, err
	}
	return &layer, nil
}

func unmarshalFeatures(layerData spec.Tile_Layer, unproject geometry.Unproject, layer *Layer) error {
	layer.Features = make([]Feature, len(layerData.Features))

	ids := make(map[uint64]struct{})
	for i, data := range layerData.Features {
		feature := Feature{}

		if id := data.Id; id != nil {
			if _, ok := ids[*id]; ok {
				return fmt.Errorf("layer with ID '%d' already exists", id)
			}
			feature.ID = NewOptionalUint64(*id)
			ids[*id] = struct{}{}
		}

		if err := unmarshalTags(*data, layerData, &feature); err != nil {
			return err
		}

		if err := unmarshalGeometry(*data, unproject, &feature); err != nil {
			return err
		}
		layer.Features[i] = feature
	}
	return nil
}

func unmarshalTags(data spec.Tile_Feature, layer spec.Tile_Layer, feature *Feature) error {
	if len(data.Tags)%2 != 0 {
		return fmt.Errorf("expecting even number of tags")
	}

	props := make([]geojson.Property, len(data.Tags)/2)
	for i := 0; i < len(props); i++ {
		key := int(data.Tags[i*2])
		value := int(data.Tags[i*2+1])

		if key >= len(layer.Keys) {
			return fmt.Errorf("tag key '%d' does not exist in layer", key)
		} else if value >= len(layer.Values) {
			return fmt.Errorf("tag value '%d' does not exist in layer", value)
		}

		v, err := unmarshalValue(*layer.Values[value])
		if err != nil {
			return fmt.Errorf("failed to unmarshal value '%d': %w", value, err)
		}

		props[i] = geojson.Property{
			Name:  layer.Keys[key],
			Value: v,
		}
	}

	feature.Tags = props
	return nil
}

func unmarshalValue(v spec.Tile_Value) (interface{}, error) {
	switch {
	case v.StringValue != nil:
		return v.GetStringValue(), nil
	case v.FloatValue != nil:
		return v.GetFloatValue(), nil
	case v.DoubleValue != nil:
		return v.GetDoubleValue(), nil
	case v.IntValue != nil:
		return v.GetIntValue(), nil
	case v.UintValue != nil:
		return v.GetUintValue(), nil
	case v.SintValue != nil:
		return v.GetSintValue(), nil
	case v.BoolValue != nil:
		return v.GetBoolValue(), nil
	default:
		return nil, fmt.Errorf("missing value")
	}
}

func unmarshalGeometry(data spec.Tile_Feature, unproject geometry.Unproject, feature *Feature) error {
	if data.Type == nil {
		return fmt.Errorf("missing geometry type")
	}
	return geometry.Unmarshal(data.Geometry, *data.Type, unproject, &feature.Geometry)
}
