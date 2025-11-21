package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"power4/internal/util"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    // unique user id
	Username     string    // chosen username
	PasswordHash []byte    // bcrypt hash of the password
	CreatedAt    time.Time // account creation time
	Elo          int       // current Elo rating
	Games        int       // total games played
	Wins         int       // total wins
	Losses       int       // total losses
}

type Store struct {
	mu     sync.RWMutex     // guards maps
	byID   map[string]*User // users by id
	byName map[string]*User // users by lowercase username
	path   string           // path to users.json
}

// Open opens or creates the user store under dir and tries to load existing users from disk
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(dir, "users.json")

	s := &Store{
		byID:   make(map[string]*User),
		byName: make(map[string]*User),
		path:   p,
	}

	// tries to load existing data
	if b, err := os.ReadFile(p); err == nil {
		var users []*User
		if err := json.Unmarshal(b, &users); err != nil {
			return nil, err
		}
		for _, u := range users {
			s.byID[u.ID] = u
			s.byName[strings.ToLower(u.Username)] = u
		}
	}
	return s, nil
}

// snapshotUsers generates a copy of all users to avoid holding locks during JSON encoding
func snapshotUsers(s *Store) []*User {
	s.mu.RLock()
	out := make([]*User, 0, len(s.byID))
	for _, u := range s.byID {
		cp := *u
		out = append(out, &cp)
	}
	s.mu.RUnlock()
	return out
}

// save writes the current users snapshot to disk as JSON
func (s *Store) save() error {
	users := snapshotUsers(s)
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// randID generates a random base32 id of length n and panics on failure
func randID(n int) string {
	id, err := util.RandBase32(n)
	if err != nil {
		panic(err)
	}
	return id
}

// Create creates a new user, generates a bcrypt hash, and persists the store
func (s *Store) Create(username, password string) (*User, error) {
	un := strings.TrimSpace(username)
	if un == "" {
		return nil, errors.New("empty username")
	}
	if len(password) < 6 {
		return nil, errors.New("weak password")
	}

	// generates password hash
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	lc := strings.ToLower(un)

	// tries to reserve the username
	s.mu.Lock()
	if _, ok := s.byName[lc]; ok {
		s.mu.Unlock()
		return nil, errors.New("username taken")
	}

	u := &User{
		ID:           randID(12),
		Username:     un,
		PasswordHash: h,
		CreatedAt:    time.Now(),
		Elo:          1500,
	}
	s.byID[u.ID] = u
	s.byName[lc] = u
	s.mu.Unlock()

	if err := s.save(); err != nil {
		return nil, err
	}
	return u, nil
}

// Authenticate tries to find the user and verifies the password with bcrypt
func (s *Store) Authenticate(username, password string) (*User, error) {
	lc := strings.ToLower(strings.TrimSpace(username))

	s.mu.RLock()
	u := s.byName[lc]
	s.mu.RUnlock()
	if u == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// GetByUsername returns the user by username or nil if absent
func (s *Store) GetByUsername(username string) *User {
	lc := strings.ToLower(strings.TrimSpace(username))
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byName[lc]
}

// GetByID returns the user by id or nil if absent
func (s *Store) GetByID(id string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[id]
}

// ApplyMatch applies a match result, updates Elo and stats, and saves the store
func (s *Store) ApplyMatch(usernameA, usernameB string, scoreA float64, k int) error {
	lca := strings.ToLower(usernameA)
	lcb := strings.ToLower(usernameB)

	s.mu.Lock()
	ua := s.byName[lca]
	ub := s.byName[lcb]
	if ua == nil || ub == nil {
		s.mu.Unlock()
		return errors.New("user not found")
	}

	// calculates expected score and generated delta using Elo formula
	ea := expected(ua.Elo, ub.Elo)
	da := int(round(float64(k) * (scoreA - ea)))
	db := -da

	// updates ratings and stats
	ua.Elo += da
	ub.Elo += db
	ua.Games++
	ub.Games++
	if scoreA > 0.5 {
		ua.Wins++
		ub.Losses++
	} else if scoreA < 0.5 {
		ub.Wins++
		ua.Losses++
	}
	s.mu.Unlock()

	return s.save()
}

// UsersByElo returns users filtered by query and sorted by Elo desc then username asc
func (s *Store) UsersByElo(query string) []*User {
	q := strings.ToLower(strings.TrimSpace(query))

	// copies matched users to avoid exposing internal pointers
	s.mu.RLock()
	users := make([]*User, 0, len(s.byID))
	for _, u := range s.byID {
		if q == "" || strings.Contains(strings.ToLower(u.Username), q) {
			cp := *u
			users = append(users, &cp)
		}
	}
	s.mu.RUnlock()

	// sorts by Elo then username
	sort.Slice(users, func(i, j int) bool {
		if users[i].Elo == users[j].Elo {
			return users[i].Username < users[j].Username
		}
		return users[i].Elo > users[j].Elo
	})
	return users
}
