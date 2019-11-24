package geometry

import (
	"fmt"
	"math"
)

// ParameterInteger are encoded integers that represent coordinates
// as arguments to MoveTo and LineTo commands.
type ParameterInteger uint32

// MakeParameterInteger encodes a ParameterInteger from an integer.
func MakeParameterInteger(value int32) (ParameterInteger, error) {
	if err := validateParameterInteger(value); err != nil {
		return 0, err
	}
	return ParameterInteger(zigzag(value)), nil
}

// Value returns the encoded integer.
func (i ParameterInteger) Value() int32 {
	return unzigzag(uint32(i))
}

// Validate the encoded parameter integer.
func (i ParameterInteger) Validate() error {
	return validateParameterInteger(i.Value())
}

func validateParameterInteger(value int32) error {
	if max := math.Pow(2, 31) - 1; float64(value) > max {
		return fmt.Errorf("value exceeds maximum")
	} else if min := max * -1; float64(value) < min {
		return fmt.Errorf("value exceeds minimum")
	}
	return nil
}

func zigzag(v int32) uint32 {
	return uint32((v << 1) ^ (v >> 31))
}

func unzigzag(v uint32) int32 {
	return int32(((v >> 1) & ((1 << 32) - 1)) ^ -(v & 1))
}
