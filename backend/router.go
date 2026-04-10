package main

import "github.com/gin-gonic/gin"

// setupRouter registers middleware and all HTTP endpoints exposed by the backend.
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", healthHandler)
	r.GET("/api/pokemon/search", searchPokemonHandler)

	return r
}
