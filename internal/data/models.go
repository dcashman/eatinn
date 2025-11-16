package data

import (
	"database/sql"
	"errors"
)

// Define custom errors for our data models.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Create a Models struct which wraps the RecipeModel. We'll add other models to this,
// like a UserModel and PermissionModel, as our build progresses.
type Models struct {
	Recipes RecipeModel
}

// For ease of use, we also add a New() method which returns a Models struct containing
// the initialized RecipeModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Recipes: RecipeModel{DB: db},
	}
}
