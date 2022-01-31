package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Unit type
type Unit struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func ValidateUnit(v *validator.Validator, unit *Unit) {
	v.Check(unit.Code != "", "code", "must be provided")
	v.Check(unit.Name != "", "name", "must be provided")
}

// Define a UnitModel struct type which wraps a pgx.Conn connection pool.
type UnitModel struct {
	DB *pgxpool.Pool
}

func (m UnitModel) GetAll() ([]*Unit, error) {
	// Construct the SQL query to retrieve all movie records.
	query := "SELECT id, code, name, created_at, updated_at FROM units"

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

	units := []*Unit{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var unit Unit

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&unit.ID,
			&unit.Code,
			&unit.Name,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Unit struct to the slice.
		units = append(units, &unit)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return units, nil
}

// Add method for inserting a new record in the Units table.
func (m UnitModel) Insert(unit *Unit) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO units (code, name) VALUES ($1, $2)
		RETURNING id, code, name, created_at, updated_at`

	args := []interface{}{
		unit.Code,
		unit.Name,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&unit.ID,
		&unit.Code,
		&unit.Name,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)
}

// Add method for fetching a specific record from the units table.
func (m UnitModel) Get(id int64) (*Unit, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no units will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := "SELECT id, code, name, created_at, updated_at FROM units WHERE id = $1"

	// Declare a Unit struct to hold the data returned by the query.
	var unit Unit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&unit.ID,
		&unit.Code,
		&unit.Name,
		&unit.CreatedAt,
		&unit.UpdatedAt,
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

	return &unit, nil
}

// Add method for updating a specific record in the units table.
func (m UnitModel) Update(unit *Unit) error {
	query := `
		UPDATE units
		SET code = $1, name = $2, updated_at = NOW() 
		WHERE id = $3
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		unit.Code,
		unit.Name,
		unit.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&unit.UpdatedAt)
}

// Add method for deleting a specific record from the units table.
func (m UnitModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM units WHERE id = $1`

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

	// If no rows were affected, we know that the units table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
