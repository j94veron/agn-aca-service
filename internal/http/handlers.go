package http

import (
	"agn-service/internal/domain"
	"net/http"
	"strconv"
	"time"

	"agn-service/internal/jobs"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

func parseFloat(c *gin.Context, key string) float64 {
	val := c.Query(key)
	if val == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func buildFilters(c *gin.Context) domain.PendingFixFilters {
	return domain.PendingFixFilters{
		UniNego:    c.Query("uninego"),
		CUIT:       c.Query("cuit"),
		Segmento:   c.Query("segmento"),
		VendCta:    c.Query("vendcta"),
		CompCta:    c.Query("compcta"),
		Contrato:   c.Query("contrato"),
		ContParte:  c.Query("contparte"),
		CompNombre: c.Query("compnombre"),

		MinPendientes: parseFloat(c, "pendientes_min"),
		MinPendApli:   parseFloat(c, "pendapli_min"),

		//NUEVOS NOMBRES CLAROS (con alias viejo para compatibilidad)

		FecEntDesde: parseDateAlias(c,
			"entrega_desde",
			"fecent_desde",
		),
		FecEntHasta: parseDateAlias(c,
			"entrega_hasta",
			"fecent_hasta",
		),

		FecDesdeDesde: parseDateAlias(c,
			"fijacion_inicio_desde",
			"fecdesde_desde",
		),
		FecDesdeHasta: parseDateAlias(c,
			"fijacion_inicio_hasta",
			"fecdesde_hasta",
		),

		FecHastaDesde: parseDateAlias(c,
			"fijacion_fin_desde",
			"fechasta_desde",
		),
		FecHastaHasta: parseDateAlias(c,
			"fijacion_fin_hasta",
			"fechasta_hasta",
		),

		FecVtoEntDesde: parseDateAlias(c,
			"vencimiento_entrega_desde",
			"fecvto_desde",
		),
		FecVtoEntHasta: parseDateAlias(c,
			"vencimiento_entrega_hasta",
			"fecvto_hasta",
		),
	}
}

func parseDateAlias(c *gin.Context, keys ...string) *time.Time {
	for _, k := range keys {
		val := c.Query(k)
		if val != "" {
			t, err := time.Parse("2006-01-02", val)
			if err == nil {
				return &t
			}
		}
	}
	return nil
}

type Handlers struct {
	svc *service.PendingFixService
	job *jobs.SyncJob
}

func NewHandlers(svc *service.PendingFixService, job *jobs.SyncJob) *Handlers {
	return &Handlers{svc: svc, job: job}
}

func (h *Handlers) GetDetail(c *gin.Context) {
	filters := buildFilters(c)

	snap, err := h.svc.GetDetail(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetSummary(c *gin.Context) {
	filters := buildFilters(c)

	snap, err := h.svc.GetSummary(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetMonthly(c *gin.Context) {
	filters := buildFilters(c)

	snap, err := h.svc.GetMonthly12M(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetVencidos(c *gin.Context) {
	filters := buildFilters(c)

	snap, err := h.svc.GetVencidos(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) SyncNow(c *gin.Context) {
	if err := h.job.Run(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handlers) GetVencidosV2(c *gin.Context) {
	filters := buildFilters(c)

	snap, err := h.svc.GetVencidosV2(c.Request.Context(), filters, 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}
