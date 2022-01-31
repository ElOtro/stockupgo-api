package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// VatRate struct
type VatRate struct {
	ID          int64      `json:"id"`
	IsActive    bool       `json:"is_active,omitempty"`
	IsDefault   bool       `json:"is_default,omitempty"`
	Rate        float64    `json:"rate"`
	Name        string     `json:"name"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func ValidateVatRate(v *validator.Validator, vatRate *VatRate) {
	v.Check(vatRate.Rate >= 0, "rate", "must be provided")
	v.Check(vatRate.Name != "", "name", "must be provided")
}

// Define a VatRateModel struct type which wraps a pgx.Conn connection pool.
type VatRateModel struct {
	DB *pgxpool.Pool
}

func (m VatRateModel) GetAll() ([]*VatRate, error) {
	// Construct the SQL query to retrieve all movie records.
	query := "SELECT id, is_active, is_default, rate, name, created_at, updated_at FROM vat_rates"

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

	vatRates := []*VatRate{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var vatRate VatRate

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&vatRate.ID,
			&vatRate.IsActive,
			&vatRate.IsDefault,
			&vatRate.Rate,
			&vatRate.Name,
			&vatRate.CreatedAt,
			&vatRate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the VatRate struct to the slice.
		vatRates = append(vatRates, &vatRate)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vatRates, nil
}

// Add method for inserting a new record in the VatRates table.
func (m VatRateModel) Insert(vatRate *VatRate) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO vat_rates (is_active, is_default, rate, name) VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, is_default, rate, name, created_at, updated_at`

	args := []interface{}{
		vatRate.IsActive,
		vatRate.IsDefault,
		vatRate.Rate,
		vatRate.Name,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&vatRate.ID,
		&vatRate.IsActive,
		&vatRate.IsDefault,
		&vatRate.Rate,
		&vatRate.Name,
		&vatRate.CreatedAt,
		&vatRate.UpdatedAt,
	)
}

// Add method for fetching a specific record from the vatRates table.
func (m VatRateModel) Get(id int64) (*VatRate, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no vatRates will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `SELECT id, is_active, is_default, rate, name, created_at, updated_at 
	          FROM vat_rates WHERE id = $1`

	// Declare a VatRate struct to hold the data returned by the query.
	var vatRate VatRate

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&vatRate.ID,
		&vatRate.IsActive,
		&vatRate.IsDefault,
		&vatRate.Rate,
		&vatRate.Name,
		&vatRate.CreatedAt,
		&vatRate.UpdatedAt,
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

	return &vatRate, nil
}

// Add method for updating a specific record in the vat_rates table.
func (m VatRateModel) Update(vatRate *VatRate) error {
	query := `
		UPDATE vat_rates
		SET is_active = $1, is_default = $2, rate = $3, name = $4, updated_at = NOW() 
		WHERE id = $5
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		vatRate.IsActive,
		vatRate.IsDefault,
		vatRate.Rate,
		vatRate.Name,
		vatRate.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&vatRate.UpdatedAt)
}

// Add method for deleting a specific record from the vatRates table.
func (m VatRateModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM vat_rates WHERE id = $1`

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

	// If no rows were affected, we know that the vatRates table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
