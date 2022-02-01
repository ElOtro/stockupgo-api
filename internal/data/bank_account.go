package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// OrganisationDetails type details
type BankAccountDetails struct {
	BIK         string `json:"bik"`
	Account     string `json:"account"`
	INN         string `json:"inn"`
	KPP         string `json:"kpp"`
	CorrAccount string `json:"corr_account"`
}

// BankAccount type details
type BankAccount struct {
	ID             int64               `json:"id"`
	OrganisationID int64               `json:"organisation_id,omitempty"`
	IsDefault      bool                `json:"is_default,omitempty"`
	Name           string              `json:"name"`
	Details        *BankAccountDetails `json:"details,omitempty"`
	DestroyedAt    *time.Time          `json:"destroyed_at,omitempty"`
	CreatedAt      *time.Time          `json:"created_at,omitempty"`
	UpdatedAt      *time.Time          `json:"updated_at,omitempty"`
}

func ValidateBankAccount(v *validator.Validator, bankAccount *BankAccount) {
	v.Check(bankAccount.Name != "", "bank_accounts name", "must be provided")
}

// Define a BankAccount struct type which wraps a pgx.Conn connection pool.
type BankAccountModel struct {
	DB *pgxpool.Pool
}

func (m BankAccountModel) GetAll(organisationID int64) ([]*BankAccount, error) {
	// Construct the SQL query to retrieve all movie records.
	query := `
		SELECT id, is_default, name, details, created_at, updated_at 
		FROM bank_accounts 
		WHERE organisation_id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.DB.Query(ctx, query, organisationID)
	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	bankAccounts := []*BankAccount{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var bankAccount BankAccount

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&bankAccount.ID,
			&bankAccount.IsDefault,
			&bankAccount.Name,
			&bankAccount.Details,
			&bankAccount.CreatedAt,
			&bankAccount.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Organisation struct to the slice.
		bankAccounts = append(bankAccounts, &bankAccount)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bankAccounts, nil
}

// Add method for inserting a new record in the Organisations table.
func (m BankAccountModel) Insert(organisationID int64, bankAccount *BankAccount) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO bank_accounts (organisation_id, name, is_default, details) VALUES ($1, $2, $3, $4)
		RETURNING id, name, is_default, details, created_at, updated_at`

	args := []interface{}{
		organisationID,
		bankAccount.Name,
		bankAccount.IsDefault,
		bankAccount.Details,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&bankAccount.ID,
		&bankAccount.Name,
		&bankAccount.IsDefault,
		&bankAccount.Details,
		&bankAccount.CreatedAt,
		&bankAccount.UpdatedAt,
	)
}

// Add method for fetching a specific record from the organisations table.
func (m BankAccountModel) Get(organisationID int64, id int64) (*BankAccount, error) {

	if id < 1 || organisationID < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, is_default, name, details, created_at, updated_at 
		FROM bank_accounts 
		WHERE organisation_id = $1 AND id = $2`

	args := []interface{}{organisationID, id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Declare a BankAccount struct to hold the data returned by the query.
	var bankAccount BankAccount

	// Execute the query using the QueryRow() method
	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&bankAccount.ID,
		&bankAccount.IsDefault,
		&bankAccount.Name,
		&bankAccount.Details,
		&bankAccount.CreatedAt,
		&bankAccount.UpdatedAt,
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

	return &bankAccount, nil
}

// Add method for updating a specific record in the organisations table.
func (m BankAccountModel) Update(bankAccount *BankAccount) error {
	query := `
		UPDATE bank_accounts
		SET name = $1, is_default = $2, details = $3, updated_at = NOW() 
		WHERE id = $4
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		bankAccount.Name,
		bankAccount.IsDefault,
		bankAccount.Details,
		bankAccount.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&bankAccount.UpdatedAt)
}

// Add method for deleting a specific record from the organisations table.
func (m BankAccountModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM bank_accounts WHERE id = $1`

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
