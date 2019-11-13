package mvt21

import (
	"fmt"
	"strings"

	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/everystreet/go-geojson"
)

// Marshal returns the mvt encoding of the supplied layers.
func Marshal(layers Layers, opts ...MarshalOption) ([]byte, error) {
	tile := spec.Tile{
		Layers: make([]*spec.Tile_Layer, len(layers)),
	}

	conf := marshalConfig{}
	for _, opt := range opts {
		opt(&conf)
	}

	var i int
	for name, data := range layers {
		layer, err := marshalLayer(data, string(name), conf)
		if err != nil {
			return nil, err
		}

		tile.Layers[i] = layer
		i++
	}

	return proto.Marshal(&tile)
}

// MarshalOption funcs can be used to modify the behaviour of Marshal.
type MarshalOption func(*marshalConfig)

// WithStringValue adds a typed string value to all layers.
func WithStringValue(key, value string) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				StringValue: &value,
			},
		})
	}
}

// WithFloat32Value adds a typed float32 value to all layers.
func WithFloat32Value(key string, value float32) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				FloatValue: &value,
			},
		})
	}
}

// WithFloat64Value adds a typed float64 value to all layers.
func WithFloat64Value(key string, value float64) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				DoubleValue: &value,
			},
		})
	}
}

// WithIntValue adds a typed int value to all layers.
func WithIntValue(key string, value int64) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				IntValue: &value,
			},
		})
	}
}

// WithUintValue adds a typed uint value to all layers.
func WithUintValue(key string, value uint64) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				UintValue: &value,
			},
		})
	}
}

// WithBoolValue adds a typed boolean value to all layers.
func WithBoolValue(key string, value bool) MarshalOption {
	return func(conf *marshalConfig) {
		conf.metadata = append(conf.metadata, metadata{
			key: key,
			value: &spec.Tile_Value{
				BoolValue: &value,
			},
		})
	}
}

type marshalConfig struct {
	metadata []metadata
}

type metadata struct {
	key   string
	value *spec.Tile_Value
}

func marshalLayer(data Layer, name string, conf marshalConfig) (*spec.Tile_Layer, error) {
	var version uint32 = 2
	layer := spec.Tile_Layer{
		Version: &version,
		Name:    &name,
		Extent:  &data.Extent,
	}

	if err := marshalKeyValues(data.Metadata, conf.metadata, &layer); err != nil {
		return nil, err
	}

	if err := marshalFeatures(data.Features, &layer); err != nil {
		return nil, err
	}
	return &layer, nil
}

func marshalKeyValues(metadata geojson.PropertyList, additional []metadata, layer *spec.Tile_Layer) error {
	// marshal untyped metadata from layer
	kvs, err := keyValues(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// then merge typed metadata from options
	for _, metadata := range additional {
		if _, ok := kvs[metadata.key]; ok {
			return fmt.Errorf("metadata with name '%s' already exists", metadata.key)
		}
		kvs[metadata.key] = metadata.value
	}

	keys := make([]string, len(kvs))
	values := make([]*spec.Tile_Value, len(kvs))

	var i int
	for key, value := range kvs {
		keys[i] = key
		values[i] = value
		i++
	}

	layer.Keys = keys
	layer.Values = values
	return nil
}

func keyValues(metadata []geojson.Property) (map[string]*spec.Tile_Value, error) {
	kvs := make(map[string]*spec.Tile_Value, len(metadata))
	for _, prop := range metadata {
		if _, ok := kvs[prop.Name]; ok {
			return nil, fmt.Errorf("metadata with name '%s' already exists", prop.Name)
		}

		value := prop.Value

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
			kvs[prop.Name] = &spec.Tile_Value{
				StringValue: &v,
			}
		case float32:
			kvs[prop.Name] = &spec.Tile_Value{
				FloatValue: &v,
			}
		case float64:
			kvs[prop.Name] = &spec.Tile_Value{
				DoubleValue: &v,
			}
		case int64:
			kvs[prop.Name] = &spec.Tile_Value{
				IntValue: &v,
			}
		case uint64:
			kvs[prop.Name] = &spec.Tile_Value{
				UintValue: &v,
			}
		case bool:
			kvs[prop.Name] = &spec.Tile_Value{
				BoolValue: &v,
			}
		default:
			return nil, fmt.Errorf("metadata '%s' is of unsupported value type '%t'", prop.Name, v)
		}
	}
	return kvs, nil
}

func marshalFeatures(features []Feature, layer *spec.Tile_Layer) error {
	layer.Features = make([]*spec.Tile_Feature, len(features))

	ids := make(map[uint64]struct{})
	for i, data := range features {
		feature := &spec.Tile_Feature{}

		if id, ok := data.ID.Get(); ok {
			if _, ok = ids[id]; ok {
				return fmt.Errorf("layer with ID '%d' already exists", id)
			}
			feature.Id = &id
			ids[id] = struct{}{}
		}

		tags, err := featureTags(data.Tags, *layer)
		if err != nil {
			return err
		}
		feature.Tags = tags

		layer.Features[i] = feature
	}
	return nil
}

func featureTags(tags []string, layer spec.Tile_Layer) ([]uint32, error) {
	indexes := make([]uint32, len(tags))
	for i, tag := range tags {
		hasKey := false
		for pos, key := range layer.Keys {
			if strings.EqualFold(key, tag) {
				hasKey = true
				indexes[i] = uint32(pos)
				break
			}
		}

		if !hasKey {
			return nil, fmt.Errorf("layer does not contain tag key '%s'", tag)
		}
	}
	return indexes, nil
}
