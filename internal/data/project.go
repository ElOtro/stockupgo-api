package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Project type
type Project struct {
	ID             int64      `json:"id"`
	OrganisationID int64      `json:"organisation_id"`
	Name           string     `json:"name"`
	DestroyedAt    *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

func ValidateProject(v *validator.Validator, project *Project) {
	v.Check(project.OrganisationID != 0, "organisation_id", "must be provided")
	v.Check(project.Name != "", "name", "must be provided")
}

// Define a ProjectModel struct type which wraps a pgx.Conn connection pool.
type ProjectModel struct {
	DB *pgxpool.Pool
}

func (m ProjectModel) GetAll() ([]*Project, error) {
	// Construct the SQL query to retrieve all movie records.
	query := "SELECT id, organisation_id, name, created_at, updated_at FROM projects"

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

	projects := []*Project{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var project Project

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&project.ID,
			&project.OrganisationID,
			&project.Name,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Project struct to the slice.
		projects = append(projects, &project)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// Add method for inserting a new record in the Projects table.
func (m ProjectModel) Insert(project *Project) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO projects (organisation_id, name) VALUES ($1, $2)
		RETURNING id, organisation_id, name, created_at, updated_at`

	args := []interface{}{
		project.OrganisationID,
		project.Name,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&project.ID,
		&project.OrganisationID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
}

// Add method for fetching a specific record from the projects table.
func (m ProjectModel) Get(id int64) (*Project, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no projects will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := "SELECT id, organisation_id, name, created_at, updated_at FROM projects WHERE id = $1"

	// Declare a Project struct to hold the data returned by the query.
	var project Project

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.OrganisationID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
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

	return &project, nil
}

// Add method for updating a specific record in the projects table.
func (m ProjectModel) Update(project *Project) error {
	query := `
		UPDATE projects
		SET organisation_id = $1, name = $2, updated_at = NOW() 
		WHERE id = $3
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		project.OrganisationID,
		project.Name,
		project.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&project.UpdatedAt)
}

// Add method for deleting a specific record from the projects table.
func (m ProjectModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM projects WHERE id = $1`

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

	// If no rows were affected, we know that the projects table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
