package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/ElOtro/stockup-api/internal/validator"
)

type InvoiceInput struct {
	IsActive       *bool              `json:"is_active"`
	Date           *time.Time         `json:"date"`
	Number         *string            `json:"number"`
	OrganisationID *int64             `json:"organisation_id"`
	BankAccountID  *int64             `json:"bank_account_id"`
	CompanyID      *int64             `json:"company_id"`
	AgreementID    *int64             `json:"agreement_id"`
	InvoiceItems   []data.InvoiceItem `json:"invoice_items,omitempty"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) listInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	var input struct {
		data.Pagination
		data.InvoiceFilters
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.InvoiceFilters.OrganisationID = app.readInt64(qs, "organisation_id", 0, v)
	input.InvoiceFilters.CompanyID = app.readInt64(qs, "company_id", 0, v)
	input.InvoiceFilters.AgreementID = app.readInt64(qs, "agreement_id", 0, v)
	input.InvoiceFilters.Start = app.readDate(qs, "start", nil, v)
	input.InvoiceFilters.End = app.readDate(qs, "end", nil, v)
	// Read the page and limit query string values into the embedded struct.
	input.Pagination.Page = app.readInt(qs, "page", 1, v)
	input.Pagination.Limit = app.readInt(qs, "limit", 20, v)

	// Read the sort query string value into the embedded struct.
	input.Pagination.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Pagination.SortSafelist = []string{"id", "date", "name", "number", "created_at"}

	// Read the sort query string value into the embedded struct.
	input.Pagination.Direction = app.readString(qs, "direction", "asc")
	input.Pagination.DirectionSafelist = []string{"asc", "desc"}

	// Execute the validation checks on the Pagination struct and send a response
	// containing the errors if necessary.
	if data.ValidatePagination(v, input.Pagination); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the invoices, passing in the various filter
	// parameters.
	invoices, metadata, err := app.models.Invoices.GetAll(input.InvoiceFilters, input.Pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the invoice data.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": invoices, "meta": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showInvoiceHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific invoice. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	invoice, err := app.models.Invoices.Get(id)
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
	invoiceItems, err := app.models.InvoiceItems.GetAll(id)
	if err != nil {
		app.logger.Err(err).Msg("errors in getting invoice_items")
	}

	invoice.InvoiceItems = invoiceItems

	err = app.writeJSON(w, http.StatusOK, envelope{"data": invoice}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body
	var input struct {
		Invoice *InvoiceInput `json:"invoice"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var fields = input.Invoice

	invoice := &data.Invoice{
		IsActive:       *fields.IsActive,
		Date:           *fields.Date,
		Number:         *fields.Number,
		OrganisationID: *fields.OrganisationID,
		BankAccountID:  *fields.BankAccountID,
		CompanyID:      *fields.CompanyID,
		AgreementID:    *fields.AgreementID,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the validate function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateInvoice(v, invoice); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model, passing in a pointer to the
	// validated struct.
	err = app.models.Invoices.Insert(invoice)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Call the Insert() method on our invoice_items
	invoiceItems := invoice.InvoiceItems
	for _, item := range fields.InvoiceItems {

		invoiceItem := &data.InvoiceItem{
			ID:           item.ID,
			Position:     item.Position,
			ProductID:    item.ProductID,
			Description:  item.Description,
			UnitID:       item.UnitID,
			Quantity:     item.Quantity,
			Price:        item.Price,
			Amount:       item.Amount,
			DiscountRate: item.DiscountRate,
			Discount:     item.Discount,
			VatRateID:    item.VatRateID,
		}

		if data.ValidateInvoiceItem(v, invoiceItem); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = app.models.InvoiceItems.Insert(invoice.ID, invoiceItem)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		responseInvoiceItem := &data.InvoiceItem{
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
			Vat:          invoiceItem.Vat,
			VatRate:      invoiceItem.VatRate,
			CreatedAt:    invoiceItem.CreatedAt,
			UpdatedAt:    invoiceItem.UpdatedAt,
		}

		invoiceItems = append(invoiceItems, responseInvoiceItem)
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/invoices/%d", invoice.ID))

	// responseInvoiceItems := invoice.InvoiceItems

	responseInvoice := data.Invoice{
		ID:           invoice.ID,
		IsActive:     invoice.IsActive,
		Date:         invoice.Date,
		Number:       invoice.Number,
		Organisation: invoice.Organisation,
		BankAccount:  invoice.BankAccount,
		Company:      invoice.Company,
		Agreement:    invoice.Agreement,
		CreatedAt:    invoice.CreatedAt,
		UpdatedAt:    invoice.UpdatedAt,
		InvoiceItems: invoiceItems,
	}

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": responseInvoice}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the invoice ID from the URL.
	id, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing invoice record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	invoice, err := app.models.Invoices.Get(id)
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
		Invoice *InvoiceInput `json:"invoice"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var fields = input.Invoice

	if fields.IsActive != nil {
		invoice.IsActive = *fields.IsActive
	}

	if fields.Date != nil {
		invoice.Date = *fields.Date
	}

	if fields.Number != nil {
		invoice.Number = *fields.Number
	}

	if fields.OrganisationID != nil {
		invoice.OrganisationID = *fields.OrganisationID
	}

	if fields.BankAccountID != nil {
		invoice.BankAccountID = *fields.BankAccountID
	}

	if fields.CompanyID != nil {
		invoice.CompanyID = *fields.CompanyID
	}

	if fields.AgreementID != nil {
		invoice.AgreementID = *fields.AgreementID
	}

	// Validate the updated invoice record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateInvoice(v, invoice); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated invoice record to our new Update() method.
	err = app.models.Invoices.Update(invoice)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	responseInvoice := data.Invoice{
		ID:           invoice.ID,
		IsActive:     invoice.IsActive,
		Date:         invoice.Date,
		Number:       invoice.Number,
		Organisation: invoice.Organisation,
		BankAccount:  invoice.BankAccount,
		Company:      invoice.Company,
		Agreement:    invoice.Agreement,
		CreatedAt:    invoice.CreatedAt,
		UpdatedAt:    invoice.UpdatedAt,
	}

	// Write the updated invoice record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"data": responseInvoice}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the invoice ID from the URL.
	id, err := app.readIDParam("invoiceID", r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the invoice from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Invoices.Delete(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "invoice successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
