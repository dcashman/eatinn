package data

import (
	"database/sql"
	"time"

	"eatinn.dcashman.net/internal/validator"
)

type IngredientEntry struct {
	Ingredient string
	Amount     string
	Optional   bool
}

type Recipe struct {
	ID                int64             `json:"id"`                           // Unique integer ID for the recipe
	CreatedAt         time.Time         `json:"-"`                            // Timestamp for when the recipe is added to our database
	Name              string            `json:"name"`                         // Name of the dish which the recipe creates
	Ingredients       []IngredientEntry `json:"ingredients,omitempty"`        // List of ingredients needed to make recipe
	RequiredEquipment []string          `json:"required_equipment,omitempty"` // Any notable equipment required to make the recipe
	Instructions      []string          `json:"instructions,omitempty"`       // Steps to make the dish.
	Notes             string            `json:"notes,omitempty"`              // Additional notes added to the movie.
	DisplayURL        string            `json:"display_url,omitempty"`        // URL of the image to display for this recipe
	SourceURL         string            `json:"source_url,omitempty"`         // Source of the recipe
	PrepTime          time.Duration     `json:"prep_time,omitempty"`          // The wall-clock time required to make the recipe.
	ActiveTime        time.Duration     `json:"active_time,omitempty"`        // The amount of time actively preparing the recipe, rather than passively waiting.
	Creator           string            `json:"creator,omitempty"`            // User who created this recipe
	Public            bool              `json:"public"`                       // Whether or not this recipe should be made globally available.
	Servings          int32             `json:"servings,omitempty"`           // Number of servings for this recipe
	Version           int32             `json:"version"`                      // The version number starts at 1 and will be incremented each time the recipe is updated
}

func ValidateRecipe(v *validator.Validator, r *Recipe) {
	// Use the Check() method to execute our validation checks. This will add the
	// provided key and error message to the errors map if the check does not evaluate
	// to true. For example, in the first line here we "check that the title is not
	// equal to the empty string". In the second, we "check that the length of the title
	// is less than or equal to 500 bytes" and so on.
	v.Check(r.Name != "", "name", "must be provided")
	v.Check(len(r.Name) <= 500, "name", "must not be more than 500 bytes long")
}

// Define a RecipeModel struct type which wraps a sql.DB connection pool.
type RecipeModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the movies table.
func (r RecipeModel) Insert(recipe *Recipe) error {
	return nil
}

// Add a placeholder method for fetching a specific record from the movies table.
func (r RecipeModel) Get(id int64) (*Recipe, error) {
	return nil, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (r RecipeModel) Update(recipe *Recipe) error {
	return nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (r RecipeModel) Delete(id int64) error {
	return nil
}
