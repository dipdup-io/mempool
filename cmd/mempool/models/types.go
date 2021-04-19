package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// JSON -
type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.Errorf("Failed to unmarshal JSONB value: %v", value)
	}

	var result json.RawMessage
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON
func (j *JSON) UnmarshalJSON(data []byte) error {
	*j = JSON(data)
	return nil
}
