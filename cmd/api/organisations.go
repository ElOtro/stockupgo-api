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
func (app *application) listOrganisationsHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)
	fmt.Println(user.IsActive)
	// Call the GetAll() method to retrieve the organisations, passing in the various filter
	// parameters.
	organisations, err := app.models.Organisations.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the organisation data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": organisations}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		Name         string                   `json:"name"`
		FullName     string                   `json:"full_name"`
		CEO          string                   `json:"ceo"`
		CEOTitle     string                   `json:"ceo_title"`
		CFO          string                   `json:"cfo"`
		CFOTitle     string                   `json:"cfo_title"`
		Stamp        *string                  `json:"stamp"`
		CEOSign      *string                  `json:"ceo_sign"`
		CFOSign      *string                  `json:"cfo_sign"`
		IsVatPayer   bool                     `json:"is_vat_payer"`
		Details      data.OrganisationDetails `json:"details"`
		BankAccounts []data.BankAccount       `json:"bank_accounts"`
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

	organisation := &data.Organisation{
		Name:       input.Name,
		FullName:   input.FullName,
		CEO:        input.CEO,
		CEOTitle:   input.CEOTitle,
		CFO:        input.CFO,
		CFOTitle:   input.CFOTitle,
		Stamp:      input.Stamp,
		CEOSign:    input.CEOSign,
		CFOSign:    input.CFOSign,
		IsVatPayer: input.IsVatPayer,
		Details:    &input.Details,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateOrganisation(v, organisation); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	bankAccounts := organisation.BankAccounts
	for _, a := range input.BankAccounts {
		bankAccount := &data.BankAccount{
			IsDefault: a.IsDefault,
			Name:      a.Name,
			Details:   a.Details,
		}

		if data.ValidateBankAccount(v, bankAccount); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		bankAccounts = append(bankAccounts, bankAccount)
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Organisations.Insert(organisation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Call the Insert() method on our bank_accounts
	for _, a := range bankAccounts {
		err = app.models.BankAccounts.Insert(organisation.ID, a)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	organisation.BankAccounts = bankAccounts

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/organisations/%d", organisation.ID))

	// Write a JSON response with a 201 Created status code, the organisation data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": organisation}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific organisation. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	organisation, err := app.models.Organisations.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// get all bank accounts
	bankAccounts, err := app.models.BankAccounts.GetAll(id)
	if err != nil {
		app.logger.Err(err).Msg("errors in getting bank_accounts")
	}

	organisation.BankAccounts = bankAccounts

	err = app.writeJSON(w, http.StatusOK, envelope{"data": organisation}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the organisation ID from the URL.
	id, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing organisation record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	organisation, err := app.models.Organisations.Get(id)
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
		Name        *string                  `json:"name"`
		FullName    *string                  `json:"full_name"`
		CEO         *string                  `json:"ceo"`
		CEOTitle    *string                  `json:"ceo_title"`
		CFO         *string                  `json:"cfo"`
		CFOTitle    *string                  `json:"cfo_title"`
		Stamp       *string                  `json:"stamp"`
		CEOSign     *string                  `json:"ceo_sign"`
		CFOSign     *string                  `json:"cfo_sign"`
		IsVatPayer  bool                     `json:"is_vat_payer"`
		Details     data.OrganisationDetails `json:"details"`
		UpdatedAt   time.Time                `json:"updated_at"`
		DestroyedAt *time.Time               `json:"destroyed_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		organisation.Name = *input.Name
	}

	if input.FullName != nil {
		organisation.FullName = *input.FullName
	}

	if input.CEO != nil {
		organisation.CEO = *input.CEO
	}

	if input.CEOTitle != nil {
		organisation.CEOTitle = *input.CEOTitle
	}

	if input.CFO != nil {
		organisation.CFO = *input.CFO
	}

	if input.CFOTitle != nil {
		organisation.CFOTitle = *input.CFOTitle
	}

	organisation.Stamp = input.Stamp
	organisation.CEOSign = input.CEOSign
	organisation.CFOSign = input.CFOSign
	organisation.IsVatPayer = input.IsVatPayer
	organisation.Details = &input.Details

	// Validate the updated organisation record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateOrganisation(v, organisation); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated organisation record to our new Update() method.
	err = app.models.Organisations.Update(organisation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// get all bank accounts
	bankAccounts, err := app.models.BankAccounts.GetAll(id)
	if err != nil {
		app.logger.Err(err).Msg("errors in getting bank_accounts")
	}

	organisation.BankAccounts = bankAccounts

	// Write the updated organisation record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": organisation}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the organisation ID from the URL.
	id, err := app.readIDParam("organisationID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the organisation from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Organisations.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "organisation successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
