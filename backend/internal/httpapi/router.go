package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var publicEndpoints = []string{
	"GET /",
	"GET /health",
	"GET /api/pokemon/search?q={query}&limit={n}",
	"GET /api/pokemon/:id/evolution-options?q={query}",
	"GET /api/sessions/:id",
	"GET /api/sessions/:id/history",
	"GET /api/analytics/pokemon-usage?limit={n}",
	"GET /api/analytics/damage",
	"GET /api/analytics/knockouts",
	"GET /api/analytics/visits/summary",
	"GET /api/analytics/visits/:visitorId",
}

func endpointIndexResponse() gin.H {
	return gin.H{
		"message":   "You are probably looking for these endpoints.",
		"endpoints": publicEndpoints,
	}
}

func SetupRouter(h *Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, endpointIndexResponse())
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "route not found",
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"suggestion": "You are probably looking for these endpoints.",
			"endpoints":  publicEndpoints,
		})
	})

	r.GET("/health", h.Health)
	r.GET("/api/pokemon/search", h.SearchPokemon)
	r.GET("/api/pokemon/:id/evolution-options", h.EvolutionOptions)
	r.POST("/api/sessions", h.CreateOrLoadSession)
	r.GET("/api/sessions/:id", h.GetSession)
	r.GET("/api/sessions/:id/history", h.GetSessionHistory)
	r.POST("/api/sessions/:id/actions", h.ApplySessionAction)

	r.GET("/api/analytics/pokemon-usage", h.AnalyticsTopPokemon)
	r.GET("/api/analytics/damage", h.AnalyticsDamage)
	r.GET("/api/analytics/knockouts", h.AnalyticsKnockouts)
	r.POST("/api/analytics/visit", h.AnalyticsTrackVisit)
	r.GET("/api/analytics/visits/summary", h.AnalyticsVisitsSummary)
	r.GET("/api/analytics/visits/:visitorId", h.AnalyticsVisitorStats)

	return r
}
