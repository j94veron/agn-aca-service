package domain

import "time"

// Detalle (por contrato)
type PendingFixDetailRow struct {
	UniNego string `json:"uninego"` // CO / AC

	CUIT string `json:"cuit"`

	CompCta  string `json:"compcta"`
	Contrato string `json:"contrato"`

	// Vencimiento de fijación (fecfijahasta)
	FecHasta *time.Time `json:"fechasta"`
	// Vencimiento de contrato (fecvtoent)
	FecVtoEnt *time.Time `json:"fecVtoEnt"`

	// TONELADAS
	Pendientes float64 `json:"pendientes"`
}

type PendingFixDetailSnapshot struct {
	GeneratedAt time.Time       `json:"generatedAt"`
	FromDate    time.Time       `json:"fromDate"`
	ToDate      time.Time       `json:"toDate"`
	Months      int             `json:"months"`
	Rows        []PendingFixRow `json:"rows"`
}

// Resumen por CUIT (toneladas)
type PendingFixSummaryRow struct {
	UniNego string `json:"uninego"`
	CUIT    string `json:"cuit"`

	TnPendientes float64 `json:"tnPendientes"`
	TnVencidas   float64 `json:"tnVencidas"`
	CtSinVtoEnt  int     `json:"ctSinVtoEnt"` // fecvtoent IS NULL (conteo)
}

type PendingFixSummarySnapshot struct {
	GeneratedAt time.Time              `json:"generatedAt"`
	Rows        []PendingFixSummaryRow `json:"rows"`
}

// 12 meses por toneladas (siempre 12, con ceros)
type PendingFixMonthlyRow struct {
	Month string  `json:"month"` // YYYY-MM
	Tn    float64 `json:"tn"`
}

type PendingFixMonthlySnapshot struct {
	GeneratedAt time.Time              `json:"generatedAt"`
	FromMonth   string                 `json:"fromMonth"`
	Months      int                    `json:"months"`
	Rows        []PendingFixMonthlyRow `json:"rows"`
}

// Vencidos con vacíos (ventana corta, por CUIT)
type PendingFixVencidosMonth struct {
	Month string  `json:"month"`
	Tn    float64 `json:"tn"`
}

type PendingFixVencidosRow struct {
	UniNego     string                    `json:"uninego"`
	CUIT        string                    `json:"cuit"`
	AnchorMonth string                    `json:"anchorMonth"`
	Months      []PendingFixVencidosMonth `json:"months"` // siempre N meses (ej 3)
}

type PendingFixVencidosSnapshot struct {
	GeneratedAt  time.Time               `json:"generatedAt"`
	WindowMonths int                     `json:"windowMonths"`
	Rows         []PendingFixVencidosRow `json:"rows"`
}

type PendingFixRow struct {
	UniNego string `json:"uninego"` // "CO" o "AC"

	CUIT string `json:"cuit"` // NECESARIO para summaries

	VendCta         string `json:"vendcta"`
	VendNombre      string `json:"vendnombre"`
	CompCta         string `json:"compcta"`
	CompNombre      string `json:"compnombre"`
	VendedorNombre  string `json:"vendedorNombre"`
	CompradorNombre string `json:"compradorNombre"`

	Contrato     string  `json:"contrato"`
	TTMaxPact    float64 `json:"ttmaxpact"`
	IttAplicadas float64 `json:"ittaplicadas"`
	IttFijadas   float64 `json:"ittfijadas"`
	ContVendedor string  `json:"contVendedor"`

	FecEnt    *time.Time `json:"fecEnt"`
	FecVtoEnt *time.Time `json:"fecVtoEnt"`

	Pendientes float64 `json:"pendientes"` // TONELADAS
	PendApli   float64 `json:"pendApli"`

	IttLiqTotal   float64 `json:"ittliqttotal"`
	IttLiquidadas float64 `json:"ittliquidadas"`
	Operacion     string  `json:"operacion"`

	FecDesde *time.Time `json:"fecdesde"`
	FecHasta *time.Time `json:"fechasta"`

	Grano     string `json:"grano"`
	NomGrano  string `json:"nomgrano"`
	Cosecha   string `json:"cosecha"`
	ContParte string `json:"contparte"`

	DestinoNombre string `json:"destinoNombre"`
	Observacion   string `json:"observacion"`
}

type PendingFixFilters struct {
	UniNego string
	CUIT    string
	VendCta string
	CompCta string

	Contrato   string
	ContParte  string
	CompNombre string

	MinPendientes float64
	MinPendApli   float64

	FecEntDesde    *time.Time
	FecEntHasta    *time.Time
	FecDesdeDesde  *time.Time
	FecDesdeHasta  *time.Time
	FecHastaDesde  *time.Time
	FecHastaHasta  *time.Time
	FecVtoEntDesde *time.Time
	FecVtoEntHasta *time.Time
}
