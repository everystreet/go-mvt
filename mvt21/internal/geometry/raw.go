package geometry

import (
	"encoding/json"

	"github.com/everystreet/go-geojson/v2"
)

// RawShape is a raw encoded geometry.
type RawShape []uint32

// MarshalJSON returns the JSON encoding of s.
func (s RawShape) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalJSON sets s to the JSON decoding of of s.
func (s *RawShape) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// Type returns the geometry type.
func (s RawShape) Type() geojson.GeometryType {
	return "raw"
}

// Validate the RawShape.
func (s RawShape) Validate() error {
	return nil
}
