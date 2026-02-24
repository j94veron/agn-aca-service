package repository

import (
	"agn-service/internal/domain"
	"context"
)

const (
	KeyPendingDelivery = "agn:stats:v1:pending_delivery:snapshot"
)

func (r *RedisRepo) LoadPendingDelivery(ctx context.Context) (*domain.PendingDeliverySnapshot, error) {
	var snap domain.PendingDeliverySnapshot
	if err := getJSON(ctx, r.rdb, KeyPendingDelivery, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}
