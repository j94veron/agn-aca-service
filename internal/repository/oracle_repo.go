package repository

import (
	"context"
	"database/sql"
	"fmt"

	"agn-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type OracleRepo struct {
	db *sqlx.DB
}

func NewOracleRepo(db *sqlx.DB) *OracleRepo {
	return &OracleRepo{db: db}
}

/*
SQL basado en el que pasaron:
- Alias alineados al struct
- Parámetros :fecha_desde / :fecha_hasta
- Sin filtro por CUIT (snapshot anualizado global)
*/
const sqlPendingFixAll = `
SELECT
  SUBSTR(TO_CHAR(b.vendcta),1,12)                AS cuit,
  b.vendnombre                                   AS vendnombre,
  SUBSTR(TO_CHAR(b.compcta),1,12)                AS compcta,
  b.compnombre                                   AS compnombre,
  SUBSTR(TO_CHAR(b.vendcta),1,12)||'-'||b.vendnombre AS vendedorNombre,
  SUBSTR(TO_CHAR(b.compcta),1,12)||'-'||b.compnombre AS compradorNombre,

  b.contrato                                     AS contrato,
  b.ttmaxpact                                    AS ttmaxpact,
  b.ittaplicadas                                 AS ittaplicadas,
  b.ittfijadas                                   AS ittfijadas,
  b.contvendedor                                 AS contvendedor,

  date_c(b.fecent)                               AS fecent,
  date_c(b.fecvtoent)                            AS fecvtoent,

  (CASE
     WHEN b.ittliqttotal > b.ittfijadas
     THEN (b.ttmaxfin - b.ittliqttotal)
     ELSE (b.ttmaxfin - b.ittfijadas)
   END)                                          AS pendientes,

  (b.ttmaxfin - b.ittaplicadas)                  AS pendapli,

  b.ittliqttotal                                 AS ittliqttotal,
  b.ittliquidadas                                AS ittliquidadas,
  b.operacion                                    AS operacion,

  date_c(d.fecfijadesde)                         AS fecdesde,
  date_c(d.fecfijahasta)                         AS fechasta,

  b.grano                                        AS grano,
  c.nombre                                       AS nomgrano,
  b.cosecha                                      AS cosecha,
  b.contparte                                    AS contparte,

  SUBSTR(TO_CHAR(b.destino),1,12)||'-'||e.nombre AS destinoNombre,
  b.observacion                                  AS observacion

FROM CORRETAJE.contrato b
JOIN CORRETAJE.grano c
  ON b.grano = c.grano
 AND b.cosecha = c.cosecha
JOIN mcuenta e
  ON b.destino = e.cuenta
LEFT JOIN (
  SELECT continterno, fecfijadesde, fecfijahasta
    FROM CORRETAJE.ctfijar t
   WHERE fecfijahasta = (
     SELECT MAX(u.fecfijahasta)
       FROM CORRETAJE.ctfijar u
      WHERE u.continterno = t.continterno
   )
) d
  ON b.continterno = d.continterno

WHERE b.operacion IN (
  'AFC','AFL','CFC','SBOA','AFALM','AFANEP','AFAPSE',
  'AFBOL','AFCPSE','AFCNEP','AFWAR','SBOANP',
  'CANAF','AFBNEP','AFWNEP','CANFNE'
)
AND b.status = 50
AND (b.ttmaxfin - b.ittliqttotal) > 0
AND (b.ttmaxfin - b.ittfijadas)   > 0
AND (
     d.fecfijahasta BETWEEN NDATE_C(TO_DATE(:fecha_desde, 'DD/MM/YYYY')) AND NDATE_C(TO_DATE(:fecha_hasta, 'DD/MM/YYYY'))
     OR d.fecfijahasta IS NULL
)
`

// ===== struct intermedio para scan (NULL-safe) =====

type scanPendingFixRow struct {
	CUIT sql.NullString `db:"cuit"`
	//VendCta         sql.NullString `db:"vendcta"`
	VendNombre      sql.NullString `db:"vendnombre"`
	CompCta         sql.NullString `db:"compcta"`
	CompNombre      sql.NullString `db:"compnombre"`
	VendedorNombre  sql.NullString `db:"vendedorNombre"`
	CompradorNombre sql.NullString `db:"compradorNombre"`

	Contrato     sql.NullString  `db:"contrato"`
	TTMaxPact    sql.NullFloat64 `db:"ttmaxpact"`
	IttAplicadas sql.NullFloat64 `db:"ittaplicadas"`
	IttFijadas   sql.NullFloat64 `db:"ittfijadas"`
	ContVendedor sql.NullString  `db:"contvendedor"`

	FecEnt    sql.NullTime `db:"fecent"`
	FecVtoEnt sql.NullTime `db:"fecvtoent"`

	Pendientes sql.NullFloat64 `db:"pendientes"`
	PendApli   sql.NullFloat64 `db:"pendapli"`

	IttLiqTotal   sql.NullFloat64 `db:"ittliqttotal"`
	IttLiquidadas sql.NullFloat64 `db:"ittliquidadas"`
	Operacion     sql.NullString  `db:"operacion"`

	FecDesde sql.NullTime `db:"fecdesde"`
	FecHasta sql.NullTime `db:"fechasta"`

	Grano     sql.NullString `db:"grano"`
	NomGrano  sql.NullString `db:"nomgrano"`
	Cosecha   sql.NullString `db:"cosecha"`
	ContParte sql.NullString `db:"contparte"`

	DestinoNombre sql.NullString `db:"destinoNombre"`
	Observacion   sql.NullString `db:"observacion"`
}

// ===== método público =====

func (r *OracleRepo) FetchPendingFixAll(
	ctx context.Context,
	fechaDesdeDDMMYYYY string,
	fechaHastaDDMMYYYY string,
	uninego string,
) ([]domain.PendingFixRow, error) {

	params := map[string]any{
		"fecha_desde": fechaDesdeDDMMYYYY,
		"fecha_hasta": fechaHastaDDMMYYYY,
	}

	// 1️ Convertir named params -> positional
	query, args, err := sqlx.Named(sqlPendingFixAll, params)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Named: %w", err)
	}

	// 2️ Rebind para Oracle (:1, :2)
	query = r.db.Rebind(query)

	// 3️ Ejecutar con args...
	var rows []scanPendingFixRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("oracle FetchPendingFixAll: %w", err)
	}

	out := make([]domain.PendingFixRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, mapPendingFixRow(x, uninego))
	}

	return out, nil
}

// ===== mapper =====

func mapPendingFixRow(x scanPendingFixRow, uninego string) domain.PendingFixRow {
	r := domain.PendingFixRow{
		UniNego: uninego,
		CUIT:    ns(x.CUIT),
		//VendCta:         ns(x.VendCta),
		VendNombre:      ns(x.VendNombre),
		CompCta:         ns(x.CompCta),
		CompNombre:      ns(x.CompNombre),
		VendedorNombre:  ns(x.VendedorNombre),
		CompradorNombre: ns(x.CompradorNombre),
		Contrato:        ns(x.Contrato),
		TTMaxPact:       nf(x.TTMaxPact),
		IttAplicadas:    nf(x.IttAplicadas),
		IttFijadas:      nf(x.IttFijadas),
		ContVendedor:    ns(x.ContVendedor),
		Pendientes:      nf(x.Pendientes),
		PendApli:        nf(x.PendApli),
		IttLiqTotal:     nf(x.IttLiqTotal),
		IttLiquidadas:   nf(x.IttLiquidadas),
		Operacion:       ns(x.Operacion),
		Grano:           ns(x.Grano),
		NomGrano:        ns(x.NomGrano),
		Cosecha:         ns(x.Cosecha),
		ContParte:       ns(x.ContParte),
		DestinoNombre:   ns(x.DestinoNombre),
		Observacion:     ns(x.Observacion),
	}

	if x.FecEnt.Valid {
		t := x.FecEnt.Time
		r.FecEnt = &t
	}
	if x.FecVtoEnt.Valid {
		t := x.FecVtoEnt.Time
		r.FecVtoEnt = &t
	}
	if x.FecDesde.Valid {
		t := x.FecDesde.Time
		r.FecDesde = &t
	}
	if x.FecHasta.Valid {
		t := x.FecHasta.Time
		r.FecHasta = &t
	}

	return r
}

// ===== helpers =====

func ns(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nf(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0
}
