package main

import "github.com/gin-gonic/gin"

// setupRouter registers middleware and all HTTP endpoints exposed by the backend.
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", healthHandler)
	r.GET("/api/pokemon/search", searchPokemonHandler)
	r.GET("/api/pokemon/:id/evolution-options", evolutionOptionsHandler)
	r.POST("/api/sessions", createOrLoadSessionHandler)
	r.GET("/api/sessions/:id", getSessionHandler)
	r.GET("/api/sessions/:id/history", getSessionHistoryHandler)
	r.POST("/api/sessions/:id/actions", applySessionActionHandler)

	r.GET("/api/analytics/pokemon-usage", analyticsTopPokemonHandler)
	r.GET("/api/analytics/damage", analyticsDamageHandler)
	r.GET("/api/analytics/knockouts", analyticsKnockoutsHandler)

	return r
}
