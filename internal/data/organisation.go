package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// OrganisationDetails type details
type OrganisationDetails struct {
	INN     string `json:"inn,omitempty"`
	KPP     string `json:"kpp,omitempty"`
	OGRN    string `json:"ogrn,omitempty"`
	Address string `json:"address,omitempty"`
}

// Organisation type details
type Organisation struct {
	ID                 int64                `json:"id"`
	Name               string               `json:"name"`
	FullName           string               `json:"full_name,omitempty"`
	CEO                string               `json:"ceo,omitempty"`
	CEOTitle           string               `json:"ceo_title,omitempty"`
	CFO                string               `json:"cfo,omitempty"`
	CFOTitle           string               `json:"cfo_title,omitempty"`
	Stamp              *string              `json:"stamp,omitempty"`
	CEOSign            *string              `json:"ceo_sign,omitempty"`
	CFOSign            *string              `json:"cfo_sign,omitempty"`
	IsVatPayer         bool                 `json:"is_vat_payer,omitempty"`
	Details            *OrganisationDetails `json:"details,omitempty"`
	UUID               string               `json:"uuid,omitempty"`
	DestroyedAt        *time.Time           `json:"destroyed_at,omitempty"`
	CreatedAt          *time.Time           `json:"created_at,omitempty"`
	UpdatedAt          *time.Time           `json:"updated_at,omitempty"`
	DefaultBankAccount *BankAccount         `json:"default_bank_account,omitempty"`
	BankAccounts       []*BankAccount       `json:"bank_accounts,omitempty"`
}

func ValidateOrganisation(v *validator.Validator, organisation *Organisation) {
	v.Check(organisation.Name != "", "name", "must be provided")
	v.Check(organisation.FullName != "", "full_name", "must be provided")
}

// Define a OrganisationModel struct type which wraps a pgx.Conn connection pool.
type OrganisationModel struct {
	DB *pgxpool.Pool
}

func (m OrganisationModel) GetAll() ([]*Organisation, error) {
	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf(`
		SELECT id, name, full_name, ceo, ceo_title, cfo, cfo_title, stamp, ceo_sign, cfo_sign, is_vat_payer, 
		details, created_at, updated_at 
		FROM organisations`)

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	organisations := []*Organisation{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var organisation Organisation

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&organisation.ID,
			&organisation.Name,
			&organisation.FullName,
			&organisation.CEO,
			&organisation.CEOTitle,
			&organisation.CFO,
			&organisation.CFOTitle,
			&organisation.Stamp,
			&organisation.CEOSign,
			&organisation.CFOSign,
			&organisation.IsVatPayer,
			&organisation.Details,
			&organisation.CreatedAt,
			&organisation.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Organisation struct to the slice.
		organisations = append(organisations, &organisation)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return organisations, nil
}

// Add method for inserting a new record in the Organisations table.
func (m OrganisationModel) Insert(organisation *Organisation) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO organisations (
			name, full_name, ceo, ceo_title, cfo, cfo_title, stamp, ceo_sign, cfo_sign, is_vat_payer, 
			details) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, full_name, ceo, ceo_title, cfo, cfo_title, stamp, ceo_sign, cfo_sign, is_vat_payer, 
		          details, uuid, created_at, updated_at`

	args := []interface{}{
		organisation.Name,
		organisation.FullName,
		organisation.CEO,
		organisation.CEOTitle,
		organisation.CFO,
		organisation.CFOTitle,
		organisation.Stamp,
		organisation.CEOSign,
		organisation.CFOSign,
		organisation.IsVatPayer,
		organisation.Details,
	}

	// fmt.Println(args)

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&organisation.ID, &organisation.Name,
		&organisation.FullName, &organisation.CEO, &organisation.CEOTitle, &organisation.CFO,
		&organisation.CFOTitle, &organisation.Stamp, &organisation.CEOSign, &organisation.CFOSign,
		&organisation.IsVatPayer, &organisation.Details, &organisation.UUID, &organisation.CreatedAt,
		&organisation.UpdatedAt,
	)
}

// Add method for fetching a specific record from the organisations table.
func (m OrganisationModel) Get(id int64) (*Organisation, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no organisations will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, name, full_name, ceo, ceo_title, cfo, cfo_title, stamp, ceo_sign, cfo_sign, is_vat_payer, 
		details, uuid, created_at, updated_at, 
		(SELECT row_to_json(oba)
		 FROM
		 (SELECT id, name
		  FROM bank_accounts
		  WHERE organisation_id = $1 AND bank_accounts.is_default = true) oba) AS default_bank_account 
		FROM organisations WHERE id = $1`

	// Declare a Organisation struct to hold the data returned by the query.
	var organisation Organisation

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&organisation.ID,
		&organisation.Name,
		&organisation.FullName,
		&organisation.CEO,
		&organisation.CEOTitle,
		&organisation.CFO,
		&organisation.CFOTitle,
		&organisation.Stamp,
		&organisation.CEOSign,
		&organisation.CFOSign,
		&organisation.IsVatPayer,
		&organisation.Details,
		&organisation.UUID,
		&organisation.CreatedAt,
		&organisation.UpdatedAt,
		&organisation.DefaultBankAccount,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
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

	return &organisation, nil
}

// Add method for updating a specific record in the organisations table.
func (m OrganisationModel) Update(organisation *Organisation) error {
	query := `
		UPDATE organisations
		SET name = $1, full_name = $2, ceo = $3, ceo_title = $4, cfo = $5, cfo_title = $6,
		stamp = $7, ceo_sign = $8, cfo_sign = $9, is_vat_payer = $10, details = $11, updated_at =  NOW() 
		WHERE id = $12
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		organisation.Name,
		organisation.FullName,
		organisation.CEO,
		organisation.CEOTitle,
		organisation.CFO,
		organisation.CFOTitle,
		organisation.Stamp,
		organisation.CEOSign,
		organisation.CFOSign,
		organisation.IsVatPayer,
		organisation.Details,
		organisation.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&organisation.UpdatedAt)
}

// Add method for deleting a specific record from the organisations table.
func (m OrganisationModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM organisations WHERE id = $1`

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

	// If no rows were affected, we know that the organisations table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
