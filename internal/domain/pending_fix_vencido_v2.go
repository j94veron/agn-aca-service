package domain

import "time"

type PendingFixVencidosV2Grano struct {
	Grano   string  `json:"grano"`
	Cosecha string  `json:"cosecha"`
	Tn      float64 `json:"tn"`
}

type PendingFixVencidosV2Month struct {
	Month  string                      `json:"month"`  // YYYY-MM
	Granos []PendingFixVencidosV2Grano `json:"granos"` // grano+cosecha
}

type PendingFixVencidosV2Row struct {
	UniNego string                      `json:"uninego"`
	CUIT    string                      `json:"cuit"`
	Months  []PendingFixVencidosV2Month `json:"months"` // siempre 12 meses
}

type PendingFixVencidosV2Snapshot struct {
	GeneratedAt  time.Time                 `json:"generatedAt"`
	WindowMonths int                       `json:"windowMonths"`
	Rows         []PendingFixVencidosV2Row `json:"rows"`
}
