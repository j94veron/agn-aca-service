package repository

import (
	"fmt"
	"strings"

	"agn-service/internal/domain"
	"agn-service/internal/logger"

	"go.uber.org/zap"
)

// arma el WHERE dinámico
func buildFilters(f domain.OliquiFilter) (string, []interface{}) {
	var where []string
	var args []interface{}
	i := 1

	if f.ContInterno != nil {
		where = append(where, fmt.Sprintf("c.continterno = :%d", i))
		args = append(args, *f.ContInterno)
		i++
	}
	if f.Cosecha != nil {
		where = append(where, fmt.Sprintf("c.cosecha = :%d", i))
		args = append(args, *f.Cosecha)
		i++
	}
	if f.Destino != nil {
		where = append(where, fmt.Sprintf("c.destino = :%d", i))
		args = append(args, *f.Destino)
		i++
	}
	if f.Zona != nil {
		where = append(where, fmt.Sprintf("c.zona = :%d", i))
		args = append(args, *f.Zona)
		i++
	}
	if f.Comprador != nil {
		where = append(where, fmt.Sprintf("c.compcuit = :%d", i))
		args = append(args, *f.Comprador)
		i++
	}
	if f.TipoEntrega != nil {
		where = append(where, fmt.Sprintf("c.tipoentrega = :%d", i))
		args = append(args, *f.TipoEntrega)
		i++
	}
	if f.FechaDesde != nil {
		where = append(where, fmt.Sprintf("date_c(c.fecent) >= TO_DATE(:%d,'YYYY-MM-DD')", i))
		args = append(args, *f.FechaDesde)
		i++
	}
	if f.FechaHasta != nil {
		where = append(where, fmt.Sprintf("date_c(c.fecent) <= TO_DATE(:%d,'YYYY-MM-DD')", i))
		args = append(args, *f.FechaHasta)
		i++
	}

	if len(where) == 0 {
		return "", args
	}
	return " AND " + strings.Join(where, " AND "), args
}

// arma el SELECT paginado para Oracle
func buildPagedQuery(base string, f domain.OliquiFilter) string {

	page := f.Page
	size := f.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	from := (page-1)*size + 1
	to := page * size

	orderBy := orderExpr(f.OrderBy)

	dir := strings.ToUpper(f.OrderDir)
	if dir != "DESC" {
		dir = "ASC"
	}

	query := fmt.Sprintf(`
SELECT *
FROM (
  SELECT q.*, COUNT(*) OVER() AS total_rows,
         ROW_NUMBER() OVER(ORDER BY %s %s) rn
  FROM (
    %s
  ) q
)
WHERE rn BETWEEN %d AND %d
`, orderBy, dir, base, from, to)

	logger.Log.Debug("buildPagedQuery",
		zap.String("orderBy_raw", f.OrderBy),
		zap.String("orderBy_expr", orderBy),
		zap.String("sql", query),
	)

	return query
}

func qualify(schema, table string) string {
	switch schema {
	case "ACOPIO":
		return "ACOPIO." + table
	case "CORRETAJE":
		return "CORRETAJE." + table
	default:
		panic("schema inválido")
	}
}

func orderExpr(o string) string {
	switch strings.ToUpper(strings.TrimSpace(o)) {
	case "ENTREGADESDE":
		return "date_c(c.fecent)"
	case "ENTREGAHASTA":
		return "date_c(c.fecvtoent)"
	case "COSECHA":
		return "c.cosecha"
	case "DESTINO":
		return "c.destino"
	case "ZONA":
		return "c.zona"
	case "COMPRADOR":
		return "c.compcuit"
	default:
		return "date_c(c.fecent)"
	}
}
