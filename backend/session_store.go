package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// GameHistoryEntry captures user-visible state mutations for restore/debug/audit flows.
type GameHistoryEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	Action     string         `json:"action"`
	Side       Side           `json:"side,omitempty"`
	Zone       Zone           `json:"zone,omitempty"`
	BenchIndex *int           `json:"benchIndex,omitempty"`
	PokemonID  *int           `json:"pokemonId,omitempty"`
	Amount     *int           `json:"amount,omitempty"`
	Status     *SpecialStatus `json:"status,omitempty"`
}

// GameSession is the persisted unit restored by the frontend between reloads.
type GameSession struct {
	SessionID string             `json:"sessionId"`
	State     GameState          `json:"state"`
	History   []GameHistoryEntry `json:"history"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// SessionStore persists current match sessions and their history.
type SessionStore struct {
	mu       sync.RWMutex
	dirPath  string
	sessions map[string]*GameSession
}

func newSessionStore(dirPath string) *SessionStore {
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		panic(err)
	}
	return &SessionStore{
		dirPath:  dirPath,
		sessions: map[string]*GameSession{},
	}
}

func (store *SessionStore) GetOrCreate(sessionID string) (*GameSession, error) {
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
		generated, err := generateSessionID()
		if err != nil {
			return nil, err
		}
		sessionID = generated
	}

	session := &GameSession{
		SessionID: sessionID,
		State:     initialState(),
		History:   []GameHistoryEntry{},
		UpdatedAt: time.Now().UTC(),
	}
	normalizeGameState(&session.State)
	store.sessions[sessionID] = session
	if err := store.persistLocked(session); err != nil {
		return nil, err
	}
	return cloneSession(session)
}

func (store *SessionStore) Get(sessionID string) (*GameSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, err := store.loadLocked(sessionID)
	if err != nil {
		return nil, err
	}
	return cloneSession(session)
}

func (store *SessionStore) ApplyAction(sessionID string, req SessionActionRequest, catalog *PokemonCatalog) (*GameSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, err := store.loadLocked(sessionID)
	if err != nil {
		return nil, err
	}

	if err := applyAction(&session.State, req, catalog); err != nil {
		return nil, err
	}
	normalizeGameState(&session.State)

	session.History = append(session.History, GameHistoryEntry{
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

func (store *SessionStore) loadLocked(sessionID string) (*GameSession, error) {
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
	normalizeGameState(&session.State)
	if session.History == nil {
		session.History = []GameHistoryEntry{}
	}
	store.sessions[sessionID] = &session
	return &session, nil
}

func (store *SessionStore) persistLocked(session *GameSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session %s: %w", session.SessionID, err)
	}
	if err := os.WriteFile(store.filePath(session.SessionID), data, 0o644); err != nil {
		return fmt.Errorf("write session %s: %w", session.SessionID, err)
	}
	return nil
}

func (store *SessionStore) filePath(sessionID string) string {
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
	normalizeGameState(&clone.State)
	if clone.History == nil {
		clone.History = []GameHistoryEntry{}
	}
	return &clone, nil
}

func generateSessionID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func randomBool() bool {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return false
	}
	return n.Int64() == 1
}

func randomDieRoll() int {
	n, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return 1
	}
	return int(n.Int64()) + 1
}
