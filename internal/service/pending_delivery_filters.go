package service

import (
	"agn-service/internal/domain"
	"time"
)

func matchPendingDelivery(r domain.PendingDeliveryRow, f domain.PendingDeliveryFilter) bool {

	if f.UniNego != "" && r.UniNego != f.UniNego {
		return false
	}

	if f.Segmento != "" && r.Segmento != f.Segmento {
		return false
	}

	if f.CuitVendedor != "" && r.CUITVendedor != f.CuitVendedor {
		return false
	}

	if f.CuitComprador != "" && r.CUITComprador != f.CuitComprador {
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

	if f.Grano != "" && r.Grano != f.Grano {
		return false
	}

	if f.NomComp != "" && r.NomComp != f.NomComp {
		return false
	}

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

	if f.FecVtoDesde != nil || f.FecVtoHasta != nil {
		if !inRange(r.FecVtoEnt, f.FecVtoDesde, f.FecVtoHasta) {
			return false
		}

	}

	return true
}
