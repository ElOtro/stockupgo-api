package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

type ProductInput struct {
	ID          *int64  `json:"id"`
	IsActive    bool    `json:"is_active"`
	ProductType int     `json:"product_type"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SKU         string  `json:"sku"`
	Price       float64 `json:"price"`
	VatRateID   *int64  `json:"vat_rate_id"`
	UnitID      *int64  `json:"unit_id"`
	UserID      *int64  `json:"user_id"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listProductsHandler(w http.ResponseWriter, r *http.Request) {

	// Call the GetAll() method to retrieve the products, passing in the various filter
	// parameters.
	products, err := app.models.Products.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the product data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": products}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProductHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		Product *ProductInput `json:"product"`
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

	var fields = input.Product

	product := &data.Product{
		IsActive:    fields.IsActive,
		ProductType: fields.ProductType,
		Name:        fields.Name,
		Description: fields.Description,
		SKU:         fields.SKU,
		Price:       fields.Price,
		VatRateID:   fields.VatRateID,
		UnitID:      fields.UnitID,
		UserID:      fields.UserID,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Products.Insert(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/products/%d", product.ID))

	// Write a JSON response with a 201 Created status code, the product data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": product}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam("productID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific product. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	product, err := app.models.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID from the URL.
	id, err := app.readIDParam("productID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing product record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	product, err := app.models.Products.Get(id)
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
		Product *ProductInput `json:"product"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var fields = input.Product

	product.IsActive = fields.IsActive
	product.ProductType = fields.ProductType
	product.Name = fields.Name
	product.Description = fields.Description
	product.SKU = fields.SKU
	product.Price = fields.Price
	product.VatRateID = fields.VatRateID
	product.UnitID = fields.UnitID
	product.UserID = fields.UserID

	// Validate the updated product record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated product record to our new Update() method.
	err = app.models.Products.Update(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated product record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID from the URL.
	id, err := app.readIDParam("productID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the product from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Products.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "product successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
