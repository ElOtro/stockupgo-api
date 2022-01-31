package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/pascaldekloe/jwt"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup the user record based on the email address. If no matching user was
	// found, then we call the app.invalidCredentialsResponse() helper to send a 401
	// Unauthorized response to the client (we will create this helper in a moment).
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check if the provided password matches the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If the passwords don't match, then we call the app.invalidCredentialsResponse()
	// helper again and return.
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Create a JWT claims struct containing the user ID as the subject, with an issued
	// time of now and validity window of the next 24 hours. We also set the issuer and
	// audience to a unique identifier for our application.
	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(user.ID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "stockup-api"
	claims.Audiences = []string{"stockup-api"}

	// Sign the JWT claims using the HMAC-SHA256 algorithm and the secret key from the // application config. This returns a []byte slice containing the JWT as a base64- // encoded string.
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//Encode the token to JSON and send it in the response along with a 201 Created
	//status code.
	err = app.writeJSON(w, http.StatusCreated, envelope{"token": string(jwtBytes)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showUserHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	// Call the Get() method to fetch the data for a specific unit. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	// user, err := app.models.Users.GetByEmail(input.Email)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, data.ErrRecordNotFound):
	// 		app.notFoundResponse(w, r)
	// 	default:
	// 		app.serverErrorResponse(w, r, err)
	// 	}
	// 	return
	// }

	err := app.writeJSON(w, http.StatusOK, envelope{"data": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
