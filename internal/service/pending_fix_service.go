package service

import (
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
func (s *PendingFixService) GetDetail(ctx context.Context, uninego, cuit string) (*domain.PendingFixDetailSnapshot, error) {
	snap, err := s.redis.LoadDetail12M(ctx)
	if err != nil {
		return nil, err
	}
	uninego = strings.TrimSpace(strings.ToUpper(uninego))
	cuit = strings.TrimSpace(cuit)

	if uninego == "" && cuit == "" {
		return snap, nil
	}

	out := make([]domain.PendingFixRow, 0, len(snap.Rows))
	for _, r := range snap.Rows {
		if uninego != "" && r.UniNego != uninego {
			continue
		}
		if cuit != "" && r.CUIT != cuit {
			continue
		}
		out = append(out, r)
	}
	snap.Rows = out
	return snap, nil
}

// 2) Summary por CUIT
func (s *PendingFixService) GetSummary(ctx context.Context, uninego, cuit string) (*domain.PendingFixSummarySnapshot, error) {
	snap, err := s.redis.LoadSummary(ctx)
	if err != nil {
		return nil, err
	}
	uninego = strings.TrimSpace(strings.ToUpper(uninego))
	cuit = strings.TrimSpace(cuit)
	if uninego == "" && cuit == "" {
		return snap, nil
	}

	out := make([]domain.PendingFixSummaryRow, 0, len(snap.Rows))
	for _, r := range snap.Rows {
		if uninego != "" && r.UniNego != uninego {
			continue
		}
		if cuit != "" && r.CUIT != cuit {
			continue
		}
		out = append(out, r)
	}
	snap.Rows = out
	return snap, nil
}

// 3) Monthly 12M (toneladas)
func (s *PendingFixService) GetMonthly12M(ctx context.Context) (*domain.PendingFixMonthlySnapshot, error) {
	return s.redis.LoadMonthly12M(ctx)
}

// 4) Vencidos con vacíos (por CUIT)
func (s *PendingFixService) GetVencidos(ctx context.Context, uninego, cuit string) (*domain.PendingFixVencidosSnapshot, error) {
	snap, err := s.redis.LoadVencidos(ctx)
	if err != nil {
		return nil, err
	}

	uninego = strings.TrimSpace(strings.ToUpper(uninego))
	cuit = strings.TrimSpace(cuit)
	if uninego == "" && cuit == "" {
		return snap, nil
	}

	out := make([]domain.PendingFixVencidosRow, 0, len(snap.Rows))
	for _, r := range snap.Rows {
		if uninego != "" && r.UniNego != uninego {
			continue
		}
		if cuit != "" && r.CUIT != cuit {
			continue
		}
		out = append(out, r)
	}
	snap.Rows = out
	return snap, nil
}

func (s *PendingFixService) GetVencidosV2(ctx context.Context, uninego, cuit string, windowMonths int) (*domain.PendingFixVencidosV2Snapshot, error) {

	snap, err := s.redis.LoadDetail12M(ctx)
	if err != nil {
		return nil, err
	}

	uninego = strings.TrimSpace(strings.ToUpper(uninego))
	cuit = strings.TrimSpace(cuit)

	now := time.Now()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// ========= meses =========
	months := make([]string, 0, windowMonths+1)
	for i := 0; i < windowMonths; i++ {
		m := startMonth.AddDate(0, i, 0).Format("2006-01")
		months = append(months, m)
	}
	months = append(months, "SIN_FECHA")

	// hash lookup ultra rápido
	monthSet := make(map[string]struct{}, len(months))
	for _, m := range months {
		monthSet[m] = struct{}{}
	}

	// ========= agg =========
	type keyCU struct{ cuit, uninego string }

	type aggItem struct {
		Grano    string
		Cosecha  string
		NomGrano string
		Tn       float64
	}

	// agg[cuit+uninego][month][grano|cosecha]
	agg := make(map[keyCU]map[string]map[string]*aggItem)

	for _, r := range snap.Rows {

		if uninego != "" && r.UniNego != uninego {
			continue
		}
		if cuit != "" && r.CUIT != cuit {
			continue
		}

		month := "SIN_FECHA"

		if r.FecHasta != nil {
			month = r.FecHasta.Format("2006-01")

			if _, ok := monthSet[month]; !ok {
				continue
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
				Tn:       0,
			}
		}

		agg[k][month][gk].Tn += r.Pendientes
	}

	// ========= build response =========

	out := &domain.PendingFixVencidosV2Snapshot{
		GeneratedAt:  snap.GeneratedAt,
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
