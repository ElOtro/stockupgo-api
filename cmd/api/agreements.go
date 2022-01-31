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
func (app *application) listAgreementsHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	var input struct {
		CompanyID int64
		data.Pagination
		data.AgreementFilters
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.AgreementFilters.CompanyID = app.readInt64(qs, "company_id", 0, v)
	input.AgreementFilters.Start = app.readDate(qs, "start", nil, v)
	input.AgreementFilters.End = app.readDate(qs, "end", nil, v)
	// Read the page and limit query string values into the embedded struct.
	input.Pagination.Page = app.readInt(qs, "page", 1, v)
	input.Pagination.Limit = app.readInt(qs, "limit", 20, v)

	// Read the sort query string value into the embedded struct.
	input.Pagination.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Pagination.SortSafelist = []string{"id", "name", "created_at"}

	// Read the sort query string value into the embedded struct.
	input.Pagination.Direction = app.readString(qs, "direction", "asc")
	input.Pagination.DirectionSafelist = []string{"asc", "desc"}

	// Execute the validation checks on the Pagination struct and send a response
	// containing the errors if necessary.
	if data.ValidatePagination(v, input.Pagination); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the agreements, passing in the various filter
	// parameters.
	agreements, metadata, err := app.models.Agreements.GetAll(input.AgreementFilters, input.Pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the agreement data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": agreements, "meta": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createAgreementHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		StartAt   *time.Time `json:"start_at"`
		EndAt     *time.Time `json:"end_at"`
		Name      string     `json:"name"`
		CompanyID int64      `json:"company_id"`
		UserID    *int64     `json:"user_id"`
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

	agreement := &data.Agreement{
		StartAt:   input.StartAt,
		EndAt:     input.EndAt,
		Name:      input.Name,
		CompanyID: input.CompanyID,
		UserID:    input.UserID,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateAgreement(v, agreement); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Agreements.Insert(agreement)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/agreements/%d", agreement.ID))

	// Write a JSON response with a 201 Created status code, the agreement data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": agreement}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showAgreementHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("agreementID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific agreement. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	agreement, err := app.models.Agreements.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": agreement}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateAgreementHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the agreement ID from the URL.
	id, err := app.readIDParam("agreementID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing agreement record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	agreement, err := app.models.Agreements.Get(id)
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
		StartAt   *time.Time `json:"start_at"`
		EndAt     *time.Time `json:"end_at"`
		Name      string     `json:"name"`
		CompanyID int64      `json:"company_id"`
		UserID    *int64     `json:"user_id"`
		UpdatedAt time.Time  `json:"updated_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	agreement.StartAt = input.StartAt
	agreement.EndAt = input.EndAt
	agreement.Name = input.Name
	agreement.CompanyID = input.CompanyID
	agreement.UserID = input.UserID

	// Validate the updated agreement record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateAgreement(v, agreement); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated agreement record to our new Update() method.
	err = app.models.Agreements.Update(agreement)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated agreement record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": agreement}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteAgreementHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the agreement ID from the URL.
	id, err := app.readIDParam("agreementID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the agreement from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Agreements.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "agreement successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
