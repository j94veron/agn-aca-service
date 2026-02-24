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
