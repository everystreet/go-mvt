package mvt21

import (
	"strconv"

	"github.com/everystreet/go-geojson/v2"
	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
)

type (
	// Layer is a single layer in a tile. A layer consists of zero or more featues.
	Layer struct {
		Extent   uint32
		Features []Feature
	}

	// Layers is an ordered list of named layers in a tile.
	// Layer names must be unique inside a single tile.
	Layers map[LayerName]Layer

	// LayerName is a string.
	LayerName string
)

// MakeLayer setting the required extent field.
func MakeLayer(extent uint32, features ...Feature) Layer {
	return Layer{
		Extent:   extent,
		Features: features,
	}
}

// Feature represents a geographical feature and optional attributes.
// Tags is a list of properties.
type Feature struct {
	Geometry geojson.Geometry
	ID       OptionalUint64
	Tags     geojson.PropertyList
}

// NewFeature makes a new feature, setting the required geometry field.
func NewFeature(geo geojson.Geometry) *Feature {
	return &Feature{Geometry: geo}
}

// OptionalUint64 is a type that represents a Uint64 that can be optionally set.
type OptionalUint64 struct {
	value *uint64
}

// NewOptionalUint64 creates a new OptionalUint64 set to the specified value.
func NewOptionalUint64(val uint64) OptionalUint64 {
	return OptionalUint64{value: &val}
}

// Value returns the value. Should call this method if OptionalUint64.IsSet() returns true.
func (o OptionalUint64) Value() uint64 {
	return *o.value
}

// IsSet returns true if the value is set, and false if not.
func (o OptionalUint64) IsSet() bool {
	return o.value != nil
}

// Get the Uint64 value and whether or not it's set.
func (o OptionalUint64) Get() (uint64, bool) {
	if o.value == nil {
		return 0, false
	}
	return *o.value, true
}

func (o OptionalUint64) String() string {
	if o.IsSet() {
		return strconv.FormatUint(*o.value, 10)
	}
	return "{unset}"
}

// UnknownGeometry implements to geojson.Geometry interface.
type UnknownGeometry struct {
	geometry.RawShape
}
