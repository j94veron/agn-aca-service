package service

import (
	"agn-service/internal/utils"
	"context"
	"strings"
	"time"

	"agn-service/internal/domain"
	"agn-service/internal/repository"
)

type PendingFixService struct {
	redis *repository.RedisRepo
}

func NewPendingFixService(redis *repository.RedisRepo) *PendingFixService {
	return &PendingFixService{redis: redis}
}

// 1) Detalle
func (s *PendingFixService) GetDetail(
	ctx context.Context,
	filters domain.PendingFixFilters,
) (*domain.PendingFixDetailSnapshot, error) {

	snap, err := s.redis.LoadDetail12M(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]domain.PendingFixRow, 0, len(snap.Rows))

	for _, r := range snap.Rows {
		if matchPendingFix(r, filters) {
			out = append(out, r)
		}
	}

	snap.Rows = out
	return snap, nil
}

// 2) Summary por CUIT
func (s *PendingFixService) GetSummary(
	ctx context.Context,
	filters domain.PendingFixFilters,
) (*domain.PendingFixSummarySnapshot, error) {

	detail, err := s.GetDetail(ctx, filters)
	if err != nil {
		return nil, err
	}

	type key struct{ UniNego, CUIT string }
	m := make(map[key]*domain.PendingFixSummaryRow)
	today := time.Now()

	for _, r := range detail.Rows {
		k := key{r.UniNego, r.CUIT}
		if m[k] == nil {
			m[k] = &domain.PendingFixSummaryRow{
				UniNego: r.UniNego,
				CUIT:    r.CUIT,
			}
		}

		m[k].TnPendientes += r.Pendientes

		if r.FecHasta != nil && r.FecHasta.Before(today) {
			m[k].TnVencidas += r.Pendientes
		}

		if r.FecVtoEnt == nil {
			m[k].CtSinVtoEnt++
		}
	}

	out := make([]domain.PendingFixSummaryRow, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}

	return &domain.PendingFixSummarySnapshot{
		GeneratedAt: detail.GeneratedAt,
		Rows:        out,
	}, nil
}

// 3) Monthly 12M (toneladas) - ahora con filtros opcionales
func (s *PendingFixService) GetMonthly12M(
	ctx context.Context,
	filters domain.PendingFixFilters,
) (*domain.PendingFixMonthlySnapshot, error) {

	detail, err := s.GetDetail(ctx, filters)
	if err != nil {
		return nil, err
	}

	fromMonth := utils.MonthStart(detail.FromDate)

	rows := make([]domain.PendingFixMonthlyRow, 0, detail.Months)
	idx := map[string]int{}

	for i := 0; i < detail.Months; i++ {
		m := fromMonth.AddDate(0, i, 0)
		k := utils.MonthKey(m)
		idx[k] = i
		rows = append(rows, domain.PendingFixMonthlyRow{Month: k})
	}

	for _, r := range detail.Rows {
		if r.FecHasta == nil {
			continue
		}
		k := utils.MonthKey(*r.FecHasta)
		if pos, ok := idx[k]; ok {
			rows[pos].Tn += r.Pendientes
		}
	}

	return &domain.PendingFixMonthlySnapshot{
		GeneratedAt: detail.GeneratedAt,
		FromMonth:   utils.MonthKey(fromMonth),
		Months:      detail.Months,
		Rows:        rows,
	}, nil
}

// 4) Vencidos con vacíos (por CUIT)
func (s *PendingFixService) GetVencidos(
	ctx context.Context,
	filters domain.PendingFixFilters,
) (*domain.PendingFixVencidosSnapshot, error) {

	detail, err := s.GetDetail(ctx, filters)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type key struct{ UniNego, CUIT string }

	group := map[key][]domain.PendingFixRow{}

	for _, r := range detail.Rows {
		k := key{r.UniNego, r.CUIT}
		group[k] = append(group[k], r)
	}

	var out []domain.PendingFixVencidosRow

	for k, rows := range group {

		anchorMonth := todayMonth.Format("2006-01")

		months := make([]domain.PendingFixVencidosMonth, 0)

		monthAgg := make(map[string]float64)

		for _, r := range rows {

			if r.FecHasta == nil {
				continue
			}

			month := r.FecHasta.Format("2006-01")

			if r.FecHasta.Before(now) {
				month = "VENCIDO"
			}

			monthAgg[month] += r.Pendientes
		}

		for m, tn := range monthAgg {
			months = append(months, domain.PendingFixVencidosMonth{
				Month: m,
				Tn:    tn,
			})
		}

		out = append(out, domain.PendingFixVencidosRow{
			UniNego:     k.UniNego,
			CUIT:        k.CUIT,
			AnchorMonth: anchorMonth,
			Months:      months,
		})
	}

	return &domain.PendingFixVencidosSnapshot{
		GeneratedAt: detail.GeneratedAt,
		Rows:        out,
	}, nil
}

func (s *PendingFixService) GetVencidosV2(
	ctx context.Context,
	filters domain.PendingFixFilters,
	windowMonths int,
) (*domain.PendingFixVencidosV2Snapshot, error) {

	detail, err := s.GetDetail(ctx, filters)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	months := make([]string, 0, windowMonths+2)
	monthSet := make(map[string]struct{}, windowMonths+2)

	for i := 0; i < windowMonths; i++ {
		m := startMonth.AddDate(0, i, 0).Format("2006-01")
		months = append(months, m)
		monthSet[m] = struct{}{}
	}

	months = append(months, "SIN_FECHA")
	months = append(months, "VENCIDO")

	monthSet["SIN_FECHA"] = struct{}{}
	monthSet["VENCIDO"] = struct{}{}

	type keyCU struct{ cuit, uninego string }

	type aggItem struct {
		Grano    string
		Cosecha  string
		NomGrano string
		Tn       float64
	}

	agg := make(map[keyCU]map[string]map[string]*aggItem)

	for _, r := range detail.Rows {

		month := "SIN_FECHA"

		if r.FecHasta != nil {

			if r.FecHasta.Before(now) {
				month = "VENCIDO"
			} else {
				month = r.FecHasta.Format("2006-01")
				if _, ok := monthSet[month]; !ok {
					continue
				}
			}
		}

		k := keyCU{cuit: r.CUIT, uninego: r.UniNego}

		if _, ok := agg[k]; !ok {
			agg[k] = make(map[string]map[string]*aggItem)
		}

		if _, ok := agg[k][month]; !ok {
			agg[k][month] = make(map[string]*aggItem)
		}

		gk := r.Grano + "|" + r.Cosecha

		if _, ok := agg[k][month][gk]; !ok {
			agg[k][month][gk] = &aggItem{
				Grano:    r.Grano,
				Cosecha:  r.Cosecha,
				NomGrano: r.NomGrano,
			}
		}

		agg[k][month][gk].Tn += r.Pendientes
	}

	out := &domain.PendingFixVencidosV2Snapshot{
		GeneratedAt:  detail.GeneratedAt,
		WindowMonths: windowMonths,
		Rows:         make([]domain.PendingFixVencidosV2Row, 0, len(agg)),
	}

	for k, byMonth := range agg {

		row := domain.PendingFixVencidosV2Row{
			UniNego: k.uninego,
			CUIT:    k.cuit,
			Months:  make([]domain.PendingFixVencidosV2Month, 0, len(months)),
		}

		for _, m := range months {

			bucket := byMonth[m]
			granos := make([]domain.PendingFixVencidosV2Grano, 0, len(bucket))

			for _, it := range bucket {
				granos = append(granos, domain.PendingFixVencidosV2Grano{
					Grano:   it.Grano,
					Cosecha: it.Cosecha,
					Nombre:  it.NomGrano,
					Tn:      it.Tn,
				})
			}

			row.Months = append(row.Months, domain.PendingFixVencidosV2Month{
				Month:  m,
				Granos: granos,
			})
		}

		out.Rows = append(out.Rows, row)
	}

	return out, nil
}

func matchPendingFix(r domain.PendingFixRow, f domain.PendingFixFilters) bool {

	if f.UniNego != "" && r.UniNego != f.UniNego {
		return false
	}

	if f.CUIT != "" && r.CUIT != f.CUIT {
		return false
	}

	if f.Segmento != "" && r.Segmento != f.Segmento {
		return false
	}

	if f.VendCta != "" && r.VendCta != f.VendCta {
		return false
	}

	if f.CompCta != "" && r.CompCta != f.CompCta {
		return false
	}

	if f.Contrato != "" && r.Contrato != f.Contrato {
		return false
	}

	if f.ContParte != "" && r.ContParte != f.ContParte {
		return false
	}

	if f.CompNombre != "" && !strings.Contains(strings.ToUpper(r.CompNombre), strings.ToUpper(f.CompNombre)) {
		return false
	}

	if f.MinPendientes > 0 && r.Pendientes < f.MinPendientes {
		return false
	}

	if f.MinPendApli > 0 && r.PendApli < f.MinPendApli {
		return false
	}

	// Fechas helper
	inRange := func(date *time.Time, desde, hasta *time.Time) bool {
		if date == nil {
			return false
		}
		if desde != nil && date.Before(*desde) {
			return false
		}
		if hasta != nil && date.After(*hasta) {
			return false
		}
		return true
	}

	if f.FecEntDesde != nil || f.FecEntHasta != nil {
		if !inRange(r.FecEnt, f.FecEntDesde, f.FecEntHasta) {
			return false
		}
	}

	if f.FecDesdeDesde != nil || f.FecDesdeHasta != nil {
		if !inRange(r.FecDesde, f.FecDesdeDesde, f.FecDesdeHasta) {
			return false
		}
	}

	if f.FecHastaDesde != nil || f.FecHastaHasta != nil {
		if !inRange(r.FecHasta, f.FecHastaDesde, f.FecHastaHasta) {
			return false
		}
	}

	if f.FecVtoEntDesde != nil || f.FecVtoEntHasta != nil {
		if !inRange(r.FecVtoEnt, f.FecVtoEntDesde, f.FecVtoEntHasta) {
			return false
		}
	}

	return true
}
