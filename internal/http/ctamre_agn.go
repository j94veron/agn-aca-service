package http

import (
	"net/http"

	"agn-service/internal/domain"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

func CreateCtamre(svc *service.CtamreService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req domain.CtamreRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := svc.Create(c.Request.Context(), req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, domain.ApiResponse{
			Status: "OK",
		})
	}
}
