package http

import (
	"log"
	"net/http"

	"agn-service/internal/domain"
	"agn-service/internal/service"

	"github.com/gin-gonic/gin"
)

func CreateOliqui(svc *service.OliquiService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req domain.OliqUiRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.Create(c.Request.Context(), req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, domain.ApiResponse{
			Status: "OK",
		})
	}
}

func GetOliquiTooltip(svc *service.OliquiService) gin.HandlerFunc {
	return func(c *gin.Context) {

		f, err := parseOliquiFilter(c.Request)
		if err != nil {
			log.Printf("[ERROR] tooltip parse filters: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		items, total, err := svc.GetTooltip(c.Request.Context(), *f)
		if err != nil {
			log.Printf("[ERROR] tooltip query: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"page":     f.Page,
			"pageSize": f.PageSize,
			"total":    total,
			"items":    items,
		})
	}
}

func GetOliquiGrid(svc *service.OliquiService) gin.HandlerFunc {
	return func(c *gin.Context) {

		f, err := parseOliquiFilter(c.Request)
		if err != nil {
			log.Printf("[ERROR] grid parse filters: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		items, total, err := svc.GetGrid(c.Request.Context(), *f)
		if err != nil {
			log.Printf("[ERROR] grid query: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"page":     f.Page,
			"pageSize": f.PageSize,
			"total":    total,
			"items":    items,
		})
	}
}
