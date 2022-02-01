package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

type BankAccountInput struct {
	IsDefault bool                     `json:"is_default,omitempty"`
	Name      string                   `json:"name"`
	Details   *data.BankAccountDetails `json:"details,omitempty"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listBankAccountsHandler(w http.ResponseWriter, r *http.Request) {
	// here organisationID is organisation_id
	organisationID, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the GetAll() method to retrieve the movies, passing in the various filter
	// parameters.
	bankAccounts, err := app.models.BankAccounts.GetAll(organisationID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": bankAccounts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	organisationID, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP request body
	var input struct {
		BankAccount *BankAccountInput `json:"bank_account"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err = app.readJSON(w, r, &input)
	if err != nil {
		// Use the new badRequestResponse() helper.
		app.badRequestResponse(w, r, err)
		return
	}

	var fields = input.BankAccount

	bankAccount := &data.BankAccount{
		IsDefault: fields.IsDefault,
		Name:      fields.Name,
		Details:   fields.Details,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateBankAccount(v, bankAccount); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.BankAccounts.Insert(organisationID, bankAccount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/bank_accounts/%d", bankAccount.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": bankAccount}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	organisationID, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	bankAccount, err := app.models.BankAccounts.Get(organisationID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": bankAccount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	organisationID, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Extract the movie ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	bankAccount, err := app.models.BankAccounts.Get(organisationID, id)
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
		BankAccount *BankAccountInput `json:"bank_account"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var fields = input.BankAccount

	bankAccount.IsDefault = fields.IsDefault
	bankAccount.Name = fields.Name
	bankAccount.Details = fields.Details

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateBankAccount(v, bankAccount); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated movie record to our new Update() method.
	err = app.models.BankAccounts.Update(bankAccount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": bankAccount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.BankAccounts.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "bank_account successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
