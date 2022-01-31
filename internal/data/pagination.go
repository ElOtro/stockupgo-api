package data

import (
	"math"

	"github.com/ElOtro/stockup-api/internal/validator"
)

// Define a new Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int   `json:"current_page,omitempty"`
	PageSize     int   `json:"page_size,omitempty"`
	FirstPage    int   `json:"first_page,omitempty"`
	LastPage     int   `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

type Pagination struct {
	Page              int
	Limit             int
	Sort              string
	Direction         string
	SortSafelist      []string
	DirectionSafelist []string
}

func ValidatePagination(v *validator.Validator, p Pagination) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(p.Page > 0, "page", "must be greater than zero")
	v.Check(p.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(p.Limit > 0, "limit", "must be greater than zero")
	v.Check(p.Limit <= 100, "limit", "must be a maximum of 100")
	// Check that the sort parameter matches a value in the safelist.
	v.Check(validator.In(p.Sort, p.SortSafelist...), "sort", "invalid sort value")
	v.Check(validator.In(p.Direction, p.DirectionSafelist...), "direction", "invalid direction value")
}

// Check that the client-provided Sort field matches one of the entries in our safelist
// and if it does, extract the column name from the Sort field by stripping the leading
// hyphen character (if one exists).
func (p Pagination) sortColumn() string {
	for _, safeValue := range p.SortSafelist {
		if p.Sort == safeValue {
			return p.Sort
		}
	}

	panic("unsafe sort parameter: " + p.Sort)
}

// Return the sort direction ("ASC" or "DESC") depending on the prefix character of the
// Sort field.
func (p Pagination) sortDirection() string {
	for _, safeValue := range p.DirectionSafelist {
		if p.Direction == safeValue {
			return p.Direction
		}
	}

	return "ASC"
}

func (p Pagination) limit() int {
	return p.Limit
}

func (p Pagination) offset() int {
	return (p.Page - 1) * p.Limit
}

// The calculateMetadata() function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that the last page value is calculated using the math.Ceil() function, which rounds
// up a float to the nearest integer. So, for example, if there were 12 records in total
// and a page size of 5, the last page value would be math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords int64, page, limit int) Metadata {
	if totalRecords == 0 {
		// Note that we return an empty Metadata struct if there are no records.
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     limit,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(limit))),
		TotalRecords: totalRecords,
	}
}
