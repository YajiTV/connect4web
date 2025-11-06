package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Username     string
	PasswordHash []byte
	CreatedAt    time.Time
	Elo          int
	Games        int
	Wins         int
	Losses       int
}

type Store struct {
	mu     sync.RWMutex
	byID   map[string]*User
	byName map[string]*User
	path   string
}

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

func (s *Store) save() error {
	users := snapshotUsers(s)
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func randID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), "=")
}

func (s *Store) Create(username, password string) (*User, error) {
	un := strings.TrimSpace(username)
	if un == "" {
		return nil, errors.New("empty username")
	}
	if len(password) < 6 {
		return nil, errors.New("weak password")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	lc := strings.ToLower(un)

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

func (s *Store) GetByUsername(username string) *User {
	lc := strings.ToLower(strings.TrimSpace(username))
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byName[lc]
}

func (s *Store) GetByID(id string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[id]
}

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
	ea := expected(ua.Elo, ub.Elo)
	da := int(round(float64(k) * (scoreA - ea)))
	db := -da
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

func (s *Store) UsersByElo(query string) []*User {
	q := strings.ToLower(strings.TrimSpace(query))
	s.mu.RLock()
	users := make([]*User, 0, len(s.byID))
	for _, u := range s.byID {
		if q == "" || strings.Contains(strings.ToLower(u.Username), q) {
			cp := *u
			users = append(users, &cp)
		}
	}
	s.mu.RUnlock()
	sort.Slice(users, func(i, j int) bool {
		if users[i].Elo == users[j].Elo {
			return users[i].Username < users[j].Username
		}
		return users[i].Elo > users[j].Elo
	})
	return users
}
