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

// CompanyDetails type details
type CompanyDetails struct {
	INN     string `json:"inn,omitempty"`
	KPP     string `json:"kpp,omitempty"`
	OGRN    string `json:"ogrn,omitempty"`
	Address string `json:"address,omitempty"`
}

// Company type
type Company struct {
	ID           int64           `json:"id"`
	Logo         *string         `json:"logo,omitempty"`
	Name         string          `json:"name"`
	FullName     string          `json:"full_name,omitempty"`
	CompanyType  int             `json:"company_type,omitempty"`
	Details      *CompanyDetails `json:"details,omitempty"`
	UserID       *int64          `json:"user_id,omitempty"`
	DestroyedAt  *time.Time      `json:"destroyed_at,omitempty"`
	CreatedAt    *time.Time      `json:"created_at,omitempty"`
	UpdatedAt    *time.Time      `json:"updated_at,omitempty"`
	Organisation *Organisation   `json:"organisation,omitempty"`
	Contacts     []*Contact      `json:"contacts,omitempty"`
}

// CompanySearch  type
type CompanySearch struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompanyFilters struct {
	Name string
}

func ValidateCompany(v *validator.Validator, company *Company) {
	v.Check(company.Name != "", "name", "must be provided")
	v.Check(company.CompanyType != 0, "company_type", "must be provided")
}

// Define a CompanyModel struct type which wraps a pgx.Conn connection pool.
type CompanyModel struct {
	DB *pgxpool.Pool
}

func (m CompanyModel) GetAll(filters CompanyFilters, pagination Pagination) ([]*Company, Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	queryElements := []string{}
	filterQuery := ""

	if len(queryElements) > 0 {
		filterQuery = " WHERE " + strings.Join(queryElements, " AND ") + " "
	}

	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf(`
		SELECT id, logo, name, full_name, company_type, details, user_id, created_at, updated_at
		FROM companies
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

	companies := []*Company{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var company Company

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&company.ID,
			&company.Logo,
			&company.Name,
			&company.FullName,
			&company.CompanyType,
			&company.Details,
			&company.UserID,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Company struct to the slice.
		companies = append(companies, &company)
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

	return companies, metadata, nil
}

// Use for search companies
func (m CompanyModel) Search(filters CompanyFilters) ([]*CompanySearch, error) {
	// Construct the SQL query to retrieve all movie records.
	queryElements := []string{}
	filterQuery := ""
	q := ""

	if filters.Name != "" {
		q = fmt.Sprintf("(to_tsvector('simple', name) @@ plainto_tsquery('simple', '%s') OR name = '')", filters.Name)
		queryElements = append(queryElements, q)
	}

	if len(queryElements) > 0 {
		filterQuery = " WHERE " + strings.Join(queryElements, " AND ") + " "
	}

	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf("SELECT id, name FROM companies %s  ORDER BY name LIMIT 10", filterQuery)

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

	companies := []*CompanySearch{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var company CompanySearch

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&company.ID,
			&company.Name,
		)
		if err != nil {
			return nil, err
		}

		// Add the Company struct to the slice.
		companies = append(companies, &company)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return companies, nil
}

// Add method for inserting a new record in the Companys table.
func (m CompanyModel) Insert(company *Company) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO companies (
			name, full_name, company_type, details) VALUES ($1, $2, $3, $4)
		RETURNING id, name, full_name, company_type, details, created_at, updated_at`

	args := []interface{}{
		company.Name,
		company.FullName,
		company.CompanyType,
		company.Details,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&company.ID,
		&company.Name,
		&company.FullName,
		&company.CompanyType,
		&company.Details,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
}

// Add method for fetching a specific record from the companies table.
func (m CompanyModel) Get(id int64) (*Company, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no companies will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, name, full_name, company_type, details, created_at, updated_at 
		FROM companies WHERE id = $1`

	// Declare a Company struct to hold the data returned by the query.
	var company Company

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.FullName,
		&company.CompanyType,
		&company.Details,
		&company.CreatedAt,
		&company.UpdatedAt,
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

	return &company, nil
}

// Add method for updating a specific record in the companies table.
func (m CompanyModel) Update(company *Company) error {
	query := `
		UPDATE companies
		SET logo = $1, name = $2, full_name = $3, company_type = $4, details = $5, updated_at = NOW() 
		WHERE id = $6
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		company.Logo,
		company.Name,
		company.FullName,
		company.CompanyType,
		company.Details,
		company.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&company.UpdatedAt)
}

// Add method for deleting a specific record from the companies table.
func (m CompanyModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM companies WHERE id = $1`

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

	// If no rows were affected, we know that the companies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Count records in a table
func (m CompanyModel) CountIDs() (int64, error) {
	query := "select count(id) from companies"
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
