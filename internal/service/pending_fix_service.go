package service

import (
	"context"
	"strings"

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
