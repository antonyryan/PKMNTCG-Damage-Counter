package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var pokemonCatalog = mustLoadPokemonCatalog()
var sessionStore = newSessionStore("data/sessions")
var analyticsStore = mustNewAnalyticsStore("data/analytics.db")

type createSessionRequest struct {
	SessionID string `json:"sessionId"`
}

// healthHandler provides a simple liveness endpoint for development and tooling.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// searchPokemonHandler serves autocomplete results from the backend-owned catalog.
func searchPokemonHandler(c *gin.Context) {
	query := c.Query("q")
	limit := parseLimit(c.DefaultQuery("limit", strconv.Itoa(defaultSearchLimit)))
	c.JSON(http.StatusOK, pokemonCatalog.Search(query, limit))
}

// evolutionOptionsHandler returns only valid evolution or de-evolution choices for one pokemon.
func evolutionOptionsHandler(c *gin.Context) {
	pokemonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pokemon id"})
		return
	}

	results, err := pokemonCatalog.EvolutionOptions(pokemonID, c.Query("q"))
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// createOrLoadSessionHandler restores a persisted session or creates a new one.
func createOrLoadSessionHandler(c *gin.Context) {
	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	session, err := sessionStore.GetOrCreate(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// getSessionHandler returns the current authoritative snapshot for one session.
func getSessionHandler(c *gin.Context) {
	session, err := sessionStore.Get(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// getSessionHistoryHandler returns the persisted action log for one session.
func getSessionHistoryHandler(c *gin.Context) {
	session, err := sessionStore.Get(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session.History)
}

// applySessionActionHandler applies one validated state mutation and persists both state and history.
func applySessionActionHandler(c *gin.Context) {
	var req SessionActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	now := time.Now()
	session, err := sessionStore.ApplyAction(c.Param("id"), req, pokemonCatalog)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	analyticsStore.Record(c.Param("id"), req, pokemonCatalog, now)
	c.JSON(http.StatusOK, session)
}

// analyticsTopPokemonHandler returns the most used Pokemon sorted by usage count.
func analyticsTopPokemonHandler(c *gin.Context) {
	limit := parseLimit(c.DefaultQuery("limit", "10"))
	entries, err := analyticsStore.QueryTopPokemon(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pokemon": entries})
}

// analyticsDamageHandler returns total damage dealt and total intentional damage healed.
func analyticsDamageHandler(c *gin.Context) {
	totals, err := analyticsStore.QueryDamageTotals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, totals)
}

// analyticsKnockoutsHandler returns the total number of knockouts recorded.
func analyticsKnockoutsHandler(c *gin.Context) {
	totals, err := analyticsStore.QueryKnockouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, totals)
}

func parseLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultSearchLimit
	}
	if limit > 100 {
		return 100
	}
	return limit
}
