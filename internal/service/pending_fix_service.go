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
	// Reusamos el snapshot de detalle ya guardado en Redis
	snap, err := s.redis.LoadDetail12M(ctx)
	if err != nil {
		return nil, err
	}

	uninego = strings.TrimSpace(strings.ToUpper(uninego))
	cuit = strings.TrimSpace(cuit)

	now := time.Now()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Prearmar lista fija de 12 meses (o windowMonths)
	months := make([]string, 0, windowMonths)
	for i := 0; i < windowMonths; i++ {
		m := startMonth.AddDate(0, i, 0).Format("2006-01")
		months = append(months, m)
	}

	// agg[cuit][uninego][month][grano|cosecha] = tn
	type keyCU struct{ cuit, uninego string }
	agg := make(map[keyCU]map[string]map[string]float64)

	for _, r := range snap.Rows {
		// filtros opcionales
		if uninego != "" && r.UniNego != uninego {
			continue
		}
		if cuit != "" && r.CUIT != cuit {
			continue
		}

		// criterio "vencidos": fecHasta < hoy  OR fecHasta nil (si querés incluir NULL como “sin fecha”)
		// Para V2 lo pedido es “vencimientos 12 meses” → normalmente se usa FecHasta como “fecha de vencimiento”
		// Si r.FecHasta es nil: lo podés ignorar o mandarlo a un bucket especial.
		if r.FecHasta == nil {
			continue
		}

		month := r.FecHasta.Format("2006-01")

		// Solo meses dentro de la ventana (12 meses desde hoy)
		inWindow := false
		for _, m := range months {
			if m == month {
				inWindow = true
				break
			}
		}
		if !inWindow {
			continue
		}

		k := keyCU{cuit: r.CUIT, uninego: r.UniNego}
		if _, ok := agg[k]; !ok {
			agg[k] = make(map[string]map[string]float64)
		}
		if _, ok := agg[k][month]; !ok {
			agg[k][month] = make(map[string]float64)
		}

		gk := r.Grano + "|" + r.Cosecha
		agg[k][month][gk] += r.Pendientes // toneladas
	}

	// Convertir agg → snapshot
	out := &domain.PendingFixVencidosV2Snapshot{
		GeneratedAt:  snap.GeneratedAt,
		WindowMonths: windowMonths,
		Rows:         make([]domain.PendingFixVencidosV2Row, 0, len(agg)),
	}

	for k, byMonth := range agg {
		row := domain.PendingFixVencidosV2Row{
			UniNego: k.uninego,
			CUIT:    k.cuit,
			Months:  make([]domain.PendingFixVencidosV2Month, 0, windowMonths),
		}

		for _, m := range months {
			bucket := byMonth[m] // map[grano|cosecha]tn
			granos := make([]domain.PendingFixVencidosV2Grano, 0, len(bucket))
			for gk, tn := range bucket {
				parts := strings.SplitN(gk, "|", 2)
				gr := ""
				co := ""
				if len(parts) > 0 {
					gr = parts[0]
				}
				if len(parts) == 2 {
					co = parts[1]
				}
				granos = append(granos, domain.PendingFixVencidosV2Grano{
					Grano:   gr,
					Cosecha: co,
					Tn:      tn,
				})
			}

			row.Months = append(row.Months, domain.PendingFixVencidosV2Month{
				Month:  m,
				Granos: granos, // puede estar vacío
			})
		}

		out.Rows = append(out.Rows, row)
	}

	return out, nil
}
