package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Product struct
type Product struct {
	ID          int64      `json:"id"`
	IsActive    bool       `json:"is_active,omitempty"`
	ProductType int        `json:"product_type,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	SKU         string     `json:"sku,omitempty"`
	Price       float64    `json:"price,omitempty"`
	VatRateID   *int64     `json:"vat_rate_id,omitempty"`
	VatRate     *VatRate   `json:"vat_rate,omitempty"`
	UnitID      *int64     `json:"unit_id,omitempty"`
	Unit        *Unit      `json:"unit,omitempty"`
	UserID      *int64     `json:"user_id,omitempty"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.Name != "", "name", "must be provided")
}

// Define a ProductModel struct type which wraps a pgx.Conn connection pool.
type ProductModel struct {
	DB *pgxpool.Pool
}

func (m ProductModel) GetAll() ([]*Product, error) {
	// Construct the SQL query to retrieve all movie records.
	query := `SELECT id, 
	                 is_active, product_type, name, description, 
	                 sku, price, 
					 (SELECT row_to_json(row)
		 			 FROM
		              (SELECT id, rate, name
		               FROM vat_rates
		               WHERE vat_rates.id = vat_rate_id) row) AS vat_rate,
					 (SELECT row_to_json(row)
		 			 FROM
		              (SELECT id, code, name
		               FROM units
		               WHERE units.id = unit_id) row) AS unit, 
					 user_id, created_at, updated_at 
			 FROM products 
			 WHERE destroyed_at IS NULL`

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

	products := []*Product{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var product Product

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&product.ID,
			&product.IsActive,
			&product.ProductType,
			&product.Name,
			&product.Description,
			&product.SKU,
			&product.Price,
			&product.VatRate,
			&product.Unit,
			&product.UserID,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Product struct to the slice.
		products = append(products, &product)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// Add method for inserting a new record in the Products table.
func (m ProductModel) Insert(product *Product) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO products (is_active, product_type, name, description, 
			sku, price, vat_rate_id, unit_id, user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, is_active, product_type, name, description, sku, price, 
		          vat_rate_id, unit_id, user_id, created_at, updated_at`

	args := []interface{}{
		product.IsActive,
		product.ProductType,
		product.Name,
		product.Description,
		product.SKU,
		product.Price,
		product.VatRateID,
		product.UnitID,
		product.UserID,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&product.ID,
		&product.IsActive,
		&product.ProductType,
		&product.Name,
		&product.Description,
		&product.SKU,
		&product.Price,
		&product.VatRateID,
		&product.UnitID,
		&product.UserID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
}

// Add method for fetching a specific record from the products table.
func (m ProductModel) Get(id int64) (*Product, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no products will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, is_active, product_type, name, description, sku, price, 
	       (SELECT row_to_json(row) FROM (SELECT id, rate, name FROM vat_rates WHERE vat_rates.id = vat_rate_id) row) AS vat_rate,
		   (SELECT row_to_json(row) FROM (SELECT id, code, name FROM units WHERE units.id = unit_id) row) AS unit,  
		   user_id, created_at, updated_at 
		FROM products WHERE id = $1`

	// Declare a Product struct to hold the data returned by the query.
	var product Product

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.IsActive,
		&product.ProductType,
		&product.Name,
		&product.Description,
		&product.SKU,
		&product.Price,
		&product.VatRate,
		&product.Unit,
		&product.UserID,
		&product.CreatedAt,
		&product.UpdatedAt,
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

	return &product, nil
}

// Add method for updating a specific record in the products table.
func (m ProductModel) Update(product *Product) error {
	query := `
		UPDATE products
		SET is_active = $1, product_type = $2, name = $3, description = $4, sku = $5, 
		price = $6, vat_rate_id = $7, unit_id = $8, updated_at = NOW() 
		WHERE id = $9
		RETURNING updated_at`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		product.IsActive,
		product.ProductType,
		product.Name,
		product.Description,
		product.SKU,
		product.Price,
		product.VatRateID,
		product.UnitID,
		product.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(&product.UpdatedAt)
}

// Add method for deleting a specific record from the products table.
func (m ProductModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM products WHERE id = $1`

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

	// If no rows were affected, we know that the products table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
