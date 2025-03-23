package main

import (
	"fmt"
	"net/http"
	"time"

	"eatinn.dcashman.net/internal/data"
	"eatinn.dcashman.net/internal/validator"
)

func (app *application) showRecipeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Create a new instance of the Recipe struct, containing the ID we extracted from
	// the URL and some dummy data.
	recipe := data.Recipe{
		ID:                id,
		CreatedAt:         time.Now(),
		Name:              "Beef Pot Roast",
		Ingredients:       []data.IngredientEntry{{Ingredient: "Chuck steak", Amount: "5lb"}},
		RequiredEquipment: []string{"Dutch Oven"},
		Instructions:      []string{"Step 1", "Step 2"},
		Notes:             "Notes",
		DisplayURL:        "https://xkcd.com",
		SourceURL:         "youtube.com",
		// AND MORE
		Version: 1,
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset
	// of the Recipe struct. This struct will be our *target decode destination*.
	var input struct {
		Name              string                 `json:"name"`
		Ingredients       []data.IngredientEntry `json:"ingredients"`
		RequiredEquipment []string               `json:"required_equipment"`
		Instructions      []string               `json:"instructions"`
		Notes             string                 `json:"notes"`
		DisplayURL        string                 `json:"display_url"`
		SourceURL         string                 `json:"source_url"`
		PrepTime          time.Duration          `json:"prep_time"`
		ActiveTime        time.Duration          `json:"active_time"`
		Public            bool                   `json:"public"`
		Servings          int32                  `json:"servings"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipe := &data.Recipe{
		Name:              input.Name,
		Ingredients:       input.Ingredients,
		RequiredEquipment: input.RequiredEquipment,
		Instructions:      input.Instructions,
		Notes:             input.Notes,
		DisplayURL:        input.DisplayURL,
		SourceURL:         input.SourceURL,
		PrepTime:          input.PrepTime,
		ActiveTime:        input.ActiveTime,
		Public:            input.Public,
		Servings:          input.Servings,
	}

	// Validate data received.
	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Dump the contents of the input struct in a HTTP response.
	fmt.Fprintf(w, "%+v\n", input)
}
