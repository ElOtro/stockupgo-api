package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

type ContactInput struct {
	Role    int                  `json:"role"`
	Title   string               `json:"title"`
	Name    string               `json:"name"`
	Phone   string               `json:"phone"`
	Email   string               `json:"email"`
	StartAt *time.Time           `json:"start_at"`
	Details *data.ContactDetails `json:"details,omitempty"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listContactsHandler(w http.ResponseWriter, r *http.Request) {
	// here companyID is organisation_id
	companyID, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the GetAll() method to retrieve the contacts, passing in the various filter
	// parameters.
	contacts, err := app.models.Contacts.GetAll(companyID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the contact data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": contacts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createContactHandler(w http.ResponseWriter, r *http.Request) {
	companyID, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP request body
	var input struct {
		Contact *ContactInput `json:"contact"`
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

	var fields = input.Contact

	contact := &data.Contact{
		Role:    fields.Role,
		Title:   fields.Title,
		Name:    fields.Name,
		Phone:   fields.Phone,
		Email:   fields.Email,
		StartAt: fields.StartAt,
		Details: fields.Details,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateContact(v, contact); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Contacts.Insert(companyID, contact)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/contacts/%d", contact.ID))

	// Write a JSON response with a 201 Created status code, the contact data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": contact}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showContactHandler(w http.ResponseWriter, r *http.Request) {
	companyID, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific contact. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	contact, err := app.models.Contacts.Get(companyID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": contact}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateContactHandler(w http.ResponseWriter, r *http.Request) {
	companyID, err := app.readIDParam("companyID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Extract the contact ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing contact record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	contact, err := app.models.Contacts.Get(companyID, id)
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
		Contact *ContactInput `json:"contact"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var fields = input.Contact

	contact.CompanyID = companyID
	contact.Role = fields.Role
	contact.Title = fields.Title
	contact.Name = fields.Name
	contact.Phone = fields.Phone
	contact.Email = fields.Email
	contact.StartAt = fields.StartAt
	contact.Details = fields.Details

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateContact(v, contact); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated contact record to our new Update() method.
	err = app.models.Contacts.Update(contact)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated contact record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": contact}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the contact ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the contact from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Contacts.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "contact successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
