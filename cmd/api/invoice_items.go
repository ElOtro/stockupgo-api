package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listInvoiceItemsHandler(w http.ResponseWriter, r *http.Request) {
	// here invoiceID is organisation_id
	invoiceID, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to check if invoice exists.
	_, err = app.models.Invoices.Get(invoiceID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Call the GetAll() method to retrieve the invoice_items, passing in the various filter
	// parameters.
	invoiceItems, err := app.models.InvoiceItems.GetAll(invoiceID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the invoice_item data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": invoiceItems}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createInvoiceItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the invoice ID from the URL.
	invoiceID, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to check if invoice exists.
	_, err = app.models.Invoices.Get(invoiceID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Declare an anonymous struct to hold the information that we expect to be in the HTTP request body
	var input struct {
		InvoiceID    int64   `json:"invoice_id,omitempty"`
		Position     int     `json:"position"`
		ProductID    int64   `json:"product_id,omitempty"`
		Description  string  `json:"description"`
		UnitID       int64   `json:"unit_id,omitempty"`
		Quantity     float64 `json:"quantity"`
		Price        float64 `json:"price"`
		Amount       float64 `json:"amount"`
		DiscountRate int     `json:"discount_rate"`
		Discount     float64 `json:"discount"`
		VatRateID    int64   `json:"vat_rate_id,omitempty"`
		Vat          float64 `json:"vat,omitempty"`
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

	input.InvoiceID = invoiceID

	invoiceItem := &data.InvoiceItem{
		InvoiceID:    input.InvoiceID,
		Position:     input.Position,
		ProductID:    input.ProductID,
		Description:  input.Description,
		UnitID:       input.UnitID,
		Quantity:     input.Quantity,
		Price:        input.Price,
		Amount:       input.Amount,
		DiscountRate: input.DiscountRate,
		Discount:     input.Discount,
		VatRateID:    input.VatRateID,
		Vat:          input.Vat,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateInvoiceItem(v, invoiceItem); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.InvoiceItems.Insert(invoiceItem.InvoiceID, invoiceItem)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Update totals in the invoice
	err = app.models.Invoices.UpdateTotals(invoiceItem.InvoiceID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/invoice_items/%d", invoiceItem.ID))

	responseInvoiceItem := data.InvoiceItem{
		ID:           invoiceItem.ID,
		Position:     invoiceItem.Position,
		Product:      invoiceItem.Product,
		Description:  invoiceItem.Description,
		Unit:         invoiceItem.Unit,
		Quantity:     invoiceItem.Quantity,
		Price:        invoiceItem.Price,
		Amount:       invoiceItem.Amount,
		DiscountRate: invoiceItem.DiscountRate,
		Discount:     invoiceItem.Discount,
		VatRate:      invoiceItem.VatRate,
		CreatedAt:    invoiceItem.CreatedAt,
		UpdatedAt:    invoiceItem.UpdatedAt,
	}

	// Write a JSON response with a 201 Created status code, the invoice_item data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": responseInvoiceItem}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showInvoiceItemHandler(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to check if invoice exists.
	_, err = app.models.Invoices.Get(invoiceID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Call the Get() method to fetch the data for a specific invoice_item. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	invoiceItem, err := app.models.InvoiceItems.Get(invoiceID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": invoiceItem}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateInvoiceItemHandler(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Extract the invoice_item ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to check if invoice exists.
	_, err = app.models.Invoices.Get(invoiceID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Fetch the existing invoice_item record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	invoiceItem, err := app.models.InvoiceItems.Get(invoiceID, id)
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
		Position     *int     `json:"position"`
		ProductID    *int64   `json:"product_id"`
		Description  *string  `json:"description"`
		UnitID       *int64   `json:"unit_id"`
		Quantity     *float64 `json:"quantity"`
		Price        *float64 `json:"price"`
		Amount       *float64 `json:"amount"`
		DiscountRate *int     `json:"discount_rate"`
		Discount     *float64 `json:"discount"`
		VatRateID    *int64   `json:"vat_rate_id"`
		Vat          *float64 `json:"vat"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Position != nil {
		invoiceItem.Position = *input.Position
	}

	if input.ProductID != nil {
		invoiceItem.ProductID = *input.ProductID
	}

	if input.Description != nil {
		invoiceItem.Description = *input.Description
	}

	if input.UnitID != nil {
		invoiceItem.UnitID = *input.UnitID
	}

	if input.Quantity != nil {
		invoiceItem.Quantity = *input.Quantity
	}

	if input.Price != nil {
		invoiceItem.Price = *input.Price
	}

	if input.Amount != nil {
		invoiceItem.Amount = *input.Amount
	}

	if input.DiscountRate != nil {
		invoiceItem.DiscountRate = *input.DiscountRate
	}

	if input.Discount != nil {
		invoiceItem.Discount = *input.Discount
	}

	if input.VatRateID != nil {
		invoiceItem.VatRateID = *input.VatRateID
	}

	if input.Vat != nil {
		invoiceItem.Vat = *input.Vat
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call vakidate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateInvoiceItem(v, invoiceItem); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated invoice_item record to our new Update() method.
	err = app.models.InvoiceItems.Update(invoiceItem)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Update totals in the invoice
	err = app.models.Invoices.UpdateTotals(invoiceItem.InvoiceID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	responseInvoiceItem := data.InvoiceItem{
		ID:           invoiceItem.ID,
		Position:     invoiceItem.Position,
		Product:      invoiceItem.Product,
		Description:  invoiceItem.Description,
		Unit:         invoiceItem.Unit,
		Quantity:     invoiceItem.Quantity,
		Price:        invoiceItem.Price,
		Amount:       invoiceItem.Amount,
		DiscountRate: invoiceItem.DiscountRate,
		Discount:     invoiceItem.Discount,
		VatRate:      invoiceItem.VatRate,
		CreatedAt:    invoiceItem.CreatedAt,
		UpdatedAt:    invoiceItem.UpdatedAt,
	}

	// Write the updated invoice_item record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": responseInvoiceItem}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteInvoiceItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the invoice ID from the URL.
	invoiceID, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Extract the invoice_item ID from the URL.
	id, err := app.readIDParam("ID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the invoice_item from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.InvoiceItems.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update totals in the invoice
	err = app.models.Invoices.UpdateTotals(invoiceID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "invoice_item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
