package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Contact type details
type Contact struct {
	ID          int64      `json:"id"`
	Role        int        `json:"role"`
	Title       string     `json:"title"`
	Name        string     `json:"name"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	StartAt     *time.Time `json:"start_at"`
	Sign        *string    `json:"sign,omitempty"`
	CompanyID   int64      `json:"company_id,omitempty"`
	UserID      int64      `json:"user_id,omitempty"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func ValidateContact(v *validator.Validator, contact *Contact) {
	v.Check(contact.Role != 0, "role", "must be provided")
	v.Check(contact.Name != "", "name", "must be provided")
	v.Check(!contact.StartAt.IsZero(), "start_at", "must be provided")
}

// Define a ContactModel struct type which wraps a pgx.Conn connection pool.
type ContactModel struct {
	DB *pgxpool.Pool
}

func (m ContactModel) GetAll(companyID int64) ([]*Contact, error) {
	// Construct the SQL query to retrieve all movie records.
	query := `
		SELECT id, role, title, name, phone, email, start_at, created_at, updated_at 
		FROM contacts 
		WHERE company_id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.DB.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	contacts := []*Contact{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var contact Contact

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&contact.ID,
			&contact.Role,
			&contact.Title,
			&contact.Name,
			&contact.Phone,
			&contact.Email,
			&contact.StartAt,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Organisation struct to the slice.
		contacts = append(contacts, &contact)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}

// Add method for inserting a new record in the contacts table.
func (m ContactModel) Insert(companyID int64, contact *Contact) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO contacts (company_id, role, title, name, phone, email, start_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, role, title, name, phone, email, start_at, created_at, updated_at`

	args := []interface{}{
		companyID,
		contact.Role,
		contact.Title,
		contact.Name,
		contact.Phone,
		contact.Email,
		contact.StartAt,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&contact.ID,
		&contact.Role,
		&contact.Title,
		&contact.Name,
		&contact.Phone,
		&contact.Email,
		&contact.StartAt,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)
}

// Add method for fetching a specific record from the organisations table.
func (m ContactModel) Get(companyID int64, id int64) (*Contact, error) {

	if id < 1 || companyID < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, role, title, name, phone, email, start_at, created_at, updated_at 
		FROM contacts 
		WHERE company_id = $1 AND id = $2`

	args := []interface{}{companyID, id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Declare a Contact struct to hold the data returned by the query.
	var contact Contact

	// Execute the query using the QueryRow() method
	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&contact.ID,
		&contact.Role,
		&contact.Title,
		&contact.Name,
		&contact.Phone,
		&contact.Email,
		&contact.StartAt,
		&contact.CreatedAt,
		&contact.UpdatedAt,
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

	return &contact, nil
}

// Add method for updating a specific record in the organisations table.
// role, title, name, phone, email, start_at
func (m ContactModel) Update(contact *Contact) error {
	query := `
		UPDATE contacts
		SET role = $1, title = $2, name = $3, phone = $4, email = $5, start_at = $6, updated_at = NOW() 
		WHERE id = $7
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		contact.Role,
		contact.Title,
		contact.Name,
		contact.Phone,
		contact.Email,
		contact.StartAt,
		contact.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&contact.UpdatedAt)
}

// Add method for deleting a specific record from the organisations table.
func (m ContactModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM contacts WHERE id = $1`

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
