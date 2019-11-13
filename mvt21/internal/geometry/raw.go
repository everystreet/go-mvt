package geometry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mercatormaps/go-geojson"
)

// RawShape is a raw encoded geometry.
type RawShape []uint32

// MarshalJSON returns the JSON encoding of s.
func (s *RawShape) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalJSON sets s to the JSON decoding of of s.
func (s *RawShape) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// Type returns the geometry type.
func (s *RawShape) Type() geojson.GeometryType {
	return "raw"
}

// MarshalRaw returns the stored encoded geometry sequence.
func MarshalRaw(v geojson.Geometry) ([]uint32, error) {
	return nil, nil // TODO
}

// UnmarshalRaw stores the encoded geometry in v without decoding it.
func UnmarshalRaw(data []uint32, v geojson.Geometry) error {
	rv, err := indirect(v)
	if err != nil {
		return err
	}
	rv.Set(reflect.ValueOf(data))
	return nil
}

func indirect(v geojson.Geometry) (*reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr && rv.IsNil() {
		return nil, fmt.Errorf("v must be a pointer")
	}

	i := reflect.Indirect(rv)
	return &i, nil
}
