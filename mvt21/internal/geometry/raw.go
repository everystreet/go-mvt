package geometry

import (
	"encoding/json"

	"github.com/mercatormaps/go-geojson"
)

type RawShape []uint32

func (s *RawShape) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *RawShape) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *RawShape) Type() geojson.GeometryType {
	return "raw"
}
