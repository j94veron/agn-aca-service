package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
  g.cuit as cuit,
  (CASE 
		WHEN h.stipprovee = 51 THEN 'COOP'
		WHEN  h.stipprovee = 63 THEN 'CDC'
		ELSE 'TERCERO'
  END) AS segmento,
  SUBSTR(TO_CHAR(b.vendcta),1,12)                AS vendcta,
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

FROM %s.contrato b
JOIN %s.grano c
  ON b.grano = c.grano
 AND b.cosecha = c.cosecha
JOIN mcuenta e
  ON b.destino = e.cuenta
LEFT JOIN (
  SELECT continterno, fecfijadesde, fecfijahasta
    FROM %s.ctfijar t
   WHERE fecfijahasta = (
     SELECT MAX(u.fecfijahasta)
       FROM %s.ctfijar u
      WHERE u.continterno = t.continterno
   )
) d
  ON b.continterno = d.continterno
JOIN mcuenta h
  ON b.vendcta = h.cuenta
JOIN mcuit g
  ON h.ctamadre = g.ctamadre
WHERE b.operacion IN (
  'AFC','AFL','CFC','SBOA','AFALM','AFANEP','AFAPSE',
  'AFBOL','AFCPSE','AFCNEP','AFWAR','SBOANP',
  'CANAF','AFBNEP','AFWNEP','CANFNE'
)
AND b.negocio = 'CV'
AND b.status = 50
AND (b.ttmaxfin - b.ittliqttotal) > 0
AND (b.ttmaxfin - b.ittfijadas)   > 0
`

// ===== struct intermedio para scan (NULL-safe) =====

type scanPendingFixRow struct {
	CUIT            sql.NullString `db:"CUIT"`
	Segmento        sql.NullString `db:"SEGMENTO"`
	VendCta         sql.NullString `db:"VENDCTA"`
	VendNombre      sql.NullString `db:"VENDNOMBRE"`
	CompCta         sql.NullString `db:"COMPCTA"`
	CompNombre      sql.NullString `db:"COMPNOMBRE"`
	VendedorNombre  sql.NullString `db:"VENDEDORNOMBRE"`
	CompradorNombre sql.NullString `db:"COMPRADORNOMBRE"`

	Contrato     sql.NullString  `db:"CONTRATO"`
	TTMaxPact    sql.NullFloat64 `db:"TTMAXPACT"`
	IttAplicadas sql.NullFloat64 `db:"ITTAPLICADAS"`
	IttFijadas   sql.NullFloat64 `db:"ITTFIJADAS"`
	ContVendedor sql.NullString  `db:"CONTVENDEDOR"`

	FecEnt    sql.NullTime `db:"FECENT"`
	FecVtoEnt sql.NullTime `db:"FECVTOENT"`

	Pendientes sql.NullFloat64 `db:"PENDIENTES"`
	PendApli   sql.NullFloat64 `db:"PENDAPLI"`

	IttLiqTotal   sql.NullFloat64 `db:"ITTLIQTTOTAL"`
	IttLiquidadas sql.NullFloat64 `db:"ITTLIQUIDADAS"`
	Operacion     sql.NullString  `db:"OPERACION"`

	FecDesde sql.NullTime `db:"FECDESDE"`
	FecHasta sql.NullTime `db:"FECHASTA"`

	Grano     sql.NullString `db:"GRANO"`
	NomGrano  sql.NullString `db:"NOMGRANO"`
	Cosecha   sql.NullString `db:"COSECHA"`
	ContParte sql.NullString `db:"CONTPARTE"`

	DestinoNombre sql.NullString `db:"DESTINONOMBRE"`
	Observacion   sql.NullString `db:"OBSERVACION"`
}

// ===== método público =====

func (r *OracleRepo) FetchPendingFixAll(
	ctx context.Context,
	schema string,
	uninego string,
) ([]domain.PendingFixRow, error) {
	return r.fetchBySchema(ctx, schema, uninego)
}

// ===== mapper =====

func mapPendingFixRow(x scanPendingFixRow, uninego string) domain.PendingFixRow {
	r := domain.PendingFixRow{
		UniNego:         uninego,
		CUIT:            ns(x.CUIT),
		Segmento:        ns(x.Segmento),
		VendCta:         ns(x.VendCta),
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

func (r *OracleRepo) fetchBySchema(
	ctx context.Context,
	schema string,
	uninego string,
) ([]domain.PendingFixRow, error) {

	query := fmt.Sprintf(sqlPendingFixAll, schema, schema, schema, schema)

	var rows []scanPendingFixRow

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := r.db.SelectContext(ctxTimeout, &rows, query); err != nil {
		return nil, fmt.Errorf("oracle FetchPendingFixAll %s: %w", schema, err)
	}

	out := make([]domain.PendingFixRow, len(rows))
	for i := range rows {
		out[i] = mapPendingFixRow(rows[i], uninego)
	}

	return out, nil
}
