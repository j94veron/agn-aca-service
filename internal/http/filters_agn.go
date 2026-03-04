package http

import (
	"agn-service/internal/domain"
	"agn-service/internal/service"
	"net/http"
	"strconv"
)

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseOliquiFilter(r *http.Request) (*domain.OliquiFilter, error) {
	q := r.URL.Query()

	uni := q.Get("uniNego")
	if err := service.ValidateUniNego(uni); err != nil {
		return nil, err
	}

	schema := service.ResolveSchemaByUniNego(uni)

	f := &domain.OliquiFilter{
		UniNego:  uni,
		Schema:   schema,
		Page:     atoi(q.Get("page"), 1),
		PageSize: atoi(q.Get("pageSize"), 10),
		OrderBy:  q.Get("orderBy"),
		OrderDir: q.Get("orderDir"),
	}
	if v := q.Get("contInterno"); v != "" {
		x, _ := strconv.ParseInt(v, 10, 64)
		f.ContInterno = &x
	}
	if v := q.Get("cosecha"); v != "" {
		f.Cosecha = &v
	}

	if v := q.Get("page"); v != "" {
		f.Page, _ = strconv.Atoi(v)
	}
	if v := q.Get("pageSize"); v != "" {
		f.PageSize, _ = strconv.Atoi(v)
	}

	return f, nil
}
