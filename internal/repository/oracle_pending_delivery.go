package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"agn-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

// OJO: esto usa database/sql por debajo.
// Mientras tu db.NewOracleDB() abra con driver "godror", sqlx está OK.
type OraclePendingEntregaRepo struct {
	db *sqlx.DB
}

func NewOraclePendingEntregaRepo(db *sqlx.DB) *OraclePendingEntregaRepo {
	return &OraclePendingEntregaRepo{db: db}
}

// SQL “normalizado”:
// - Sin TO_CHAR para fechas (scan NullTime)
// - Sin filtros tipo '?XDVEND' (traemos TODO y filtramos en Redis)
// - Subqueries agregan Aplicados y Liquidados por ContInterno
const sqlPendingDeliveryAll = `
SELECT
  (CASE 
            WHEN e.stipprovee = 51 THEN 'COOP'
            WHEN  e.stipprovee = 63 THEN 'CDC'
            ELSE 'TERCERO'
  END) AS segmento,
  a.zona                                       AS zona,
  a.contrato                                   AS contrato,
  a.contparte                                  AS contparte,
  h.cuit                                       AS cuitcorre,
  a.compcuit                                   AS cuitcomprador,
  a.vendcuit                                   AS cuitvendedor,
  SUBSTR(TO_CHAR(a.vendcta),1,12)              AS vendcta,
  SUBSTR(cambocur(e.nombre,'"',''),1,30)       AS nomvend,
  SUBSTR(TO_CHAR(a.compcta),1,12)              AS compcta,
  SUBSTR(cambocur(f.nombre,'"',''),1,30)       AS nomcomp,
  SUBSTR(TO_CHAR(a.corre1cta),1,12)            AS corrcta,
  SUBSTR(cambocur(a.corre1nombre,'"',''),1,30) AS nomcorr,
  SUBSTR(TO_CHAR(a.grano),1,3)                 AS grano,
  SUBSTR(b_tb0('GRANO',LPAD(TO_CHAR(a.grano),3,' '),'GR2'),1,20) AS nomgrano,
  %s.date_c(a.fecha)                           AS fecha,
  %s.date_c(a.fecent)                          AS fecent,
  %s.date_c(a.fecvtoent)                       AS fecvtoent,
  a.cosecha                                    AS cosecha,
  NVL(a.precio,0)                              AS precio,
  SUBSTR(%s.b_tb0('MONEDA',a.moneda,'MN2'),1,15) AS moneda,
  a.operacion                                  AS operacion,
  SUBSTR(TO_CHAR(a.destino),1,12)              AS destino,
  SUBSTR(cambocur(d.nombre,'"',''),1,30)       AS nomdestino,
  d.locali                                     AS localidad,
  NVL(a.ttmaxpact*1000,0)                      AS kilpact,
  NVL((a.ttmaxfin-a.ttmaxpact)*1000,0)         AS kilamre,
  NVL(b.aplicados,0)                           AS kilaplica,
  NVL((a.ttmaxfin*1000)-NVL(b.aplicados,0),0)  AS kilpenapli,
  NVL(c.liquidados,0)                          AS killiquid,
  NVL((a.ttmaxfin*1000)-NVL(c.liquidados,0),0) AS kilpenliq,
  SUBSTR(a.observacion,1,200)                  AS observaciones
FROM %s.contrato a
LEFT JOIN (
  SELECT m.continterno, SUM(m.kilaplica) aplicados
  FROM %s.relapli m
  GROUP BY m.continterno
) b ON b.continterno = a.continterno
LEFT JOIN (
  SELECT oc.continterno,
         SUM(
           oc.kilos * DECODE(oc.tipcalculo,'67',-1,'62',-1,1)
           - DECODE(oc.tipcalculo,'67',0,'62',0,
               NVL((
                 SELECT SUM(s.kilos)
                 FROM %s.olparcial s
                 WHERE s.ordenliquid = oc.ordenliquid
                   AND s.tipcalculo = '5'
               ),0)
             )
         ) liquidados
  FROM %s.olcabez oc
  WHERE oc.tipcalculo IN('4','5','62','67')
  GROUP BY oc.continterno
) c ON c.continterno = a.continterno
JOIN %s.mcuenta d ON d.empresa = a.empresa AND d.cuenta = a.destino
JOIN %s.mcuenta e ON e.empresa = a.empresa AND e.cuenta = a.vendcta
JOIN %s.mcuenta f ON f.empresa = a.empresa AND f.cuenta = a.compcta
JOIN %s.mcuenta g ON g.cuenta  = a.corre1cta
JOIN %s.mcuit   h ON g.ctamadre= h.ctamadre
WHERE a.status IN('50','90')
  AND (a.ttmaxfin*1000) - NVL(b.aplicados,0) > 0
`

type scanPendingDeliveryRow struct {
	Segmento  sql.NullString `db:"SEGMENTO"`
	Zona      sql.NullString `db:"ZONA"`
	Contrato  sql.NullString `db:"CONTRATO"`
	ContParte sql.NullString `db:"CONTPARTE"`

	CUITCorre     sql.NullString `db:"CUITCORRE"`
	CUITComprador sql.NullString `db:"CUITCOMPRADOR"`
	CUITVendedor  sql.NullString `db:"CUITVENDEDOR"`

	VendCta sql.NullString `db:"VENDCTA"`
	NomVend sql.NullString `db:"NOMVEND"`
	CompCta sql.NullString `db:"COMPCTA"`
	NomComp sql.NullString `db:"NOMCOMP"`
	CorrCta sql.NullString `db:"CORRCTA"`
	NomCorr sql.NullString `db:"NOMCORR"`

	Grano    sql.NullString `db:"GRANO"`
	NomGrano sql.NullString `db:"NOMGRANO"`
	Cosecha  sql.NullString `db:"COSECHA"`

	Operacion  sql.NullString `db:"OPERACION"`
	Moneda     sql.NullString `db:"MONEDA"`
	Destino    sql.NullString `db:"DESTINO"`
	NomDestino sql.NullString `db:"NOMDESTINO"`
	Localidad  sql.NullString `db:"LOCALIDAD"`

	Fecha     sql.NullTime `db:"FECHA"`
	FecEnt    sql.NullTime `db:"FECENT"`
	FecVtoEnt sql.NullTime `db:"FECVTOENT"`

	Precio     sql.NullFloat64 `db:"PRECIO"`
	KilPact    sql.NullFloat64 `db:"KILPACT"`
	KilAmRe    sql.NullFloat64 `db:"KILAMRE"`
	KilAplica  sql.NullFloat64 `db:"KILAPLICA"`
	KilPenApli sql.NullFloat64 `db:"KILPENAPLI"`
	KilLiquid  sql.NullFloat64 `db:"KILLIQUID"`
	KilPenLiq  sql.NullFloat64 `db:"KILPENLIQ"`

	Observaciones sql.NullString `db:"OBSERVACIONES"`
}

func (r *OraclePendingEntregaRepo) FetchAll(ctx context.Context, schema string, uninego string) ([]domain.PendingDeliveryRow, error) {
	// Repetimos schema en todos los lugares y el prefijo para date_c / b_tb0
	q := fmt.Sprintf(
		sqlPendingDeliveryAll,
		schema, schema, schema, schema, // date_c, date_c, date_c, b_tb0
		schema,         // FROM %s.contrato
		schema,         // relapli
		schema, schema, // olparcial, olcabez
		schema, schema, schema, schema, schema, // mcuenta d,e,f,g y mcuit h
	)

	ctxTimeout, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	var rows []scanPendingDeliveryRow
	if err := r.db.SelectContext(ctxTimeout, &rows, q); err != nil {
		return nil, fmt.Errorf("oracle pending_entrega FetchAll %s: %w", schema, err)
	}

	out := make([]domain.PendingDeliveryRow, len(rows))
	for i := range rows {
		out[i] = mapPendingDeliveryRow(rows[i], schema, uninego)
	}
	return out, nil
}

func mapPendingDeliveryRow(x scanPendingDeliveryRow, schema, uninego string) domain.PendingDeliveryRow {
	r := domain.PendingDeliveryRow{
		UniNego: uninego,
		Schema:  schema,

		Segmento:  ns(x.Segmento),
		Zona:      ns(x.Zona),
		Contrato:  ns(x.Contrato),
		ContParte: ns(x.ContParte),

		CUITCorre:     ns(x.CUITCorre),
		CUITComprador: ns(x.CUITComprador),
		CUITVendedor:  ns(x.CUITVendedor),

		VendCta: ns(x.VendCta),
		NomVend: ns(x.NomVend),
		CompCta: ns(x.CompCta),
		NomComp: ns(x.NomComp),
		CorrCta: ns(x.CorrCta),
		NomCorr: ns(x.NomCorr),

		Grano:    ns(x.Grano),
		NomGrano: ns(x.NomGrano),
		Cosecha:  ns(x.Cosecha),

		Operacion: ns(x.Operacion),
		Precio:    nf(x.Precio),
		Moneda:    ns(x.Moneda),

		Destino:    ns(x.Destino),
		NomDestino: ns(x.NomDestino),
		Localidad:  ns(x.Localidad),

		KilPact:    nf(x.KilPact),
		KilAmRe:    nf(x.KilAmRe),
		KilAplica:  nf(x.KilAplica),
		KilPenApli: nf(x.KilPenApli),
		KilLiquid:  nf(x.KilLiquid),
		KilPenLiq:  nf(x.KilPenLiq),

		Observaciones: ns(x.Observaciones),
	}

	if x.Fecha.Valid {
		t := x.Fecha.Time
		r.Fecha = &t
	}
	if x.FecEnt.Valid {
		t := x.FecEnt.Time
		r.FecEnt = &t
	}
	if x.FecVtoEnt.Valid {
		t := x.FecVtoEnt.Time
		r.FecVtoEnt = &t
	}

	return r
}
