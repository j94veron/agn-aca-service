package domain

// Reglas de negocio (high-level) y DTOs

type OliqUiRequest struct {
	// Clave idempotente del origen (antes IDAGN / pUValueOrigen)
	IDAGN         string  `json:"idagn" validate:"required"`
	ContInterno   int64   `json:"contInterno" validate:"required"`
	UniNego       string  `json:"uninego" validate:"required"`
	FecCreacion   int64   `json:"fecCreacion"`
	FecDesde      int64   `json:"fecDesde"`
	DiasVto       int     `json:"diasVto"`
	FechaHasta    string  `json:"fechaHasta"`
	FechaPizarra  string  `json:"fechaPizarra"`
	HoraFijar     string  `json:"horaFijar"`
	OlTTMin       float64 `json:"olTTMin" validate:"gt=0"`
	OlTTMax       float64 `json:"olTTMax" validate:"gtefield=OlTTMin"`
	Precio        float64 `json:"precio" validate:"gt=0"`
	FecOrigen     int64   `json:"fecOrigen"`
	POrdenLiq     string  `json:"pOrdenLiq"` // Tipo precio de Orden de Liquidar (negocio)
	ComiCta       float64 `json:"comiCta"`
	Cotizacion    float64 `json:"cotizacion"`
	TipoPago      string  `json:"tipoPago"`
	CodigoSio     int64   `json:"codigoSio"`
	FechaSio      int64   `json:"fechaSio"`
	HoraSio       int64   `json:"horaSio"`
	MonFijada     string  `json:"monFijada"` // '1' dolar, sino peso (reglas previas)
	Moneda        string  `json:"moneda"`    // '1' dolar, sino peso
	Observaciones string  `json:"observaciones"`
	OrigenTipo    string  `json:"origenTipo"` // 'OP' | 'OL'
}

type CtamreRequest struct {
	IDAGN        string  `json:"idagn" validate:"required"`
	ContInterno  int64   `json:"contInterno" validate:"required"`
	UniNego      string  `json:"uninego" validate:"required"`
	Fecha        string  `json:"fecha"`
	Toneladas    float64 `json:"toneladas" validate:"gt=0"`
	Precio       float64 `json:"precio" validate:"gte=0"`
	AmpRed       int     `json:"ampRed" validate:"oneof=1 2"` // 1=Ampliación, 2=Reducción
	Camiones     float64 `json:"camiones"`
	Pizarra      int     `json:"pizarra"`
	Porcentaje   float64 `json:"porcentaje"`
	ImpAdicional float64 `json:"impAdicional"`
	Observacion  string  `json:"observacion"`
}

type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type OliquiTooltipDTO struct {
	Continterno    *int64   `json:"continterno"`
	IDAGN          *string  `json:"idagn"`
	Comision       *float64 `json:"comision,omitempty"`
	Cosecha        *string  `json:"cosecha"`
	AreaGeografica *string  `json:"areaGeografica"`
	TipoEntrega    *string  `json:"tipoEntrega"`
	Pauta          *float64 `json:"pauta,omitempty"`
	ValorVenta     *float64 `json:"valorVenta,omitempty"`
	ValorPizarra   *float64 `json:"valorPizarra,omitempty"`
	Flete          *float64 `json:"flete,omitempty"`
	EntregaDesde   *string  `json:"entregaDesde"`
	EntregaHasta   *string  `json:"entregaHasta"`
	Comprador      *string  `json:"comprador"`
	Centro         *string  `json:"centro"`
	Destino        *string  `json:"destino"`
	Zona           *string  `json:"zona"`
	ContParte      *string  `json:"contparte"`
	Mercaderia     *string  `json:"tipocali"`
	Negocio        *int64   `json:"negocio,omitempty"`
	Grano          *int64   `json:"grano,omitempty"`
}

type OliquiGridDTO struct {
	Continterno    *int64   `json:"continterno"`
	IDAGN          *string  `json:"idagn"`
	Comprador      *string  `json:"comprador"`
	Comision       *float64 `json:"comision"`
	Cosecha        *string  `json:"cosecha"`
	Destino        *string  `json:"destino"`
	EntregaDesde   *string  `json:"entregaDesde"`
	EntregaHasta   *string  `json:"entregaHasta"`
	FechaVto       *string  `json:"fechaVto"`
	Hora           *string  `json:"hora"`
	ImpAdicional   *float64 `json:"impAdicional"`
	Moneda         *string  `json:"moneda"`
	Negocio        *int64   `json:"negocio"`
	AreaGeografica *string  `json:"areaGeografica"`
	TipoEntrega    *string  `json:"tipoEntrega"`
	Paridad        *string  `json:"paridad"`
	TipoDeNegocio  *string  `json:"tipoDeNegocio"`
	TTMax          *float64 `json:"ttMax"`
	Grano          *int64   `json:"grano"`
	Zona           *string  `json:"zona"`
}
