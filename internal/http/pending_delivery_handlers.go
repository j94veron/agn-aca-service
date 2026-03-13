package http

import (
	"agn-service/internal/domain"
	"net/http"
	"time"

	"agn-service/internal/jobs"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PendingDeliveryHandlers struct {
	svc *service.PendingDeliveryService
	job *jobs.PendingDeliverySyncJob
}

func NewPendingDeliveryHandlers(
	svc *service.PendingDeliveryService,
	job *jobs.PendingDeliverySyncJob,
) *PendingDeliveryHandlers {
	return &PendingDeliveryHandlers{
		svc: svc,
		job: job,
	}
}

func (h *PendingDeliveryHandlers) GetList(c *gin.Context) {

	cuitVend := c.Query("cuit_vendedor")
	cuitComp := c.Query("cuit_comprador")
	grano := c.Query("grano")
	uninego := c.Query("uninego")

	desdeStr := c.Query("fecvto_desde")
	hastaStr := c.Query("fecvto_hasta")

	var desde *time.Time
	var hasta *time.Time

	layout := "2006-01-02" // formato ISO (recomendado)

	if desdeStr != "" {
		if t, err := time.Parse(layout, desdeStr); err == nil {
			desde = &t
		}
	}

	if hastaStr != "" {
		if t, err := time.Parse(layout, hastaStr); err == nil {
			hasta = &t
		}
	}

	resp, err := h.svc.Get(
		c.Request.Context(),
		uninego,
		cuitVend,
		cuitComp,
		grano,
		desde,
		hasta,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PendingDeliveryHandlers) SyncNow(c *gin.Context) {

	if err := h.job.Run(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PendingDeliveryHandlers) GetPendingDeliveryList(c *gin.Context) {

	f := domain.PendingDeliveryFilter{
		UniNego:       c.Query("uninego"),
		Segmento:      c.Query("segmento"),
		CuitVendedor:  c.Query("cuit_vendedor"),
		CuitComprador: c.Query("cuit_comprador"),
		VendCta:       c.Query("vendcta"),
		CompCta:       c.Query("compcta"),
		Contrato:      c.Query("contrato"),
		ContParte:     c.Query("contparte"),
		Grano:         c.Query("grano"),
		NomComp:       c.Query("nomcomp"),
	}

	f.FecEntDesde = parseDate(c.Query("fecent_desde"))
	f.FecEntHasta = parseDate(c.Query("fecent_hasta"))

	f.FecVtoDesde = parseDate(c.Query("fecvto_desde"))
	f.FecVtoHasta = parseDate(c.Query("fecvto_hasta"))

	resp, err := h.svc.GetList(c.Request.Context(), f)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *PendingDeliveryHandlers) GetVencidos(c *gin.Context) {

	f := domain.PendingDeliveryFilter{
		UniNego:       c.Query("uninego"),
		Segmento:      c.Query("segmento"),
		CuitVendedor:  c.Query("cuit_vendedor"),
		CuitComprador: c.Query("cuit_comprador"),
	}

	f.FecVtoDesde = parseDate(c.Query("fecvto_desde"))
	f.FecVtoHasta = parseDate(c.Query("fecvto_hasta"))

	resp, err := h.svc.GetVencidos(c.Request.Context(), f)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *PendingDeliveryHandlers) GetSummary(c *gin.Context) {

	f := domain.PendingDeliveryFilter{
		UniNego:       c.Query("uninego"),
		Segmento:      c.Query("segmento"),
		CuitVendedor:  c.Query("cuit_vendedor"),
		CuitComprador: c.Query("cuit_comprador"),
		VendCta:       c.Query("vendcta"),
		CompCta:       c.Query("compcta"),
		Contrato:      c.Query("contrato"),
		ContParte:     c.Query("contparte"),
		Grano:         c.Query("grano"),
		NomComp:       c.Query("nomcomp"),
	}

	f.FecEntDesde = parseDate(c.Query("fecent_desde"))
	f.FecEntHasta = parseDate(c.Query("fecent_hasta"))

	f.FecVtoDesde = parseDate(c.Query("fecvto_desde"))
	f.FecVtoHasta = parseDate(c.Query("fecvto_hasta"))

	resp, err := h.svc.GetSummary(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PendingDeliveryHandlers) GetMonthly(c *gin.Context) {

	f := domain.PendingDeliveryFilter{
		UniNego:       c.Query("uninego"),
		Segmento:      c.Query("segmento"),
		CuitVendedor:  c.Query("cuit_vendedor"),
		CuitComprador: c.Query("cuit_comprador"),
		VendCta:       c.Query("vendcta"),
		CompCta:       c.Query("compcta"),
		Contrato:      c.Query("contrato"),
		ContParte:     c.Query("contparte"),
		Grano:         c.Query("grano"),
		NomComp:       c.Query("nomcomp"),
	}

	f.FecEntDesde = parseDate(c.Query("fecent_desde"))
	f.FecEntHasta = parseDate(c.Query("fecent_hasta"))

	f.FecVtoDesde = parseDate(c.Query("fecvto_desde"))
	f.FecVtoHasta = parseDate(c.Query("fecvto_hasta"))

	resp, err := h.svc.GetMonthly(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
