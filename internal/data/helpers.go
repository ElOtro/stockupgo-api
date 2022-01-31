package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Define a ContactModel struct type which wraps a pgx.Conn connection pool.
type Helper struct {
	DB *pgxpool.Pool
}

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func (h Helper) pluckIDs(table string) ([]int64, error) {
	query := fmt.Sprintf("select id from %s", table)
	var ids []int64
	rows, _ := h.DB.Query(context.Background(), query)
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}
	return ids, nil
}
