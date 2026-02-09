package http

import (
	"agn-service/internal/auth"
	"agn-service/internal/jobs"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(svc *service.PendingFixService, job *jobs.SyncJob) *gin.Engine {
	r := gin.Default()
	h := NewHandlers(svc, job)

	// 🔒 Grupo protegido por JWT
	api := r.Group("/api/v1/pending-fix", auth.MiddlewareJWTGin())
	{
		api.GET("/detail", h.GetDetail)
		api.GET("/summary", h.GetSummary)
		api.GET("/monthly", h.GetMonthly)
		api.GET("/vencidos", h.GetVencidos)
	}

	// 🔐 Endpoint interno (si querés, podés dejarlo SIN JWT o con otro middleware)
	internal := r.Group("/api/v1/pending-fix/internal")
	{
		internal.POST("/sync", h.SyncNow)
	}

	return r
}
