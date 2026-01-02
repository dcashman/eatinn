package data

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Define custom errors for our data models.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Duration wraps time.Duration to provide custom JSON marshaling/unmarshaling.
// It accepts and outputs duration strings like "30m", "1h30m", "45s" in JSON.
type Duration time.Duration

// MarshalJSON implements the json.Marshaler interface.
// It outputs the duration as a string in Go's duration format (e.g., "1h30m").
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It accepts duration strings like "30m", "1h30m", "2h15m30s".
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
		*d = Duration(dur)
		return nil
	default:
		return fmt.Errorf("duration must be a string (e.g., \"30m\", \"1h30m\")")
	}
}

// Create a Models struct which wraps the RecipeModel. We'll add other models to this,
// like a UserModel and PermissionModel, as our build progresses.
type Models struct {
	Recipes RecipeModel
	Users   UserModel
}

// For ease of use, we also add a New() method which returns a Models struct containing
// the initialized RecipeModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Recipes: RecipeModel{DB: db},
		Users:   UserModel{DB: db},
	}
}
