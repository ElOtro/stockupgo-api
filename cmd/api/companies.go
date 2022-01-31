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
func (app *application) listCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	var input struct {
		data.Pagination
		data.CompanyFilters
	}

	// Initialize a new Validator instance.
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.CompanyFilters.OrganisationID = app.readInt64(qs, "organisation_id", 0, v)
	// Read the page and limit query string values into the embedded struct.
	input.Pagination.Page = app.readInt(qs, "page", 1, v)
	input.Pagination.Limit = app.readInt(qs, "limit", 20, v)

	// Read the sort query string value into the embedded struct.
	input.Pagination.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Pagination.SortSafelist = []string{"id", "number", "created_at"}
	// Read the sort query string value into the embedded struct.
	input.Pagination.Direction = app.readString(qs, "direction", "asc")
	input.Pagination.DirectionSafelist = []string{"asc", "desc"}

	// Execute the validation checks on the Pagination struct and send a response
	// containing the errors if necessary.
	if data.ValidatePagination(v, input.Pagination); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the companies, passing in the various filter
	// parameters.
	companies, metadata, err := app.models.Companies.GetAll(input.CompanyFilters, input.Pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the company data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": companies, "meta": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) searchCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	var input struct {
		data.CompanyFilters
	}

	// Initialize a new Validator instance.
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.CompanyFilters.OrganisationID = app.readInt64(qs, "organisation_id", 0, v)
	input.CompanyFilters.Name = app.readString(qs, "q", "")

	// Call the GetAll() method to retrieve the companies, passing in the various filter
	// parameters.
	companies, err := app.models.Companies.Search(input.CompanyFilters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the company data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": companies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createCompanyHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		OrganisationID int64               `json:"organisation_id"`
		Name           string              `json:"name"`
		FullName       string              `json:"full_name"`
		CompanyType    int                 `json:"company_type"`
		Details        data.CompanyDetails `json:"details"`
		Contacts       []data.Contact      `json:"contacts"`
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

	company := &data.Company{
		OrganisationID: input.OrganisationID,
		Name:           input.Name,
		FullName:       input.FullName,
		CompanyType:    input.CompanyType,
		Details:        &input.Details,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateCompany(v, company); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	contacts := company.Contacts
	for _, c := range input.Contacts {
		contact := &data.Contact{
			Role:    c.Role,
			Title:   c.Title,
			Name:    c.Name,
			Phone:   c.Phone,
			Email:   c.Email,
			StartAt: c.StartAt,
		}

		if data.ValidateContact(v, contact); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		contacts = append(contacts, contact)
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Companies.Insert(company)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Call the Insert() method on our contacts
	for _, c := range contacts {
		err = app.models.Contacts.Insert(company.ID, c)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	company.Contacts = contacts

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/companies/%d", company.ID))

	// Write a JSON response with a 201 Created status code, the company data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": company}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showCompanyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific company. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	company, err := app.models.Companies.Get(id)
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
	contacts, err := app.models.Contacts.GetAll(id)
	if err != nil {
		app.logger.Err(err).Msg("errors in getting contacts")
	}

	company.Contacts = contacts

	err = app.writeJSON(w, http.StatusOK, envelope{"data": company}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the company ID from the URL.
	id, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing company record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	company, err := app.models.Companies.Get(id)
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
		OrganisationID int64               `json:"organisation_id"`
		Name           string              `json:"name"`
		FullName       string              `json:"full_name"`
		CompanyType    int                 `json:"company_type"`
		Details        data.CompanyDetails `json:"details"`
		UpdatedAt      time.Time           `json:"updated_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	company.Name = input.Name
	company.FullName = input.FullName
	company.CompanyType = input.CompanyType
	company.Details = &input.Details

	// Validate the updated company record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateCompany(v, company); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated company record to our new Update() method.
	err = app.models.Companies.Update(company)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated company record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": company}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the company ID from the URL.
	id, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the company from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Companies.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "company successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
