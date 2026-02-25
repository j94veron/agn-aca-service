package http

import (
	"net/http"

	"agn-service/internal/jobs"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	svc *service.PendingFixService
	job *jobs.SyncJob
}

func NewHandlers(svc *service.PendingFixService, job *jobs.SyncJob) *Handlers {
	return &Handlers{svc: svc, job: job}
}

func (h *Handlers) GetDetail(c *gin.Context) {
	uninego := c.Query("uninego")
	cuit := c.Query("cuit")
	snap, err := h.svc.GetDetail(c.Request.Context(), uninego, cuit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetSummary(c *gin.Context) {
	uninego := c.Query("uninego")
	cuit := c.Query("cuit")
	snap, err := h.svc.GetSummary(c.Request.Context(), uninego, cuit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetMonthly(c *gin.Context) {
	uninego := c.Query("uninego")
	cuit := c.Query("cuit")

	snap, err := h.svc.GetMonthly12M(c.Request.Context(), uninego, cuit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, snap)
}

func (h *Handlers) GetVencidos(c *gin.Context) {
	uninego := c.Query("uninego")
	cuit := c.Query("cuit")
	snap, err := h.svc.GetVencidos(c.Request.Context(), uninego, cuit)
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
	uninego := c.Query("uninego")
	cuit := c.Query("cuit")

	// ventana fija 12
	snap, err := h.svc.GetVencidosV2(c.Request.Context(), uninego, cuit, 12)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, snap)
}
