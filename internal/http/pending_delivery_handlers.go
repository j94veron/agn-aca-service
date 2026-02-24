package http

import (
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
