package paginator

type Paginator interface {
	GetOffset() int
	GetLimit() int
	GetPage() int
	GetSortKey() string
	GetSortDirection() string
	GetOrder() string
}

const (
	Descending string = "desc"
	Ascending  string = "asc"
)

type DefaultPaginator struct {
	Limit         int         `json:"limit,omitempty;query:limit"`
	Page          int         `json:"page,omitempty;query:page"`
	SortKey       string      `json:"sort,omitempty;query:sort"`
	SortDirection string      `json:"direction,omitempty;query:direction"`
	TotalRows     int64       `json:"total_rows"`
	TotalPages    int64       `json:"total_pages"`
	Rows          interface{} `json:"rows"`
}

func (p *DefaultPaginator) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *DefaultPaginator) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *DefaultPaginator) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *DefaultPaginator) GetOrder() string {
	if p.SortKey == "" {
		p.SortKey = "id"
	}
	return p.SortKey
}

func (p *DefaultPaginator) GetSortKey() string {
	if p.SortKey == "" {
		p.SortKey = "Id"
	}
	return p.SortKey
}

func (p *DefaultPaginator) GetSortDirection() string {
	if p.SortDirection == Ascending {
		return Ascending
	}
	return Descending
}

func (p *DefaultPaginator) SetTotalRows(totalRows int64) {
	p.TotalRows = totalRows
}

func (p *DefaultPaginator) Paginate(rows []interface{}) {
	start := p.GetOffset()
	if start < 0 {
		start = 0
	}
	end := start + p.GetLimit()

	totalRows := len(rows)
	if end > totalRows {
		end = totalRows
	}

	if start < totalRows {
		p.Rows = rows[start:end]
	}
	p.TotalRows = int64(totalRows)
	p.TotalPages = 1
	if totalRows > p.GetLimit() && totalRows > 0 {
		p.TotalPages = int64(totalRows / p.GetLimit())
	}
}
