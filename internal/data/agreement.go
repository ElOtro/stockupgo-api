package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Agreement type
type Agreement struct {
	ID          int64      `json:"id"`
	StartAt     *time.Time `json:"start_at,omitempty"`
	EndAt       *time.Time `json:"end_at,omitempty"`
	Name        string     `json:"name"`
	CompanyID   int64      `json:"company_id,omitempty"`
	UserID      *int64     `json:"user_id,omitempty"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type AgreementFilters struct {
	CompanyID int64
	Start     *time.Time
	End       *time.Time
}

func ValidateAgreement(v *validator.Validator, agreement *Agreement) {
	v.Check(agreement.CompanyID != 0, "company_id", "must be provided")
	v.Check(agreement.Name != "", "name", "must be provided")
}

func ValidateFilters(v *validator.Validator, f AgreementFilters) {
	v.Check(f.CompanyID != 0, "company_id", "must be provided")
}

// Define a AgreementModel struct type which wraps a pgx.Conn connection pool.
type AgreementModel struct {
	DB *pgxpool.Pool
}

func (m AgreementModel) GetAll(filters AgreementFilters, pagination Pagination) ([]*Agreement, Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	queryElements := []string{}
	filterQuery := ""
	q := ""
	if filters.CompanyID > 0 {
		q = fmt.Sprintf("company_id = %d", filters.CompanyID)
		queryElements = append(queryElements, q)
	}

	if filters.Start != nil && filters.End != nil {
		q = fmt.Sprintf("start_at BETWEEN '%s' AND '%s'", filters.Start.Format(time.RFC3339), filters.End.Format(time.RFC3339))
		queryElements = append(queryElements, q)
	}

	if len(queryElements) > 0 {
		filterQuery = " WHERE " + strings.Join(queryElements, " AND ") + " "
	}

	query := fmt.Sprintf(`
				SELECT id, start_at, end_at, name, company_id, user_id, created_at, updated_at 
			  	FROM agreements
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

	agreements := []*Agreement{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var agreement Agreement

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&agreement.ID,
			&agreement.StartAt,
			&agreement.EndAt,
			&agreement.Name,
			&agreement.CompanyID,
			&agreement.UserID,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Agreement struct to the slice.
		agreements = append(agreements, &agreement)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	totalRecords, err := m.CountIDs()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, pagination.Page, pagination.Limit)

	return agreements, metadata, nil
}

// Add method for inserting a new record in the Agreements table.
func (m AgreementModel) Insert(agreement *Agreement) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO agreements (start_at, end_at, name, company_id, user_id) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, start_at, end_at, name, company_id, user_id, created_at, updated_at`

	args := []interface{}{
		agreement.StartAt,
		agreement.EndAt,
		agreement.Name,
		agreement.CompanyID,
		agreement.UserID,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&agreement.ID,
		&agreement.StartAt,
		&agreement.EndAt,
		&agreement.Name,
		&agreement.CompanyID,
		&agreement.UserID,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
	)
}

// Add method for fetching a specific record from the agreements table.
func (m AgreementModel) Get(id int64) (*Agreement, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no agreements will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `SELECT id, start_at, end_at, name, company_id, user_id, created_at, updated_at 
	          FROM agreements WHERE id = $1`

	// Declare a Agreement struct to hold the data returned by the query.
	var agreement Agreement

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&agreement.ID,
		&agreement.StartAt,
		&agreement.EndAt,
		&agreement.Name,
		&agreement.CompanyID,
		&agreement.UserID,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
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

	return &agreement, nil
}

// Add method for updating a specific record in the agreements table.
func (m AgreementModel) Update(agreement *Agreement) error {
	query := `
		UPDATE agreements
		SET start_at = $1, end_at = $2, name = $3, company_id = $4, user_id = $5, updated_at = NOW() 
		WHERE id = $6
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		agreement.StartAt,
		agreement.EndAt,
		agreement.Name,
		agreement.CompanyID,
		agreement.UserID,
		agreement.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&agreement.UpdatedAt)
}

// Add method for deleting a specific record from the agreements table.
func (m AgreementModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM agreements WHERE id = $1`

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

	// If no rows were affected, we know that the agreements table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Count records in a table
func (m AgreementModel) CountIDs() (int64, error) {
	query := "select count(id) from agreements"
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
