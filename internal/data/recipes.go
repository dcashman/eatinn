package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	PrepTime          Duration          `json:"prep_time,omitempty"`          // The wall-clock time required to make the recipe.
	ActiveTime        Duration          `json:"active_time,omitempty"`        // The amount of time actively preparing the recipe, rather than passively waiting.
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

// durationToInterval converts a time.Duration to a PostgreSQL interval string.
// Returns nil if the duration is zero, otherwise returns a string like "300.5 seconds".
func durationToInterval(d time.Duration) *string {
	if d == 0 {
		return nil
	}
	// PostgreSQL interval format: "X seconds" where X can be fractional
	seconds := d.Seconds()
	interval := fmt.Sprintf("%f seconds", seconds)
	return &interval
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

	// Convert data.Duration to PostgreSQL interval strings for database storage
	args := []any{recipe.Name, recipe.Description, instructionsJSON, recipe.Notes, recipe.SourceURL, durationToInterval(time.Duration(recipe.PrepTime)), durationToInterval(time.Duration(recipe.ActiveTime)), nilIfZero(recipe.Servings)}
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
		recipe.PrepTime = Duration(prepTime.Int64)
	}
	if activeTime.Valid {
		recipe.ActiveTime = Duration(activeTime.Int64)
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

// Update modifies an existing recipe in the database. It uses optimistic locking
// via the version field to prevent race conditions.
func (r RecipeModel) Update(recipe *Recipe) error {
	// Start a transaction
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the main recipe record with optimistic locking
	query := `
		UPDATE recipes
		SET name = $1, description = $2, notes = $3, source_url = $4,
		    prep_time = $5, active_time = $6, servings = $7, version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING version`

	// Convert data.Duration to PostgreSQL interval strings for database storage
	args := []any{
		recipe.Name,
		recipe.Description,
		recipe.Notes,
		recipe.SourceURL,
		durationToInterval(time.Duration(recipe.PrepTime)),
		durationToInterval(time.Duration(recipe.ActiveTime)),
		nilIfZero(recipe.Servings),
		recipe.ID,
		recipe.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, query, args...).Scan(&recipe.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	// Delete existing related data (we'll re-insert it)
	// This is simpler than trying to diff and update individual items

	// Delete existing ingredients
	_, err = tx.ExecContext(ctx, `
		DELETE FROM recipe_ingredients WHERE recipe_id = $1
	`, recipe.ID)
	if err != nil {
		return err
	}

	// Delete existing equipment
	_, err = tx.ExecContext(ctx, `
		DELETE FROM recipe_equipment WHERE recipe_id = $1
	`, recipe.ID)
	if err != nil {
		return err
	}

	// Delete existing instructions (CASCADE will handle instruction images)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM recipe_instructions WHERE recipe_id = $1
	`, recipe.ID)
	if err != nil {
		return err
	}

	// Delete existing display image
	_, err = tx.ExecContext(ctx, `
		DELETE FROM recipe_images WHERE recipe_id = $1 AND image_type = 'main'
	`, recipe.ID)
	if err != nil {
		return err
	}

	// Re-insert ingredients
	for _, entry := range recipe.Ingredients {
		err := tx.QueryRowContext(ctx, `
			INSERT INTO ingredients (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, entry.Ingredient).Scan(&entry.ID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit, optional)
			VALUES ($1, $2, $3, $4, $5)
		`, recipe.ID, entry.ID, entry.Amount, entry.Unit, entry.Optional)
		if err != nil {
			return err
		}
	}

	// Re-insert equipment
	for _, equip := range recipe.RequiredEquipment {
		var equipmentID int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO equipment (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, equip).Scan(&equipmentID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO recipe_equipment (recipe_id, equipment_id)
			VALUES ($1, $2)
		`, recipe.ID, equipmentID)
		if err != nil {
			return err
		}
	}

	// Re-insert instructions
	for _, step := range recipe.Instructions {
		query := `
			INSERT INTO recipe_instructions (recipe_id, step_number, instruction, notes)
			VALUES ($1, $2, $3, $4)
			RETURNING id`
		args := []any{recipe.ID, step.StepNumber, step.Text, step.Notes}
		err := tx.QueryRowContext(ctx, query, args...).Scan(&step.ID)
		if err != nil {
			return err
		}

		// Insert images for this instruction step
		for _, url := range step.ImageURLs {
			var imageID int64
			err := tx.QueryRowContext(ctx, `
				INSERT INTO recipe_images (recipe_id, image_url, image_type)
				VALUES ($1, $2, 'step')
				RETURNING id
			`, recipe.ID, url).Scan(&imageID)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO recipe_instruction_images (instruction_id, image_id)
				VALUES ($1, $2)
			`, step.ID, imageID)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert display image if provided
	if recipe.DisplayURL != "" {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO recipe_images (recipe_id, image_url, image_type)
			VALUES ($1, $2, 'main')
		`, recipe.ID, recipe.DisplayURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete removes a recipe from the database. The CASCADE constraints in the schema
// will automatically delete related records in junction tables.
func (r RecipeModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM recipes WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll retrieves a list of recipes with optional filtering, sorting, and pagination.
// Returns a slice of recipes and pagination metadata.
func (r RecipeModel) GetAll(name string, ingredients []string, equipment []string, prepTime Duration, activeTime Duration, filters Filters) ([]*Recipe, Metadata, error) {
	// Build the query with window function for total count
	// Use a CTE to filter recipes, then join for display images
	// Note: Go's time.Duration is int64 nanoseconds, but PostgreSQL prep_time/active_time
	// columns are interval type. We extract epoch (total seconds) from the interval and
	// compare it to the input nanoseconds converted to seconds.
	query := `
		WITH filtered_recipes AS (
			SELECT DISTINCT r.id, r.name, r.description, r.prep_time, r.active_time,
			       r.servings, r.created_at, r.version
			FROM recipes r
			WHERE ($1 = '' OR r.name ILIKE '%' || $1 || '%')
			  AND ($2::double precision = 0 OR EXTRACT(EPOCH FROM r.prep_time) <= $2::double precision / 1000000000.0)
			  AND ($3::double precision = 0 OR EXTRACT(EPOCH FROM r.active_time) <= $3::double precision / 1000000000.0)
	`

	// Build arguments slice - convert data.Duration to float64 nanoseconds for database query
	args := []any{name, float64(time.Duration(prepTime)), float64(time.Duration(activeTime))}
	argPos := 4

	// Add ingredients filter if provided
	if len(ingredients) > 0 {
		query += ` AND r.id IN (
			SELECT ri.recipe_id
			FROM recipe_ingredients ri
			JOIN ingredients i ON ri.ingredient_id = i.id
			WHERE i.name ILIKE ANY($` + fmt.Sprint(argPos) + `)
		)`
		// Convert ingredients to lowercase for case-insensitive matching
		lowerIngredients := make([]string, len(ingredients))
		for i, ing := range ingredients {
			lowerIngredients[i] = "%" + ing + "%"
		}
		args = append(args, lowerIngredients)
		argPos++
	}

	// Add equipment filter if provided
	if len(equipment) > 0 {
		query += ` AND r.id IN (
			SELECT re.recipe_id
			FROM recipe_equipment re
			JOIN equipment e ON re.equipment_id = e.id
			WHERE e.name ILIKE ANY($` + fmt.Sprint(argPos) + `)
		)`
		// Convert equipment to lowercase for case-insensitive matching
		lowerEquipment := make([]string, len(equipment))
		for i, eq := range equipment {
			lowerEquipment[i] = "%" + eq + "%"
		}
		args = append(args, lowerEquipment)
		argPos++
	}

	// Close the CTE and build main query with COUNT(*) OVER()
	query += `
		)
		SELECT COUNT(*) OVER() as total_records,
		       fr.id, fr.name, fr.description, fr.prep_time, fr.active_time,
		       fr.servings, fr.created_at, fr.version,
		       ri.image_url as display_url
		FROM filtered_recipes fr
		LEFT JOIN recipe_images ri ON fr.id = ri.recipe_id AND ri.image_type = 'main'
	`

	// Add ORDER BY clause
	sortColumn := filters.Sort
	sortDirection := "ASC"
	if len(sortColumn) > 0 && sortColumn[0] == '-' {
		sortDirection = "DESC"
		sortColumn = sortColumn[1:]
	}

	// Map sort column names to database columns
	sortColumns := map[string]string{
		"id":          "fr.id",
		"name":        "fr.name",
		"prep_time":   "fr.prep_time",
		"active_time": "fr.active_time",
	}

	if dbColumn, ok := sortColumns[sortColumn]; ok {
		query += fmt.Sprintf(" ORDER BY %s %s", dbColumn, sortDirection)
	} else {
		query += " ORDER BY fr.id ASC"
	}

	// Add LIMIT and OFFSET for pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	recipes := []*Recipe{}

	for rows.Next() {
		var recipe Recipe
		var description sql.NullString
		var prepTime, activeTime sql.NullInt64
		var servings sql.NullInt32
		var displayURL sql.NullString

		err := rows.Scan(
			&totalRecords,
			&recipe.ID,
			&recipe.Name,
			&description,
			&prepTime,
			&activeTime,
			&servings,
			&recipe.CreatedAt,
			&recipe.Version,
			&displayURL,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Handle NULL values
		if description.Valid {
			recipe.Description = description.String
		}
		if prepTime.Valid {
			recipe.PrepTime = Duration(prepTime.Int64)
		}
		if activeTime.Valid {
			recipe.ActiveTime = Duration(activeTime.Int64)
		}
		if servings.Valid {
			recipe.Servings = servings.Int32
		}
		if displayURL.Valid {
			recipe.DisplayURL = displayURL.String
		}

		recipes = append(recipes, &recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return recipes, metadata, nil
}
