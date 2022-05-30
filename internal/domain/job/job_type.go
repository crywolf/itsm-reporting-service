package job

import (
	"encoding/json"
	"fmt"
)

// Type of the job is enum
type Type struct {
	v string
}

// Type values
var (
	TypeAll = Type{"all"}
	TypeFE  = Type{"FE report only"}
	TypeSD  = Type{"SD report only"}
)

var jobTypeValues = []Type{
	TypeAll,
	TypeFE,
	TypeSD,
}

// NewTypeFromString creates new instance from string value
func NewTypeFromString(typeStr string) (Type, error) {
	for _, t := range jobTypeValues {
		if t.String() == typeStr {
			return t, nil
		}
	}

	return Type{}, fmt.Errorf("unknown '%s' job type", typeStr)
}

// IsZero returns true if Type has zero value.
// Every type in Go have zero value. In that case it's `Type{}`.
// It's always a good idea to check if provided value is not zero!
func (s Type) IsZero() bool {
	return s == Type{}
}

func (s Type) String() string {
	return s.v
}

// MarshalJSON returns JSON encoded Type
func (s Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON sets Type value from JSON data
func (s *Type) UnmarshalJSON(b []byte) error {
	var v string
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v == "" {
		s.v = v
		return nil
	}

	t, err := NewTypeFromString(v)
	if err != nil {
		s.v = err.Error()
		return nil
	}

	s.v = t.String()

	return nil
}
