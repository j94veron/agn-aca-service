package domain

import "time"

// Row “normalizado” para Redis (tipos seguros, fechas como *time.Time)
type PendingDeliveryRow struct {
	UniNego string `json:"uninego"` // "CO" o "AC" (origen / usuario Oracle)
	Schema  string `json:"schema"`  // opcional: "CORRETAJE" o "ACOPIO" (debug/auditoría)

	Zona      string `json:"zona"`
	Contrato  string `json:"contrato"`
	ContParte string `json:"contparte"`

	CUITCorre     string `json:"cuitcorre"`
	CUITComprador string `json:"cuitcomprador"`
	CUITVendedor  string `json:"cuitvendedor"`

	VendCta    string `json:"vendcta"`
	NomVend    string `json:"nomvend"`
	CompCta    string `json:"compcta"`
	NomComp    string `json:"nomcomp"`
	CorrCta    string `json:"corrcta"`
	NomCorr    string `json:"nomcorr"`
	Destino    string `json:"destino"`
	NomDestino string `json:"nomdestino"`
	Localidad  string `json:"localidad"`

	Grano    string `json:"grano"`
	NomGrano string `json:"nomgrano"`
	Cosecha  string `json:"cosecha"`

	Operacion string `json:"operacion"`

	Fecha     *time.Time `json:"fecha,omitempty"`
	FecEnt    *time.Time `json:"fecent,omitempty"`
	FecVtoEnt *time.Time `json:"fecvtoent,omitempty"`

	Precio float64 `json:"precio"`
	Moneda string  `json:"moneda"`

	KilPact    float64 `json:"kilpact"`
	KilAmRe    float64 `json:"kilamre"`
	KilAplica  float64 `json:"kilaplica"`
	KilPenApli float64 `json:"kilpenapli"`
	KilLiquid  float64 `json:"killiquid"`
	KilPenLiq  float64 `json:"kilpenliq"`

	Observaciones string `json:"observaciones"`
}

type PendingDeliverySnapshot struct {
	GeneratedAt time.Time            `json:"generatedAt"`
	Rows        []PendingDeliveryRow `json:"rows"`
}

type PendingDeliveryQuery struct {
	UniNego       string // "CO" | "AC" | "" (ALL)
	CUITVendedor  string
	CUITComprador string
	Grano         string

	// rango por FecVtoEnt
	FecVtoDesde *time.Time
	FecVtoHasta *time.Time

	// paginado
	Offset int
	Limit  int
}

type PendingDeliveryResponse struct {
	GeneratedAt time.Time            `json:"generatedAt"`
	Count       int                  `json:"count"`
	Offset      int                  `json:"offset"`
	Limit       int                  `json:"limit"`
	Rows        []PendingDeliveryRow `json:"rows"`
}
