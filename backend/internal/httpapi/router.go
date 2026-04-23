package httpapi

import "github.com/gin-gonic/gin"

func SetupRouter(h *Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

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
