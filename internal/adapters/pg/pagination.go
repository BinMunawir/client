package pg

// Pagination is the generic page envelope reused by every store (service standard §7).
// The COUNT scalar is scanned straight into Total via the "total" column alias.
type Pagination[T any] struct {
	Items  []T
	Total  int64
	Limit  int
	Offset int
}

func (p Pagination[T]) Pages() int64 {
	if p.Limit <= 0 {
		return 0
	}
	return (p.Total + int64(p.Limit) - 1) / int64(p.Limit)
}

func (p Pagination[T]) HasNext() bool {
	return int64(p.Offset)+int64(len(p.Items)) < p.Total
}

const (
	defaultLimit = 50
	maxLimit     = 200
)

func (p Pagination[T]) NormalizePaging(limit, offset int) (int, int) {
	switch {
	case limit <= 0:
		limit = defaultLimit
	case limit > maxLimit:
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
