package http

import (
	"agn-service/internal/auth"
	"agn-service/internal/jobs"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	fixSvc *service.PendingFixService,
	deliverySvc *service.PendingDeliveryService,
	fixJob *jobs.SyncJob,
	deliveryJob *jobs.PendingDeliverySyncJob,
) *gin.Engine {

	r := gin.Default()

	h := NewHandlers(fixSvc, fixJob)
	dh := NewPendingDeliveryHandlers(deliverySvc, deliveryJob)

	// =============================
	// 🔒 Pending Fix
	// =============================
	api := r.Group("/api/v1/pending-fix", auth.MiddlewareJWTGin())
	{
		api.GET("/detail", h.GetDetail)
		api.GET("/summary", h.GetSummary)
		api.GET("/monthly", h.GetMonthly)
		api.GET("/vencidos", h.GetVencidos)
		api.GET("/vencidos/v2", h.GetVencidosV2)
	}

	internal := r.Group("/api/v1/pending-fix/internal")
	{
		internal.POST("/sync", h.SyncNow)
	}

	// =============================
	// 🚀 Pending Delivery NUEVO
	// =============================
	delivery := r.Group("/api/v1/pending-delivery", auth.MiddlewareJWTGin())
	{
		delivery.GET("/list", dh.GetList)
	}

	internalDelivery := r.Group("/api/v1/pending-delivery/internal")
	{
		internalDelivery.POST("/sync", dh.SyncNow)
	}

	return r
}
