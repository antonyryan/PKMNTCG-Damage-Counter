package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store persists current match sessions and their history.
type Store struct {
	mu       sync.RWMutex
	dirPath  string
	sessions map[string]*GameSession
}

func NewStore(dirPath string) *Store {
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		panic(err)
	}
	return &Store{dirPath: dirPath, sessions: map[string]*GameSession{}}
}

func (store *Store) GetOrCreate(sessionID string) (*GameSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if sessionID != "" {
		session, err := store.loadLocked(sessionID)
		if err == nil {
			return cloneSession(session)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if sessionID == "" {
		generated, err := GenerateSessionID()
		if err != nil {
			return nil, err
		}
		sessionID = generated
	}

	session := &GameSession{SessionID: sessionID, State: InitialState(), History: []HistoryEntry{}, UpdatedAt: time.Now().UTC()}
	NormalizeGameState(&session.State)
	store.sessions[sessionID] = session
	if err := store.persistLocked(session); err != nil {
		return nil, err
	}
	return cloneSession(session)
}

func (store *Store) Get(sessionID string) (*GameSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, err := store.loadLocked(sessionID)
	if err != nil {
		return nil, err
	}
	return cloneSession(session)
}

func (store *Store) ApplyAction(sessionID string, req ActionRequest, catalog CatalogLookup) (*GameSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, err := store.loadLocked(sessionID)
	if err != nil {
		return nil, err
	}

	if err := ApplyAction(&session.State, req, catalog); err != nil {
		return nil, err
	}
	NormalizeGameState(&session.State)

	session.History = append(session.History, HistoryEntry{
		Timestamp:  time.Now().UTC(),
		Action:     req.Type,
		Side:       req.Side,
		Zone:       req.Zone,
		BenchIndex: req.BenchIndex,
		PokemonID:  req.PokemonID,
		Amount:     req.Amount,
		Status:     req.Status,
	})
	session.UpdatedAt = time.Now().UTC()

	if err := store.persistLocked(session); err != nil {
		return nil, err
	}
	return cloneSession(session)
}

func (store *Store) loadLocked(sessionID string) (*GameSession, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	if session, ok := store.sessions[sessionID]; ok {
		return session, nil
	}

	data, err := os.ReadFile(store.filePath(sessionID))
	if err != nil {
		return nil, err
	}

	var session GameSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("parse session %s: %w", sessionID, err)
	}
	NormalizeGameState(&session.State)
	if session.History == nil {
		session.History = []HistoryEntry{}
	}
	store.sessions[sessionID] = &session
	return &session, nil
}

func (store *Store) persistLocked(session *GameSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session %s: %w", session.SessionID, err)
	}
	if err := os.WriteFile(store.filePath(session.SessionID), data, 0o644); err != nil {
		return fmt.Errorf("write session %s: %w", session.SessionID, err)
	}
	return nil
}

func (store *Store) filePath(sessionID string) string {
	return filepath.Join(store.dirPath, sessionID+".json")
}

func cloneSession(session *GameSession) (*GameSession, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	var clone GameSession
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}
	NormalizeGameState(&clone.State)
	if clone.History == nil {
		clone.History = []HistoryEntry{}
	}
	return &clone, nil
}
