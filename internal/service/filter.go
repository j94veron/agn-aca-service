package service

import "agn-service/internal/domain"

type DeliveryFilter func(domain.PendingDeliveryRow) bool

func applyFilters(
	rows []domain.PendingDeliveryRow,
	filters []DeliveryFilter,
) []domain.PendingDeliveryRow {

	if len(filters) == 0 {
		return rows
	}

	out := make([]domain.PendingDeliveryRow, 0, len(rows))

	for _, r := range rows {

		ok := true
		for _, f := range filters {
			if !f(r) {
				ok = false
				break
			}
		}

		if ok {
			out = append(out, r)
		}
	}

	return out
}
