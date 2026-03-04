package repository

import (
	"agn-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"agn-service/internal/oracle"
)

type OliquiRepo struct{ oc *oracle.Client }

func NewOliquiRepo(oc *oracle.Client) *OliquiRepo { return &OliquiRepo{oc: oc} }

func (r *OliquiRepo) ExistsByIDAGN(ctx context.Context, schema string, idagn string) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s.OLIQUI WHERE IDAGN = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, idagn)
	var x int
	err := row.Scan(&x)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *OliquiRepo) Insert(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) (int64, error) {
	q := fmt.Sprintf(`
		INSERT INTO %s.OLIQUI (
			EMPRESA, CONTINTERNO, CONTRATO, NEGOCIO, CUENTA, UNINEGO, ZONA,
			ORDENINTER, ORDENLIQUID, TIPORDEN, FECCREACION, FECDESDE, DIASVTO,
			FECHASTA, FECHAPIZARRA, HORAFIJAR, OL_TTMIN, OL_TTMAX,
			OL_TIPPREC, OL_TIPPIZA, OL_PIZARRA, OL_PRECIO,
			PORDENLIQ, MONEDA, STATUS, OBSERVA,
			IDAGN
		)
		VALUES (
			:empresa, :continterno, :contrato, :negocio, :cuenta, :uninego, :zona,
			:ordeninter, :ordenliquid, :tiporden, :feccreacion, :fecdesde, :diasvto,
			:fechasta, :fechapizarra, :horafijar, :ol_ttmin, :ol_ttmax,
			:ol_tipprec, :ol_tippiza, :ol_pizarra, :ol_precio,
			:pordenliq, :moneda, :status, :observa,
			:idagn
		)`, schema)
	_, err := tx.ExecContext(ctx, q, cols)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func (r *OliquiRepo) baseSQL(f domain.OliquiFilter) string {

	tContrato := qualify(f.Schema, "CONTRATO")
	tGrano := qualify(f.Schema, "GRANO")
	tCtfijar := qualify(f.Schema, "CTFIJAR")

	return fmt.Sprintf(`
WITH base AS (

SELECT
  g.cuit,
  b.vendcta,
  b.vendnombre,
  b.compcta compcuit,
  b.compnombre,
  b.contrato,
  b.ttmaxpact,
  b.ittaplicadas,
  b.ittfijadas,
  b.contvendedor,
  b.fecent,
  b.fecvtoent,
  b.grano,
  b.cosecha,
  b.destino,
  b.operacion,
  b.zona,
  b.areageo,
  b.tipoentrega,
  b.centro,
  b.continterno,

  d.fecfijadesde,
  d.fecfijahasta,

  c.nombre nomgrano

FROM %s b

JOIN %s c
  ON b.grano = c.grano
 AND b.cosecha = c.cosecha

LEFT JOIN (
  SELECT
    continterno,
    MAX(fecfijadesde) KEEP (DENSE_RANK LAST ORDER BY fecfijahasta) fecfijadesde,
    MAX(fecfijahasta) fecfijahasta
  FROM %s
  GROUP BY continterno
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
AND (b.ttmaxfin - b.ittfijadas) > 0

)
`, tContrato, tGrano, tCtfijar)
}

func (r *OliquiRepo) CountTooltip(ctx context.Context, f domain.OliquiFilter) (int, error) {

	base := r.baseSQL(f)

	sql := base + `
SELECT COUNT(*)
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)
	sql += where

	var total int

	err := r.oc.DB.QueryRowContext(ctx, sql, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *OliquiRepo) CountGrid(ctx context.Context, f domain.OliquiFilter) (int, error) {

	base := r.baseSQL(f)

	sql := base + `
SELECT COUNT(*)
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)
	sql += where

	var total int

	err := r.oc.DB.QueryRowContext(ctx, sql, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// otra cosa
func (r *OliquiRepo) GetTooltip(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiTooltipDTO, int, error) {

	base := r.baseSQL(f)

	query := base + `
SELECT
  NULL,
  cosecha,
  areageo,
  tipoentrega,
  NULL,
  NULL,
  NULL,
  NULL,
  date_c(fecent),
  date_c(fecvtoent),
  compcuit,
  centro,
  destino,
  zona,
  contrato,
  grano
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)
	query = buildPagedQuery(query+where, f)

	rows, err := r.oc.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var res []domain.OliquiTooltipDTO
	var total int

	for rows.Next() {

		var d domain.OliquiTooltipDTO

		if err := rows.Scan(
			&d.Comision,
			&d.Cosecha,
			&d.AreaGeografica,
			&d.TipoEntrega,
			&d.Pauta,
			&d.ValorVenta,
			&d.ValorPizarra,
			&d.Flete,
			&d.EntregaDesde,
			&d.EntregaHasta,
			&d.Comprador,
			&d.Centro,
			&d.Destino,
			&d.Zona,
			&d.Negocio,
			&d.Grano,
			&total,
		); err != nil {
			return nil, 0, err
		}

		res = append(res, d)
	}

	return res, total, nil
}

func (r *OliquiRepo) GetGrid(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiGridDTO, int, error) {

	base := r.baseSQL(f)

	query := base + `
SELECT
  compcuit,
  NULL,
  cosecha,
  destino,
  date_c(fecent),
  date_c(fecvtoent),
  date_c(fecfijahasta),
  NULL,
  NULL,
  NULL,
  contrato,
  areageo,
  tipoentrega,
  NULL,
  operacion,
  grano,
  ttmaxpact,
  zona
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)
	query = buildPagedQuery(query+where, f)

	rows, err := r.oc.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var res []domain.OliquiGridDTO
	var total int

	for rows.Next() {

		var d domain.OliquiGridDTO
		var totalRows int

		if err := rows.Scan(
			&d.Comprador,
			&d.Comision,
			&d.Cosecha,
			&d.Destino,
			&d.EntregaDesde,
			&d.EntregaHasta,
			&d.FechaVto,
			&d.Hora,
			&d.ImpAdicional,
			&d.Moneda,
			&d.Negocio,
			&d.AreaGeografica,
			&d.TipoEntrega,
			&d.Paridad,
			&d.TipoDeNegocio,
			&d.Grano,
			&d.TTMax,
			&d.Zona,
			&totalRows,
			new(int),
		); err != nil {
			return nil, 0, err
		}

		total = totalRows
		res = append(res, d)
	}

	return res, total, nil
}

// /Gets
func (r *OliquiRepo) GetTooltipPage(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiTooltipDTO, error) {

	base := r.baseSQL(f) + `
SELECT
  continterno,
  NULL idagn,
  NULL corre1porcomiventa,
  cosecha,
  areageo,
  tipoentrega,
  NULL uvaluepauta,
  NULL ol_precio,
  NULL ol_pizarra,
  NULL tariflete,
  date_c(fecent),
  date_c(fecvtoent),
  compcuit,
  centro,
  destino,
  zona,
  contrato,
  grano
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)

	page, size := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	from := (page-1)*size + 1
	to := page * size

	order := orderExpr(f.OrderBy)
	dir := "ASC"
	if strings.ToUpper(f.OrderDir) == "DESC" {
		dir = "DESC"
	}

	query := fmt.Sprintf(`
SELECT *
FROM (
  SELECT q.*, ROWNUM rn
  FROM (
    %s %s
    ORDER BY %s %s
  ) q
  WHERE ROWNUM <= :%d
)
WHERE rn >= :%d
`, base, where, order, dir, len(args)+2, len(args)+1)

	args = append(args, to, from)

	rows, err := r.oc.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []domain.OliquiTooltipDTO

	for rows.Next() {

		var rn int

		var (
			continterno, negocio, grano                      sql.NullInt64
			idagn                                            sql.NullString
			comision, pauta, valorVenta, valorPizarra, flete sql.NullFloat64
			cosecha, areaGeo, tipoEntrega                    sql.NullString
			entDesde, entHasta                               sql.NullString
			comprador, centro, destino, zona                 sql.NullString
		)

		if err := rows.Scan(
			&continterno,
			&idagn,
			&comision,
			&cosecha,
			&areaGeo,
			&tipoEntrega,
			&pauta,
			&valorVenta,
			&valorPizarra,
			&flete,
			&entDesde,
			&entHasta,
			&comprador,
			&centro,
			&destino,
			&zona,
			&negocio,
			&grano,
			&rn,
		); err != nil {
			return nil, err
		}

		var d domain.OliquiTooltipDTO

		if continterno.Valid {
			d.Continterno = &continterno.Int64
		}

		if idagn.Valid {
			d.IDAGN = &idagn.String
		}

		if comision.Valid {
			d.Comision = &comision.Float64
		}

		if pauta.Valid {
			d.Pauta = &pauta.Float64
		}

		if valorVenta.Valid {
			d.ValorVenta = &valorVenta.Float64
		}

		if valorPizarra.Valid {
			d.ValorPizarra = &valorPizarra.Float64
		}

		if flete.Valid {
			d.Flete = &flete.Float64
		}

		if cosecha.Valid {
			d.Cosecha = &cosecha.String
		}

		if areaGeo.Valid {
			d.AreaGeografica = &areaGeo.String
		}

		if tipoEntrega.Valid {
			d.TipoEntrega = &tipoEntrega.String
		}

		if entDesde.Valid {
			d.EntregaDesde = &entDesde.String
		}

		if entHasta.Valid {
			d.EntregaHasta = &entHasta.String
		}

		if comprador.Valid {
			d.Comprador = &comprador.String
		}

		if centro.Valid {
			d.Centro = &centro.String
		}

		if destino.Valid {
			d.Destino = &destino.String
		}

		if zona.Valid {
			d.Zona = &zona.String
		}

		if negocio.Valid {
			d.Negocio = &negocio.Int64
		}

		if grano.Valid {
			d.Grano = &grano.Int64
		}

		res = append(res, d)
	}

	return res, nil
}

// ----------------------------------------------------------------------
// GRID (SIN NULL FLOATS EXTRA)
// ----------------------------------------------------------------------

func (r *OliquiRepo) GetGridPage(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiGridDTO, error) {

	base := r.baseSQL(f) + `
SELECT
  continterno,
  NULL idagn,
  compcuit,
  NULL corre1porcomiventa,
  cosecha,
  destino,
  date_c(fecent),
  date_c(fecvtoent),
  date_c(fecfijahasta),
  NULL horafijar,
  NULL impadic,
  NULL monfijada,
  contrato,
  areageo,
  tipoentrega,
  NULL campo4,
  operacion,
  grano,
  ttmaxpact,
  zona
FROM base
WHERE 1=1
`

	where, args := buildFilters(f)

	page, size := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}

	if size <= 0 {
		size = 10
	}

	from := (page-1)*size + 1
	to := page * size

	order := orderExpr(f.OrderBy)

	dir := "ASC"
	if strings.ToUpper(f.OrderDir) == "DESC" {
		dir = "DESC"
	}

	query := fmt.Sprintf(`
SELECT *
FROM (
  SELECT q.*, ROWNUM rn
  FROM (
    %s %s
    ORDER BY %s %s
  ) q
  WHERE ROWNUM <= :%d
)
WHERE rn >= :%d
`, base, where, order, dir, len(args)+2, len(args)+1)

	args = append(args, to, from)

	rows, err := r.oc.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []domain.OliquiGridDTO

	for rows.Next() {

		var rn int

		var (
			continterno, negocio, grano        sql.NullInt64
			idagn                              sql.NullString
			comprador, cosecha, destino        sql.NullString
			entDesde, entHasta, fechaVto, hora sql.NullString
			moneda, areaGeo, tipoEntrega       sql.NullString
			paridad, tipoNegocio, zona         sql.NullString
			comision, impAdic, ttMax           sql.NullFloat64
		)

		if err := rows.Scan(
			&continterno,
			&idagn,
			&comprador,
			&comision,
			&cosecha,
			&destino,
			&entDesde,
			&entHasta,
			&fechaVto,
			&hora,
			&impAdic,
			&moneda,
			&negocio,
			&areaGeo,
			&tipoEntrega,
			&paridad,
			&tipoNegocio,
			&grano,
			&ttMax,
			&zona,
			&rn,
		); err != nil {
			return nil, err
		}

		var d domain.OliquiGridDTO

		if continterno.Valid {
			d.Continterno = &continterno.Int64
		}

		if idagn.Valid {
			d.IDAGN = &idagn.String
		}

		if comprador.Valid {
			d.Comprador = &comprador.String
		}

		if comision.Valid {
			d.Comision = &comision.Float64
		}

		if cosecha.Valid {
			d.Cosecha = &cosecha.String
		}

		if destino.Valid {
			d.Destino = &destino.String
		}

		if entDesde.Valid {
			d.EntregaDesde = &entDesde.String
		}

		if entHasta.Valid {
			d.EntregaHasta = &entHasta.String
		}

		if fechaVto.Valid {
			d.FechaVto = &fechaVto.String
		}

		if hora.Valid {
			d.Hora = &hora.String
		}

		if impAdic.Valid {
			d.ImpAdicional = &impAdic.Float64
		}

		if moneda.Valid {
			d.Moneda = &moneda.String
		}

		if negocio.Valid {
			d.Negocio = &negocio.Int64
		}

		if areaGeo.Valid {
			d.AreaGeografica = &areaGeo.String
		}

		if tipoEntrega.Valid {
			d.TipoEntrega = &tipoEntrega.String
		}

		if paridad.Valid {
			d.Paridad = &paridad.String
		}

		if tipoNegocio.Valid {
			d.TipoDeNegocio = &tipoNegocio.String
		}

		if grano.Valid {
			d.Grano = &grano.Int64
		}

		if ttMax.Valid {
			d.TTMax = &ttMax.Float64
		}

		if zona.Valid {
			d.Zona = &zona.String
		}

		res = append(res, d)
	}

	return res, nil
}
