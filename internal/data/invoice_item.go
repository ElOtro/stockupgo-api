package data

import (
	"context"
	"errors"
	"time"

	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// InvoiceItem struct
type InvoiceItem struct {
	ID           int64      `json:"id"`
	InvoiceID    int64      `json:"invoice_id,omitempty"`
	Position     int        `json:"position"`
	ProductID    int64      `json:"product_id,omitempty"`
	Description  string     `json:"description"`
	UnitID       int64      `json:"unit_id,omitempty"`
	Quantity     float64    `json:"quantity"`
	Price        float64    `json:"price"`
	Amount       float64    `json:"amount"`
	DiscountRate int        `json:"discount_rate"`
	Discount     float64    `json:"discount"`
	VatRateID    int64      `json:"vat_rate_id,omitempty"`
	Vat          float64    `json:"vat,omitempty"`
	Product      *Product   `json:"product"`
	Unit         *Unit      `json:"unit"`
	VatRate      *VatRate   `json:"vat_rate"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func ValidateInvoiceItem(v *validator.Validator, invoice *InvoiceItem) {
	// v.Check(invoice.InvoiceID != 0, "invoice_id", "must be provided")
	v.Check(invoice.ProductID != 0, "product_id", "must be provided")
}

// Define a InvoiceItemModel struct type which wraps a pgx.Conn connection pool.
type InvoiceItemModel struct {
	DB *pgxpool.Pool
}

func (m InvoiceItemModel) GetAll(invoiceID int64) ([]*InvoiceItem, error) {
	// Construct the SQL query to retrieve all movie records.
	query := `
		SELECT id, position, 
		(SELECT row_to_json(row)
				FROM
				(SELECT id, name
				FROM products
				WHERE products.id = product_id) row) AS product, 
		description, 
		(SELECT row_to_json(row)
				FROM
				(SELECT id, code, name
				FROM units
				WHERE units.id = unit_id) row) AS unit, 
		quantity, price, amount, discount_rate, discount,
		(SELECT row_to_json(row)
				FROM
				(SELECT id, name
				FROM vat_rates
				WHERE vat_rates.id = vat_rate_id) row) AS vat_rate, 
		created_at, updated_at 
		FROM invoice_items 
		WHERE invoice_id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.DB.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	invoiceItems := []*InvoiceItem{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var invoiceItem InvoiceItem

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&invoiceItem.ID,
			&invoiceItem.Position,
			&invoiceItem.Product,
			&invoiceItem.Description,
			&invoiceItem.Unit,
			&invoiceItem.Quantity,
			&invoiceItem.Price,
			&invoiceItem.Amount,
			&invoiceItem.DiscountRate,
			&invoiceItem.Discount,
			&invoiceItem.VatRate,
			&invoiceItem.CreatedAt,
			&invoiceItem.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the Organisation struct to the slice.
		invoiceItems = append(invoiceItems, &invoiceItem)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return invoiceItems, nil
}

// Add method for inserting a new record in the Organisations table.
func (m InvoiceItemModel) Insert(invoiceID int64, invoiceItem *InvoiceItem) error {
	// Define the SQL query for inserting a new record
	query := `
		INSERT INTO invoice_items (
			invoice_id, position, product_id, description, unit_id, quantity, price, 
			amount, discount_rate, discount, vat_rate_id, vat
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id,
		          (SELECT row_to_json(row) FROM (SELECT id, name FROM products WHERE products.id = product_id) row) AS product,
				  (SELECT row_to_json(row) FROM (SELECT id, code, name FROM units WHERE units.id = unit_id) row) AS unit,
				  (SELECT row_to_json(row) FROM (SELECT id, name FROM vat_rates WHERE vat_rates.id = vat_rate_id) row) AS vat_rate, 
				  created_at, updated_at`

	args := []interface{}{
		invoiceID,
		invoiceItem.Position,
		invoiceItem.ProductID,
		invoiceItem.Description,
		invoiceItem.UnitID,
		invoiceItem.Quantity,
		invoiceItem.Price,
		invoiceItem.Amount,
		invoiceItem.DiscountRate,
		invoiceItem.Discount,
		invoiceItem.VatRateID,
		invoiceItem.Vat,
	}

	// Use the QueryRow() method to execute the SQL query on our connection pool
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&invoiceItem.ID,
		&invoiceItem.Product,
		&invoiceItem.Unit,
		&invoiceItem.VatRate,
		&invoiceItem.CreatedAt,
		&invoiceItem.UpdatedAt,
	)
}

// Add method for fetching a specific record from the organisations table.
func (m InvoiceItemModel) Get(invoiceID int64, id int64) (*InvoiceItem, error) {

	if id < 1 || invoiceID < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving data.
	query := `
		SELECT id, position, product_id,
		(SELECT row_to_json(row)
				FROM
				(SELECT id, name
				FROM products
				WHERE products.id = product_id) row) AS product, 
		description, unit_id,
		(SELECT row_to_json(row)
				FROM
				(SELECT id, code, name
				FROM units
				WHERE units.id = unit_id) row) AS unit, 
		quantity, price, amount, discount_rate, discount, vat_rate_id,
		(SELECT row_to_json(row)
				FROM
				(SELECT id, name
				FROM vat_rates
				WHERE vat_rates.id = vat_rate_id) row) AS vat_rate, 
		created_at, updated_at 
		FROM invoice_items 
		WHERE invoice_id = $1 AND id = $2`

	args := []interface{}{invoiceID, id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Declare a InvoiceItem struct to hold the data returned by the query.
	var invoiceItem InvoiceItem

	// Execute the query using the QueryRow() method
	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&invoiceItem.ID,
		&invoiceItem.Position,
		&invoiceItem.ProductID,
		&invoiceItem.Product,
		&invoiceItem.Description,
		&invoiceItem.UnitID,
		&invoiceItem.Unit,
		&invoiceItem.Quantity,
		&invoiceItem.Price,
		&invoiceItem.Amount,
		&invoiceItem.DiscountRate,
		&invoiceItem.Discount,
		&invoiceItem.VatRateID,
		&invoiceItem.VatRate,
		&invoiceItem.CreatedAt,
		&invoiceItem.UpdatedAt,
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

	return &invoiceItem, nil
}

// Add method for updating a specific record in the organisations table.
func (m InvoiceItemModel) Update(invoiceItem *InvoiceItem) error {
	query := `
		UPDATE invoice_items
		SET position = $1, product_id = $2, description = $3, unit_id = $4, 
		    quantity = $5, price = $6, amount = $7, discount_rate = $8, discount = $9, 
			vat_rate_id = $10, vat = $11, updated_at = NOW() 
		WHERE id = $12
		RETURNING updated_at, 
		          (SELECT row_to_json(row) FROM (SELECT id, name FROM products WHERE products.id = product_id) row) AS product,
				  (SELECT row_to_json(row) FROM (SELECT id, code, name FROM units WHERE units.id = unit_id) row) AS unit,
				  (SELECT row_to_json(row) FROM (SELECT id, name FROM vat_rates WHERE vat_rates.id = vat_rate_id) row) AS vat_rate`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		invoiceItem.Position,
		invoiceItem.ProductID,
		invoiceItem.Description,
		invoiceItem.UnitID,
		invoiceItem.Quantity,
		invoiceItem.Price,
		invoiceItem.Amount,
		invoiceItem.DiscountRate,
		invoiceItem.Discount,
		invoiceItem.VatRateID,
		invoiceItem.Vat,
		invoiceItem.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	return m.DB.QueryRow(context.Background(), query, args...).Scan(
		&invoiceItem.UpdatedAt,
		&invoiceItem.Product,
		&invoiceItem.Unit,
		&invoiceItem.VatRate,
	)
}

// Add method for deleting a specific record from the organisations table.
func (m InvoiceItemModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM invoice_items WHERE id = $1`

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
