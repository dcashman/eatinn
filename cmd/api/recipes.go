package main

import (
	"errors"
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

	// Fetch the recipe from the database
	recipe, err := app.models.Recipes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
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
		Instructions      []data.InstructionStep `json:"instructions"`
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

	// TODO: convert all strings to lower-case where appropriate.
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

	// Call the Insert() method on our recipe model, passing in a pointer to the
	// validated movie struct. This will create a record in the database and update the
	// recipe struct with the system-generated information.
	err = app.models.Recipes.Insert(recipe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new recipe in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/recipes/%d", recipe.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"recipe": recipe}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing recipe
	recipe, err := app.models.Recipes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Parse the request body
	var input struct {
		Name              *string                 `json:"name"`
		Description       *string                 `json:"description"`
		Ingredients       []data.IngredientEntry  `json:"ingredients"`
		RequiredEquipment []string                `json:"required_equipment"`
		Instructions      []data.InstructionStep  `json:"instructions"`
		Notes             *string                 `json:"notes"`
		DisplayURL        *string                 `json:"display_url"`
		SourceURL         *string                 `json:"source_url"`
		PrepTime          *time.Duration          `json:"prep_time"`
		ActiveTime        *time.Duration          `json:"active_time"`
		Servings          *int32                  `json:"servings"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Update fields if provided (partial update support)
	if input.Name != nil {
		recipe.Name = *input.Name
	}
	if input.Description != nil {
		recipe.Description = *input.Description
	}
	if input.Ingredients != nil {
		recipe.Ingredients = input.Ingredients
	}
	if input.RequiredEquipment != nil {
		recipe.RequiredEquipment = input.RequiredEquipment
	}
	if input.Instructions != nil {
		recipe.Instructions = input.Instructions
	}
	if input.Notes != nil {
		recipe.Notes = *input.Notes
	}
	if input.DisplayURL != nil {
		recipe.DisplayURL = *input.DisplayURL
	}
	if input.SourceURL != nil {
		recipe.SourceURL = *input.SourceURL
	}
	if input.PrepTime != nil {
		recipe.PrepTime = *input.PrepTime
	}
	if input.ActiveTime != nil {
		recipe.ActiveTime = *input.ActiveTime
	}
	if input.Servings != nil {
		recipe.Servings = *input.Servings
	}

	// Validate the updated recipe
	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the recipe in the database
	err = app.models.Recipes.Update(recipe)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return the updated recipe
	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Recipes.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "recipe successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
