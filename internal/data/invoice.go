package data

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Invoice type details
type Invoice struct {
	ID             int64          `json:"id"`
	IsActive       bool           `json:"is_active"`
	Date           time.Time      `json:"date"`
	Number         string         `json:"number"`
	OrganisationID int64          `json:"organisation_id,omitempty"`
	BankAccountID  int64          `json:"bank_account_id,omitempty"`
	CompanyID      int64          `json:"company_id,omitempty"`
	AgreementID    int64          `json:"agreement_id,omitempty"`
	Amount         float64        `json:"amount"`
	Discount       float64        `json:"discount"`
	Vat            float64        `json:"vat"`
	UserID         *int64         `json:"user_id,omitempty"`
	UUID           string         `json:"uuid,omitempty"`
	DestroyedAt    *time.Time     `json:"destroyed_at,omitempty"`
	Organisation   *Organisation  `json:"organisation,omitempty"`
	BankAccount    *BankAccount   `json:"bank_account,omitempty"`
	Company        *Company       `json:"company,omitempty"`
	Agreement      *Agreement     `json:"agreement,omitempty"`
	InvoiceItems   []*InvoiceItem `json:"invoice_items,omitempty"`
	CreatedAt      *time.Time     `json:"created_at,omitempty"`
	UpdatedAt      *time.Time     `json:"updated_at,omitempty"`
}

type InvoiceFilters struct {
	OrganisationID int64
	CompanyID      int64
	AgreementID    int64
	Start          *time.Time
	End            *time.Time
}

func ValidateInvoice(v *validator.Validator, invoice *Invoice) {
	v.Check(invoice.OrganisationID != 0, "organisation_id", "must be provided")
	v.Check(invoice.CompanyID != 0, "company_id", "must be provided")
}

// Define a InvoiceModel struct type which wraps a pgx.Conn connection pool.
type InvoiceModel struct {
	DB *pgxpool.Pool
}

func (m InvoiceModel) GetAll(filters InvoiceFilters, pagination Pagination) ([]*Invoice, Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	queryElements := []string{}
	filterQuery := ""
	q := ""

	if filters.OrganisationID > 0 {
		q = fmt.Sprintf("organisation_id = %d", filters.OrganisationID)
		queryElements = append(queryElements, q)
	}

	if filters.CompanyID > 0 {
		q = fmt.Sprintf("company_id = %d", filters.CompanyID)
		queryElements = append(queryElements, q)
	}

	if filters.AgreementID > 0 {
		q = fmt.Sprintf("agreement_id = %d", filters.AgreementID)
		queryElements = append(queryElements, q)
	}

	if filters.Start != nil && filters.End != nil {
		q = fmt.Sprintf("date BETWEEN '%s' AND '%s'", filters.Start.Format(time.RFC3339), filters.End.Format(time.RFC3339))
		queryElements = append(queryElements, q)
	}

	q = "destroyed_at IS NULL"
	queryElements = append(queryElements, q)

	if len(queryElements) > 0 {
		filterQuery = " WHERE " + strings.Join(queryElements, " AND ") + " "
	}
	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf(`
	SELECT id, is_active, date, number, amount, discount, vat, 
		(SELECT row_to_json(row) FROM (SELECT id, name FROM organisations WHERE organisations.id = organisation_id) row) AS organisation,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM bank_accounts WHERE bank_accounts.id = bank_account_id) row) AS bank_account,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM companies WHERE companies.id = company_id) row) AS company,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM agreements WHERE agreements.id = agreement_id) row) AS agreement,   
		user_id, uuid, created_at, updated_at 
	FROM invoices 
	%s
	ORDER BY %s %s
	LIMIT $1 OFFSET $2`, filterQuery, pagination.sortColumn(), pagination.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.DB.Query(ctx, query, pagination.limit(), pagination.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	invoices := []*Invoice{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var invoice Invoice

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&invoice.ID,
			&invoice.IsActive,
			&invoice.Date,
			&invoice.Number,
			&invoice.Amount,
			&invoice.Discount,
			&invoice.Vat,
			&invoice.Organisation,
			&invoice.BankAccount,
			&invoice.Company,
			&invoice.Agreement,
			&invoice.UserID,
			&invoice.UUID,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Invoice struct to the slice.
		invoices = append(invoices, &invoice)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	totalRecords, err := m.CountIDs(filterQuery)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, pagination.Page, pagination.Limit)

	return invoices, metadata, nil
}

// Add method for inserting a new record in the Invoices table.
func (m InvoiceModel) Insert(invoice *Invoice) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO invoices (
			is_active, date, number, organisation_id, bank_account_id, company_id, agreement_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, is_active, date, number, amount, discount, vat,
				  (SELECT row_to_json(row) FROM (SELECT id, name FROM organisations WHERE organisations.id = organisation_id) row) AS organisation,
		          (SELECT row_to_json(row) FROM (SELECT id, name FROM bank_accounts WHERE bank_accounts.id = bank_account_id) row) AS bank_account,
		          (SELECT row_to_json(row) FROM (SELECT id, name FROM companies WHERE companies.id = company_id) row) AS company,
		          (SELECT row_to_json(row) FROM (SELECT id, name FROM agreements WHERE agreements.id = agreement_id) row) AS agreement,  
				  uuid, created_at, updated_at`

	// Set new number
	if invoice.Number == "" {
		number, err := m.GetNumber(invoice.OrganisationID)
		if err == nil {
			invoice.Number = number
		}
	}

	args := []interface{}{
		invoice.IsActive,
		invoice.Date,
		invoice.Number,
		invoice.OrganisationID,
		invoice.BankAccountID,
		invoice.CompanyID,
		invoice.AgreementID,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&invoice.ID,
		&invoice.IsActive,
		&invoice.Date,
		&invoice.Number,
		&invoice.Amount,
		&invoice.Discount,
		&invoice.Vat,
		&invoice.Organisation,
		&invoice.BankAccount,
		&invoice.Company,
		&invoice.Agreement,
		&invoice.UUID,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)
}

// Add method for fetching a specific record from the invoices table.
func (m InvoiceModel) Get(id int64) (*Invoice, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no invoices will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
	SELECT id, is_active, date, number, amount, discount, vat, 
		organisation_id,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM organisations WHERE organisations.id = organisation_id) row) AS organisation,
		bank_account_id,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM bank_accounts WHERE bank_accounts.id = bank_account_id) row) AS bank_account,
		company_id,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM companies WHERE companies.id = company_id) row) AS company,
		agreement_id,
		(SELECT row_to_json(row) FROM (SELECT id, name FROM agreements WHERE agreements.id = agreement_id) row) AS agreement,   
		user_id, uuid, created_at, updated_at    
	FROM invoices WHERE id = $1`

	// Declare a Invoice struct to hold the data returned by the query.
	var invoice Invoice

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&invoice.ID,
		&invoice.IsActive,
		&invoice.Date,
		&invoice.Number,
		&invoice.Amount,
		&invoice.Discount,
		&invoice.Vat,
		&invoice.OrganisationID,
		&invoice.Organisation,
		&invoice.BankAccountID,
		&invoice.BankAccount,
		&invoice.CompanyID,
		&invoice.Company,
		&invoice.AgreementID,
		&invoice.Agreement,
		&invoice.UserID,
		&invoice.UUID,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	// Handle any errors. If there was no matching found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &invoice, nil
}

// Add method for updating a specific record in the invoices table.
func (m InvoiceModel) Update(invoice *Invoice) error {
	query := `
		UPDATE invoices
		SET is_active = $1, date = $2, number = $3, organisation_id = $4, bank_account_id = $5, 
		company_id = $6, agreement_id = $7, amount = $8, discount = $9, vat = $10, updated_at = NOW() 
		WHERE id = $11
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		invoice.IsActive,
		invoice.Date,
		invoice.Number,
		invoice.OrganisationID,
		invoice.BankAccountID,
		invoice.CompanyID,
		invoice.AgreementID,
		invoice.Amount,
		invoice.Discount,
		invoice.Vat,
		invoice.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&invoice.UpdatedAt)
}

// Add method for deleting a specific record from the invoices table.
func (m InvoiceModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM invoices WHERE id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method, passing in the id variable as
	// the value for the placeholder parameter. The Exec() method returns a sql.Result
	// object.
	result, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows
	// affected by the query.
	rowsAffected := result.RowsAffected()

	// If no rows were affected, we know that the invoices table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Add method for deleting a specific record from the invoices table.
func (m InvoiceModel) GetNumber(organisationID int64) (string, error) {
	if organisationID < 1 {
		return "", ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := "SELECT id, number FROM invoices WHERE organisation_id = $1 ORDER BY created_at DESC LIMIT 1"

	// Declare a Invoice struct to hold the data returned by the query.
	var invoice Invoice

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Importantly, use defer to make sure that we cancel the context before the Get() method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, organisationID).Scan(
		&invoice.ID,
		&invoice.Number,
	)

	// Handle any errors. If there was no matching found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return "1", nil
		default:
			return "", err
		}
	}

	n, err := strconv.Atoi(invoice.Number)
	if err != nil {
		return "", err
	}
	number := strconv.Itoa(n + 1)
	return number, nil
}

// Add method for updating a specific record in the invoices table.
func (m InvoiceModel) UpdateTotals(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	queryItems := "SELECT COALESCE(SUM(amount), 0) as amount, COALESCE(SUM(vat), 0) as vat FROM invoice_items WHERE invoice_id = $1"

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var amount float64
	var vat float64
	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, queryItems, id).Scan(&amount, &vat)

	// Handle any errors. If there was no matching found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	query := "UPDATE invoices SET amount = $1, vat = $2, updated_at = NOW() WHERE id = $3 RETURNING id"

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	err = m.DB.QueryRow(context.Background(), query, amount, vat, id).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

// Count records in a table
func (m InvoiceModel) CountIDs(filterQuery string) (int64, error) {
	query := fmt.Sprintf("select count(id) from invoices %s", filterQuery)
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := m.DB.QueryRow(ctx, query).Scan(&count)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Handle any errors. If there was no matching found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
	}
	return count, nil
}
