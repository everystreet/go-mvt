package mvt21

import (
	"fmt"

	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/golang/protobuf/proto"
	"github.com/mercatormaps/go-geojson"
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
		layer := newLayer(string(name), data.Extent)

		// marshal untyped metadata from layer
		kvs, err := keyValues(data.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}

		// then merge typed metadata from options
		for _, metadata := range conf.metadata {
			if _, ok := kvs[metadata.key]; ok {
				return nil, fmt.Errorf("metadata with name '%s' already exists", metadata.key)
			}
			kvs[metadata.key] = metadata.value
		}

		setKeyValues(kvs, layer)

		tile.Layers[i] = layer
		i++
	}

	return proto.Marshal(&tile)
}

func newLayer(name string, extent uint32) *spec.Tile_Layer {
	var version uint32 = 2
	return &spec.Tile_Layer{
		Version: &version,
		Name:    &name,
		Extent:  &extent,
	}
}

func setKeyValues(kvs map[string]*spec.Tile_Value, layer *spec.Tile_Layer) {
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
