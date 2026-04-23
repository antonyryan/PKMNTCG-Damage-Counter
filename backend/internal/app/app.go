package app

import (
	"net/http"
	"os"
	"path/filepath"

	"pkmntcg/backend/internal/analytics"
	"pkmntcg/backend/internal/catalog"
	"pkmntcg/backend/internal/httpapi"
	"pkmntcg/backend/internal/session"
)

type Application struct {
	Server    *http.Server
	Analytics *analytics.Store
}

func New(port string, dataDir string) *Application {
	catalogSvc := catalog.MustLoad()
	sessionStore := session.NewStore(filepath.Join(dataDir, "sessions"))
	analyticsStore := analytics.MustNew(filepath.Join(dataDir, "analytics.db"))

	handlers := httpapi.NewHandlers(catalogSvc, sessionStore, analyticsStore)
	router := httpapi.SetupRouter(handlers)

	server := &http.Server{Addr: ":" + port, Handler: router}
	return &Application{Server: server, Analytics: analyticsStore}
}

func NewFromEnv() *Application {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	return New(port, dataDir)
}
