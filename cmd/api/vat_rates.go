package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listVatRatesHandler(w http.ResponseWriter, r *http.Request) {

	// Call the GetAll() method to retrieve the vatRates, passing in the various filter
	// parameters.
	vatRates, err := app.models.VatRates.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the vatRate data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": vatRates}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createVatRateHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		IsActive  bool    `json:"is_active"`
		IsDefault bool    `json:"is_default"`
		Rate      float64 `json:"rate"`
		Name      string  `json:"name"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		// Use the new badRequestResponse() helper.
		app.badRequestResponse(w, r, err)
		return
	}

	vatRate := &data.VatRate{
		IsActive:  input.IsActive,
		IsDefault: input.IsDefault,
		Rate:      input.Rate,
		Name:      input.Name,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateVatRate(v, vatRate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.VatRates.Insert(vatRate)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/vat_rates/%d", vatRate.ID))

	// Write a JSON response with a 201 Created status code, the vatRate data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": vatRate}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showVatRateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("vatRateID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific vatRate. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	vatRate, err := app.models.VatRates.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": vatRate}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateVatRateHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the vatRate ID from the URL.
	id, err := app.readIDParam("vatRateID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing vatRate record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	vatRate, err := app.models.VatRates.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Declare an input struct to hold the expected data from the client.
	var input struct {
		IsActive  bool      `json:"is_active"`
		IsDefault bool      `json:"is_default"`
		Rate      float64   `json:"rate"`
		Name      string    `json:"name"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	vatRate.IsActive = input.IsActive
	vatRate.IsDefault = input.IsDefault
	vatRate.Rate = input.Rate
	vatRate.Name = input.Name

	// Validate the updated vatRate record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateVatRate(v, vatRate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated vatRate record to our new Update() method.
	err = app.models.VatRates.Update(vatRate)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated vatRate record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": vatRate}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteVatRateHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the vatRate ID from the URL.
	id, err := app.readIDParam("vatRateID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the vatRate from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.VatRates.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "vatRate successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
