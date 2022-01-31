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
func (app *application) listProjectsHandler(w http.ResponseWriter, r *http.Request) {

	// Call the GetAll() method to retrieve the projects, passing in the various filter
	// parameters.
	projects, err := app.models.Projects.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the project data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": projects}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		OrganisationID int64  `json:"organisation_id"`
		Name           string `json:"name"`
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

	project := &data.Project{
		OrganisationID: input.OrganisationID,
		Name:           input.Name,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateProject(v, project); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Projects.Insert(project)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/projects/%d", project.ID))

	// Write a JSON response with a 201 Created status code, the project data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": project}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showProjectHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("projectID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific project. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	project, err := app.models.Projects.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": project}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the project ID from the URL.
	id, err := app.readIDParam("projectID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing project record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	project, err := app.models.Projects.Get(id)
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
		OrganisationID int64     `json:"organisation_id"`
		Name           string    `json:"name"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	project.OrganisationID = input.OrganisationID
	project.Name = input.Name

	// Validate the updated project record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateProject(v, project); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated project record to our new Update() method.
	err = app.models.Projects.Update(project)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated project record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": project}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the project ID from the URL.
	id, err := app.readIDParam("projectID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the project from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Projects.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "project successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
