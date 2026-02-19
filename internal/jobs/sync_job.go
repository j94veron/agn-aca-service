package jobs

import (
	"context"
	"sort"
	"time"

	"agn-service/internal/domain"
	"agn-service/internal/repository"
	"agn-service/internal/utils"
)

type SyncJob struct {
	coRepo *repository.OracleRepo
	acRepo *repository.OracleRepo
	redis  *repository.RedisRepo

	ttl              time.Duration
	months           int
	proxWindowMonths int
}

func NewSyncJob(
	coRepo *repository.OracleRepo,
	acRepo *repository.OracleRepo,
	redis *repository.RedisRepo,
	ttl time.Duration,
	months int,
	proxWindowMonths int,
) *SyncJob {
	return &SyncJob{
		coRepo:           coRepo,
		acRepo:           acRepo,
		redis:            redis,
		ttl:              ttl,
		months:           months,
		proxWindowMonths: proxWindowMonths,
	}
}

func (j *SyncJob) Run(ctx context.Context) error {
	now := time.Now()

	base, err := j.buildDetail12M(ctx, now)
	if err != nil {
		return err
	}

	if err := j.buildSummaryByCUIT(ctx, now, base); err != nil {
		return err
	}
	if err := j.buildMonthly12M(ctx, now, base); err != nil {
		return err
	}
	if err := j.buildVencidos(ctx, now, base); err != nil {
		return err
	}

	return nil
}

func (j *SyncJob) buildDetail12M(ctx context.Context, now time.Time) (*domain.PendingFixDetailSnapshot, error) {

	from := utils.TruncDay(now)
	to := from.AddDate(0, j.months, 0)

	// 🔥 AHORA UNA SOLA LLAMADA
	var all []domain.PendingFixRow

	coRows, err := j.coRepo.FetchPendingFixAll(ctx, "CORRETAJE", "CO")
	if err != nil {
		return nil, err
	}
	all = append(all, coRows...)

	acRows, err := j.acRepo.FetchPendingFixAll(ctx, "ACOPIO", "AC")
	if err != nil {
		return nil, err
	}
	all = append(all, acRows...)

	snap := domain.PendingFixDetailSnapshot{
		GeneratedAt: now,
		FromDate:    from,
		ToDate:      to,
		Months:      j.months,
		Rows:        all,
	}

	if err := j.redis.SaveDetail12M(ctx, snap, j.ttl); err != nil {
		return nil, err
	}

	return &snap, nil
}

// Resumen por CUIT: toneladas + vencidas + null fecvtoent
func (j *SyncJob) buildSummaryByCUIT(ctx context.Context, now time.Time, base *domain.PendingFixDetailSnapshot) error {
	type key struct{ UniNego, CUIT string }

	m := make(map[key]*domain.PendingFixSummaryRow)
	today := utils.TruncDay(now)

	for _, r := range base.Rows {
		k := key{UniNego: r.UniNego, CUIT: r.CUIT}
		if k.CUIT == "" {
			continue
		}
		if m[k] == nil {
			m[k] = &domain.PendingFixSummaryRow{UniNego: r.UniNego, CUIT: r.CUIT}
		}

		m[k].TnPendientes += r.Pendientes

		// vencidas por fecfijahasta
		if r.FecHasta != nil && r.FecHasta.Before(today) {
			m[k].TnVencidas += r.Pendientes
		}
		// contratos sin vto ent
		if r.FecVtoEnt == nil {
			m[k].CtSinVtoEnt++
		}
	}

	out := make([]domain.PendingFixSummaryRow, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}

	// orden prolijo
	sort.Slice(out, func(i, j int) bool {
		if out[i].UniNego == out[j].UniNego {
			return out[i].CUIT < out[j].CUIT
		}
		return out[i].UniNego < out[j].UniNego
	})

	snap := domain.PendingFixSummarySnapshot{
		GeneratedAt: now,
		Rows:        out,
	}
	return j.redis.SaveSummary(ctx, snap, j.ttl)
}

// 12 meses por toneladas (siempre 12, con ceros)
func (j *SyncJob) buildMonthly12M(ctx context.Context, now time.Time, base *domain.PendingFixDetailSnapshot) error {
	fromMonth := utils.MonthStart(base.FromDate)

	// inicializa 12 meses
	rows := make([]domain.PendingFixMonthlyRow, 0, j.months)
	idx := map[string]int{}
	for i := 0; i < j.months; i++ {
		m := fromMonth.AddDate(0, i, 0)
		k := utils.MonthKey(m)
		idx[k] = i
		rows = append(rows, domain.PendingFixMonthlyRow{Month: k, Tn: 0})
	}

	// suma toneladas por mes usando fechasta
	for _, r := range base.Rows {
		if r.FecHasta == nil {
			continue
		}
		k := utils.MonthKey(*r.FecHasta)
		if pos, ok := idx[k]; ok {
			rows[pos].Tn += r.Pendientes
		}
	}

	snap := domain.PendingFixMonthlySnapshot{
		GeneratedAt: now,
		FromMonth:   utils.MonthKey(fromMonth),
		Months:      j.months,
		Rows:        rows,
	}
	return j.redis.SaveMonthly12M(ctx, snap, j.ttl)
}

// "Vencidos con vacíos”: por CUIT, ancla en primer vencimiento futuro y devuelve N meses (ej 3)
func (j *SyncJob) buildVencidos(ctx context.Context, now time.Time, base *domain.PendingFixDetailSnapshot) error {
	today := utils.TruncDay(now)

	// agrupamos por CUIT+UNINEGO con sus filas
	type key struct{ UniNego, CUIT string }
	group := map[key][]domain.PendingFixRow{}
	for _, r := range base.Rows {
		if r.CUIT == "" {
			continue
		}
		group[key{r.UniNego, r.CUIT}] = append(group[key{r.UniNego, r.CUIT}], r)
	}

	var out []domain.PendingFixVencidosRow

	for k, rows := range group {
		// buscamos el primer fechasta >= hoy (ancla)
		var anchor *time.Time
		for _, r := range rows {
			if r.FecHasta == nil {
				continue
			}
			if !r.FecHasta.Before(today) { // >= hoy
				if anchor == nil || r.FecHasta.Before(*anchor) {
					t := *r.FecHasta
					anchor = &t
				}
			}
		}

		// si no hay ancla, igual devolvemos ventana desde el mes actual (vacíos)
		var anchorMonth time.Time
		if anchor != nil {
			anchorMonth = utils.MonthStart(*anchor)
		} else {
			anchorMonth = utils.MonthStart(today)
		}

		// inicializa ventana N meses con ceros
		win := make([]domain.PendingFixVencidosMonth, 0, j.proxWindowMonths)
		idx := map[string]int{}
		for i := 0; i < j.proxWindowMonths; i++ {
			m := anchorMonth.AddDate(0, i, 0)
			keym := utils.MonthKey(m)
			idx[keym] = i
			win = append(win, domain.PendingFixVencidosMonth{Month: keym, Tn: 0})
		}

		// suma filas que caen en esos meses (por fechasta)
		for _, r := range rows {
			if r.FecHasta == nil {
				continue
			}
			mk := utils.MonthKey(*r.FecHasta)
			if pos, ok := idx[mk]; ok {
				win[pos].Tn += r.Pendientes
			}
		}

		out = append(out, domain.PendingFixVencidosRow{
			UniNego:     k.UniNego,
			CUIT:        k.CUIT,
			AnchorMonth: utils.MonthKey(anchorMonth),
			Months:      win,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].UniNego == out[j].UniNego {
			return out[i].CUIT < out[j].CUIT
		}
		return out[i].UniNego < out[j].UniNego
	})

	snap := domain.PendingFixVencidosSnapshot{
		GeneratedAt:  now,
		WindowMonths: j.proxWindowMonths,
		Rows:         out,
	}
	return j.redis.SaveVencidos(ctx, snap, j.ttl)
}
