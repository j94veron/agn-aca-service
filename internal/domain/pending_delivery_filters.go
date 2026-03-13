package domain

import "time"

type PendingDeliveryFilter struct {
	UniNego string

	Segmento string

	CuitVendedor  string
	CuitComprador string

	VendCta string
	CompCta string

	Contrato  string
	ContParte string

	Grano   string
	NomComp string

	FecEntDesde *time.Time
	FecEntHasta *time.Time

	FecVtoDesde *time.Time
	FecVtoHasta *time.Time
}
