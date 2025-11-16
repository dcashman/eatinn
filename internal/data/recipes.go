package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"eatinn.dcashman.net/internal/validator"
)

type IngredientEntry struct {
	ID         int64  `json:"id"`
	Ingredient string `json:"ingredient"`
	Amount     string `json:"amount"`
	Unit       string `json:"unit"`
	Optional   bool   `json:"optional"`
}

type InstructionStep struct {
	ID         int64    `json:"id"`
	StepNumber int64    `json:"step_number"`
	Text       string   `json:"text"`
	Notes      string   `json:"notes,omitempty"`
	ImageURLs  []string `json:"image_urls,omitempty"`
}

type Recipe struct {
	ID                int64             `json:"id"`                           // Unique integer ID for the recipe
	CreatedAt         time.Time         `json:"-"`                            // Timestamp for when the recipe is added to our database
	Name              string            `json:"name"`                         // Name of the dish which the recipe creates
	Description       string            `json:"description,omitempty"`        // Description of the dish which the recipe creates
	Ingredients       []IngredientEntry `json:"ingredients,omitempty"`        // List of ingredients needed to make recipe
	RequiredEquipment []string          `json:"required_equipment,omitempty"` // Any notable equipment required to make the recipe
	Instructions      []InstructionStep `json:"instructions,omitempty"`       // Steps to make the dish.
	Notes             string            `json:"notes,omitempty"`              // Additional notes added to the recipe, not attached to any step.
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

func nilIfZero[T comparable](v T) any {
	var zero T
	if v == zero {
		return nil
	}
	return v
}

func (r RecipeModel) Insert(recipe *Recipe) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instructionsJSON, err := json.Marshal(recipe.Instructions)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO recipes
		(name, description, instructions, notes, source_url, prep_time, active_time, servings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, version`

	args := []any{recipe.Name, recipe.Description, instructionsJSON, recipe.Notes, recipe.SourceURL, nilIfZero(recipe.PrepTime), nilIfZero(recipe.ActiveTime), nilIfZero(recipe.Servings)}
	err = tx.QueryRow(
		query,
		args...,
	).Scan(&recipe.ID, &recipe.CreatedAt, &recipe.Version)

	if err != nil {
		return err
	}

	for _, entry := range recipe.Ingredients {
		err := tx.QueryRow(`
			INSERT INTO ingredients (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, entry.Ingredient).Scan(&entry.ID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit, optional)
			VALUES ($1, $2, $3, $4, $5)
		`, recipe.ID, entry.ID, entry.Amount, entry.Unit, entry.Optional)
		if err != nil {
			return err
		}
	}

	for _, equip := range recipe.RequiredEquipment {
		var equipmentID int64
		err := tx.QueryRow(`
			INSERT INTO equipment (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, equip).Scan(&equipmentID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO recipe_equipment (recipe_id, equipment_id)
			VALUES ($1, $2)
		`, recipe.ID, equipmentID)
		if err != nil {
			return err
		}
	}

	for _, step := range recipe.Instructions {
		query := `
			INSERT INTO recipe_instructions (recipe_id, step_number, instruction, notes)
			VALUES ($1, $2, $3, $4)
			RETURNING id`
		args := []any{recipe.ID, step.StepNumber, step.Text, step.Notes}
		err := tx.QueryRow(query, args...).Scan(&step.ID)
		if err != nil {
			return err
		}

		for _, url := range step.ImageURLs {
			var imageID int64
			err := tx.QueryRow(`
				INSERT INTO recipe_images (recipe_id, image_url, image_type)
				VALUES ($1, $2, 'step')
				RETURNING id
			`, recipe.ID, url).Scan(&imageID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(`
				INSERT INTO recipe_instruction_images (instruction_id, image_id)
				VALUES ($1, $2)
			`, step.ID, imageID)
			if err != nil {
				return err
			}
		}
	}

	if recipe.DisplayURL != "" {
		_, err := tx.Exec(`
			INSERT INTO recipe_images (recipe_id, image_url, image_type)
			VALUES ($1, $2, 'main')
		`, recipe.ID, recipe.DisplayURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Get fetches a specific recipe by ID along with all related data (ingredients,
// equipment, instructions, and images).
func (r RecipeModel) Get(id int64) (*Recipe, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Query main recipe data
	query := `
		SELECT id, created_at, name, description, notes, source_url,
		       prep_time, active_time, servings, version
		FROM recipes
		WHERE id = $1`

	var recipe Recipe
	var description, notes, sourceURL sql.NullString
	var prepTime, activeTime sql.NullInt64
	var servings sql.NullInt32

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&recipe.ID,
		&recipe.CreatedAt,
		&recipe.Name,
		&description,
		&notes,
		&sourceURL,
		&prepTime,
		&activeTime,
		&servings,
		&recipe.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Handle NULL values
	if description.Valid {
		recipe.Description = description.String
	}
	if notes.Valid {
		recipe.Notes = notes.String
	}
	if sourceURL.Valid {
		recipe.SourceURL = sourceURL.String
	}
	if prepTime.Valid {
		recipe.PrepTime = time.Duration(prepTime.Int64)
	}
	if activeTime.Valid {
		recipe.ActiveTime = time.Duration(activeTime.Int64)
	}
	if servings.Valid {
		recipe.Servings = servings.Int32
	}

	// Fetch ingredients
	ingredientsQuery := `
		SELECT i.id, i.name, ri.quantity, ri.unit, ri.optional
		FROM ingredients i
		INNER JOIN recipe_ingredients ri ON i.id = ri.ingredient_id
		WHERE ri.recipe_id = $1
		ORDER BY i.name`

	ingredientRows, err := r.DB.QueryContext(ctx, ingredientsQuery, id)
	if err != nil {
		return nil, err
	}
	defer ingredientRows.Close()

	recipe.Ingredients = []IngredientEntry{}
	for ingredientRows.Next() {
		var ingredient IngredientEntry
		err := ingredientRows.Scan(
			&ingredient.ID,
			&ingredient.Ingredient,
			&ingredient.Amount,
			&ingredient.Unit,
			&ingredient.Optional,
		)
		if err != nil {
			return nil, err
		}
		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	if err = ingredientRows.Err(); err != nil {
		return nil, err
	}

	// Fetch equipment
	equipmentQuery := `
		SELECT e.name
		FROM equipment e
		INNER JOIN recipe_equipment re ON e.id = re.equipment_id
		WHERE re.recipe_id = $1
		ORDER BY e.name`

	equipmentRows, err := r.DB.QueryContext(ctx, equipmentQuery, id)
	if err != nil {
		return nil, err
	}
	defer equipmentRows.Close()

	recipe.RequiredEquipment = []string{}
	for equipmentRows.Next() {
		var equipmentName string
		err := equipmentRows.Scan(&equipmentName)
		if err != nil {
			return nil, err
		}
		recipe.RequiredEquipment = append(recipe.RequiredEquipment, equipmentName)
	}

	if err = equipmentRows.Err(); err != nil {
		return nil, err
	}

	// Fetch instructions
	instructionsQuery := `
		SELECT id, step_number, instruction, notes
		FROM recipe_instructions
		WHERE recipe_id = $1
		ORDER BY step_number`

	instructionRows, err := r.DB.QueryContext(ctx, instructionsQuery, id)
	if err != nil {
		return nil, err
	}
	defer instructionRows.Close()

	recipe.Instructions = []InstructionStep{}
	for instructionRows.Next() {
		var step InstructionStep
		var notes sql.NullString
		err := instructionRows.Scan(
			&step.ID,
			&step.StepNumber,
			&step.Text,
			&notes,
		)
		if err != nil {
			return nil, err
		}
		if notes.Valid {
			step.Notes = notes.String
		}

		// Fetch images for this instruction step
		imageQuery := `
			SELECT ri.image_url
			FROM recipe_images ri
			INNER JOIN recipe_instruction_images rii ON ri.id = rii.image_id
			WHERE rii.instruction_id = $1
			ORDER BY ri.id`

		imageRows, err := r.DB.QueryContext(ctx, imageQuery, step.ID)
		if err != nil {
			return nil, err
		}

		step.ImageURLs = []string{}
		for imageRows.Next() {
			var imageURL string
			err := imageRows.Scan(&imageURL)
			if err != nil {
				imageRows.Close()
				return nil, err
			}
			step.ImageURLs = append(step.ImageURLs, imageURL)
		}
		imageRows.Close()

		if err = imageRows.Err(); err != nil {
			return nil, err
		}

		recipe.Instructions = append(recipe.Instructions, step)
	}

	if err = instructionRows.Err(); err != nil {
		return nil, err
	}

	// Fetch display image (main image)
	displayImageQuery := `
		SELECT image_url
		FROM recipe_images
		WHERE recipe_id = $1 AND image_type = 'main'
		LIMIT 1`

	var displayURL sql.NullString
	err = r.DB.QueryRowContext(ctx, displayImageQuery, id).Scan(&displayURL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if displayURL.Valid {
		recipe.DisplayURL = displayURL.String
	}

	return &recipe, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (r RecipeModel) Update(recipe *Recipe) error {
	return nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (r RecipeModel) Delete(id int64) error {
	return nil
}
