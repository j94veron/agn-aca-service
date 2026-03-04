package domain

type OliquiFilter struct {
	UniNego string // obligatorio
	Schema  string // resuelto internamente

	ContInterno *int64
	IdAGN       *string
	Cosecha     *string
	Destino     *string
	Zona        *string
	Comprador   *string
	TipoEntrega *string

	FechaDesde *string
	FechaHasta *string

	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}
