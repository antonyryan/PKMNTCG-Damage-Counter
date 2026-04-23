package httpapi

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"pkmntcg/backend/internal/analytics"
	"pkmntcg/backend/internal/catalog"
	"pkmntcg/backend/internal/session"
)

type createSessionRequest struct {
	SessionID string `json:"sessionId"`
}

type SessionStore interface {
	GetOrCreate(sessionID string) (*session.GameSession, error)
	Get(sessionID string) (*session.GameSession, error)
	ApplyAction(sessionID string, req session.ActionRequest, catalog session.CatalogLookup) (*session.GameSession, error)
}

type AnalyticsStore interface {
	Record(sessionID string, req session.ActionRequest, resolver analytics.NameResolver, now time.Time)
	QueryTopPokemon(limit int) ([]analytics.PokemonUsageEntry, error)
	QueryDamageTotals() (analytics.DamageTotalsResponse, error)
	QueryKnockouts() (analytics.KnockoutTotalResponse, error)
}

type Handlers struct {
	catalog   *catalog.Service
	sessions  SessionStore
	analytics AnalyticsStore
}

func NewHandlers(catalogSvc *catalog.Service, sessionStore SessionStore, analyticsStore AnalyticsStore) *Handlers {
	return &Handlers{catalog: catalogSvc, sessions: sessionStore, analytics: analyticsStore}
}

func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) SearchPokemon(c *gin.Context) {
	query := c.Query("q")
	limit := parseLimit(c.DefaultQuery("limit", strconv.Itoa(catalog.DefaultSearchLimit)), catalog.DefaultSearchLimit)
	c.JSON(http.StatusOK, h.catalog.Search(query, limit))
}

func (h *Handlers) EvolutionOptions(c *gin.Context) {
	pokemonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pokemon id"})
		return
	}
	results, err := h.catalog.EvolutionOptions(pokemonID, c.Query("q"))
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

func (h *Handlers) CreateOrLoadSession(c *gin.Context) {
	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	sessionData, err := h.sessions.GetOrCreate(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionData)
}

func (h *Handlers) GetSession(c *gin.Context) {
	sessionData, err := h.sessions.Get(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionData)
}

func (h *Handlers) GetSessionHistory(c *gin.Context) {
	sessionData, err := h.sessions.Get(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionData.History)
}

func (h *Handlers) ApplySessionAction(c *gin.Context) {
	var req session.ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	now := time.Now()
	sessionData, err := h.sessions.ApplyAction(c.Param("id"), req, catalogLookupAdapter{svc: h.catalog})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	h.analytics.Record(c.Param("id"), req, h.catalog, now)
	log.Printf("session=%s action=%s summary=%s", c.Param("id"), req.Type, actionSummary(req, sessionData, h.catalog))
	c.JSON(http.StatusOK, sessionData)
}

func (h *Handlers) AnalyticsTopPokemon(c *gin.Context) {
	limit := parseLimit(c.DefaultQuery("limit", "10"), catalog.DefaultSearchLimit)
	entries, err := h.analytics.QueryTopPokemon(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pokemon": entries})
}

func (h *Handlers) AnalyticsDamage(c *gin.Context) {
	totals, err := h.analytics.QueryDamageTotals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, totals)
}

func (h *Handlers) AnalyticsKnockouts(c *gin.Context) {
	totals, err := h.analytics.QueryKnockouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, totals)
}

type catalogLookupAdapter struct {
	svc *catalog.Service
}

func (a catalogLookupAdapter) FindByID(id int) (session.PokemonInfo, bool) {
	entry, ok := a.svc.Get(id)
	if !ok {
		return session.PokemonInfo{}, false
	}
	return session.PokemonInfo{ID: entry.Pokemon.ID, Name: entry.Pokemon.Name}, true
}

func (a catalogLookupAdapter) IsValidEvolutionTarget(currentPokemonID, targetPokemonID int) bool {
	return a.svc.IsValidEvolutionTarget(currentPokemonID, targetPokemonID)
}

func actionSummary(req session.ActionRequest, gameSession *session.GameSession, catalogSvc *catalog.Service) string {
	zoneLabel := string(req.Zone)
	if req.Zone == session.ZoneBench && req.BenchIndex != nil {
		zoneLabel = fmt.Sprintf("bench#%d", *req.BenchIndex+1)
	}

	sideLabel := string(req.Side)
	if sideLabel == "" {
		sideLabel = "n/a"
	}

	switch req.Type {
	case session.ActionAdjust:
		amount := 0
		if req.Amount != nil {
			amount = *req.Amount
		}
		sign := ""
		if amount > 0 {
			sign = "+"
		}
		pokemonName := pokemonNameFromSlot(gameSession, req)
		if pokemonName == "" {
			pokemonName = "Unknown Pokemon"
		}
		return fmt.Sprintf("%s, %s%d damage (%s, %s)", pokemonName, sign, amount, zoneLabel, sideLabel)
	case session.ActionSetPokemon, session.ActionEvolve:
		if req.PokemonID != nil {
			if pokemon, ok := catalogSvc.Get(*req.PokemonID); ok {
				return fmt.Sprintf("%s on %s (%s)", pokemon.Pokemon.Name, zoneLabel, sideLabel)
			}
			return fmt.Sprintf("pokemonId=%d on %s (%s)", *req.PokemonID, zoneLabel, sideLabel)
		}
		return fmt.Sprintf("pokemon update on %s (%s)", zoneLabel, sideLabel)
	case session.ActionKnockout:
		return fmt.Sprintf("knockout on %s (%s)", zoneLabel, sideLabel)
	case session.ActionStatus:
		if req.Status != nil {
			return fmt.Sprintf("toggle %s status (%s)", *req.Status, sideLabel)
		}
		return fmt.Sprintf("toggle status (%s)", sideLabel)
	case session.ActionPromote:
		return fmt.Sprintf("promote from %s (%s)", zoneLabel, sideLabel)
	case session.ActionToggleGX:
		return fmt.Sprintf("toggle GX (%s)", sideLabel)
	case session.ActionToggleVSTAR:
		return fmt.Sprintf("toggle VSTAR (%s)", sideLabel)
	case session.ActionFlipCoin:
		if gameSession.State.CoinResult != nil {
			return fmt.Sprintf("flip coin -> %s", *gameSession.State.CoinResult)
		}
		return "flip coin"
	case session.ActionRollDie:
		if gameSession.State.DieResult != nil {
			return fmt.Sprintf("roll die -> %d", *gameSession.State.DieResult)
		}
		return "roll die"
	case session.ActionReset:
		return "reset game"
	default:
		return req.Type
	}
}

func pokemonNameFromSlot(gameSession *session.GameSession, req session.ActionRequest) string {
	var player *session.PlayerState
	switch req.Side {
	case session.SideMe:
		player = &gameSession.State.Me
	case session.SideOpponent:
		player = &gameSession.State.Opponent
	default:
		return ""
	}

	var slot *session.SlotState
	switch req.Zone {
	case session.ZoneActive:
		slot = &player.Active
	case session.ZoneBench:
		if req.BenchIndex == nil || *req.BenchIndex < 0 || *req.BenchIndex >= len(player.Bench) {
			return ""
		}
		slot = &player.Bench[*req.BenchIndex]
	default:
		return ""
	}

	if slot.Pokemon == nil {
		return ""
	}
	return slot.Pokemon.Name
}

func parseLimit(raw string, defaultValue int) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultValue
	}
	if limit > 100 {
		return 100
	}
	return limit
}
