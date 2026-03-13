package service

import (
	"agn-service/internal/domain"
	"agn-service/internal/repository"
	"context"
	"time"
)

type PendingDeliveryService struct {
	redis *repository.RedisRepo
}

func NewPendingDeliveryService(redis *repository.RedisRepo) *PendingDeliveryService {
	return &PendingDeliveryService{redis: redis}
}

func (s *PendingDeliveryService) Get(
	ctx context.Context,
	uninego string,
	cuitVend string,
	cuitComp string,
	grano string,
	desde *time.Time,
	hasta *time.Time,
) ([]domain.PendingDeliveryRow, error) {

	snap, err := s.redis.GetPendingDelivery(ctx)
	if err != nil {
		return nil, err
	}

	if snap == nil {
		return []domain.PendingDeliveryRow{}, nil
	}

	filters := make([]DeliveryFilter, 0)

	// ========================
	// 🔥 FILTROS DINÁMICOS
	// ========================

	if uninego != "" {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.UniNego == uninego
		})
	}

	if cuitVend != "" {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.CUITVendedor == cuitVend
		})
	}

	if cuitComp != "" {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.CUITComprador == cuitComp
		})
	}

	if grano != "" {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.Grano == grano
		})
	}

	if desde != nil {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.FecVtoEnt != nil && !r.FecVtoEnt.Before(*desde)
		})
	}

	if hasta != nil {
		filters = append(filters, func(r domain.PendingDeliveryRow) bool {
			return r.FecVtoEnt != nil && !r.FecVtoEnt.After(*hasta)
		})
	}

	return applyFilters(snap.Rows, filters), nil
}

func (s *PendingDeliveryService) GetList(
	ctx context.Context,
	f domain.PendingDeliveryFilter,
) (*domain.PendingDeliverySnapshot, error) {

	snap, err := s.redis.LoadPendingDelivery(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]domain.PendingDeliveryRow, 0, len(snap.Rows))

	for _, r := range snap.Rows {

		if !matchPendingDelivery(r, f) {
			continue
		}

		out = append(out, r)
	}

	snap.Rows = out
	return snap, nil
}

func (s *PendingDeliveryService) GetSummary(
	ctx context.Context,
	f domain.PendingDeliveryFilter,
) (*domain.PendingDeliverySummarySnapshot, error) {

	snap, err := s.redis.LoadPendingDelivery(ctx)
	if err != nil {
		return nil, err
	}

	type key struct {
		uninego string
		cuit    string
	}

	agg := map[key]float64{}

	for _, r := range snap.Rows {

		if !matchPendingDelivery(r, f) {
			continue
		}

		k := key{r.UniNego, r.CUITVendedor}

		agg[k] += r.KilPenApli
	}

	rows := make([]domain.PendingDeliverySummaryRow, 0, len(agg))

	for k, tn := range agg {

		rows = append(rows, domain.PendingDeliverySummaryRow{
			UniNego: k.uninego,
			Cuit:    k.cuit,
			Tn:      tn,
		})
	}

	return &domain.PendingDeliverySummarySnapshot{
		GeneratedAt: snap.GeneratedAt,
		Rows:        rows,
	}, nil
}

func (s *PendingDeliveryService) GetMonthly(
	ctx context.Context,
	f domain.PendingDeliveryFilter,
) (*domain.PendingDeliveryMonthlySnapshot, error) {

	snap, err := s.redis.LoadPendingDelivery(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	fromMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	rows := make(map[string]float64)

	for i := 0; i < 12; i++ {

		m := fromMonth.AddDate(0, i, 0).Format("2006-01")
		rows[m] = 0
	}

	for _, r := range snap.Rows {

		if !matchPendingDelivery(r, f) {
			continue
		}

		if r.FecVtoEnt == nil {
			continue
		}

		m := r.FecVtoEnt.Format("2006-01")

		if _, ok := rows[m]; ok {
			rows[m] += r.KilPenLiq
		}
	}

	out := make([]domain.PendingDeliveryMonthlyRow, 0)

	for m, tn := range rows {

		out = append(out, domain.PendingDeliveryMonthlyRow{
			Month: m,
			Tn:    tn,
		})
	}

	return &domain.PendingDeliveryMonthlySnapshot{
		GeneratedAt: snap.GeneratedAt,
		Rows:        out,
	}, nil
}

func (s *PendingDeliveryService) GetVencidos(
	ctx context.Context,
	f domain.PendingDeliveryFilter,
) (*domain.PendingDeliveryMonthlySnapshot, error) {

	snap, err := s.redis.LoadPendingDelivery(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	rows := map[string]float64{
		"VENCIDO":   0,
		"SIN_FECHA": 0,
	}

	for _, r := range snap.Rows {

		if !matchPendingDelivery(r, f) {
			continue
		}

		if r.FecVtoEnt == nil {

			rows["SIN_FECHA"] += r.KilPenLiq
			continue
		}

		if r.FecVtoEnt.Before(now) {

			rows["VENCIDO"] += r.KilPenLiq
		}
	}

	out := make([]domain.PendingDeliveryMonthlyRow, 0)

	for m, tn := range rows {

		out = append(out, domain.PendingDeliveryMonthlyRow{
			Month: m,
			Tn:    tn,
		})
	}

	return &domain.PendingDeliveryMonthlySnapshot{
		GeneratedAt: snap.GeneratedAt,
		Rows:        out,
	}, nil
}
