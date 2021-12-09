package models

import (
	"database/sql/driver"
	"errors"
)

// JSONB -
type JSONB []byte

// Value -
func (j JSONB) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

// Scan -
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("scan source was not []byte")
	}
	*j = append((*j)[0:0], s...)

	return nil
}

// MarshalJSON returns *m as the JSON encoding of m.
func (m JSONB) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *JSONB) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// IsNull -
func (j JSONB) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}
