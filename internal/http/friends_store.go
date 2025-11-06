package httphandler

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"power4/internal/game"
)

type challengeInvite struct {
	Ticket        string
	Challenger    string
	ChallengerPID string
	Target        string
	Ch            chan string
	Created       time.Time
}

var (
	fmu                sync.Mutex
	friendEdges        = map[string]map[string]struct{}{}
	friendReqsIncoming = map[string]map[string]struct{}{}
	friendReqsOutgoing = map[string]map[string]struct{}{}
	invitesByTicket    = map[string]*challengeInvite{}
	invitesByTarget    = map[string]map[string]*challengeInvite{}

	friendsFilePath string
)

type friendsSnapshot struct {
	Friends          map[string][]string `json:"friends"`
	RequestsIncoming map[string][]string `json:"requests_incoming"`
	RequestsOutgoing map[string][]string `json:"requests_outgoing"`
}

func InitFriendsStore(path string) error {
	fmu.Lock()
	defer fmu.Unlock()
	friendsFilePath = path
	return loadFriendsLocked()
}

func loadFriendsLocked() error {
	if friendsFilePath == "" {
		return nil
	}
	f, err := os.Open(friendsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var snap friendsSnapshot
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	friendEdges = map[string]map[string]struct{}{}
	friendReqsIncoming = map[string]map[string]struct{}{}
	friendReqsOutgoing = map[string]map[string]struct{}{}

	for u, lst := range snap.Friends {
		us := ensureSet(friendEdges, norm(u))
		for _, v := range lst {
			us[norm(v)] = struct{}{}
		}
	}
	for t, lst := range snap.RequestsIncoming {
		ts := ensureSet(friendReqsIncoming, norm(t))
		for _, f := range lst {
			ts[norm(f)] = struct{}{}
		}
	}
	for f, lst := range snap.RequestsOutgoing {
		fs := ensureSet(friendReqsOutgoing, norm(f))
		for _, t := range lst {
			fs[norm(t)] = struct{}{}
		}
	}
	return nil
}

func snapshotLocked() friendsSnapshot {
	snap := friendsSnapshot{
		Friends:          map[string][]string{},
		RequestsIncoming: map[string][]string{},
		RequestsOutgoing: map[string][]string{},
	}
	for u, set := range friendEdges {
		var lst []string
		for v := range set {
			lst = append(lst, v)
		}
		sort.Strings(lst)
		snap.Friends[u] = lst
	}
	for t, set := range friendReqsIncoming {
		var lst []string
		for f := range set {
			lst = append(lst, f)
		}
		sort.Strings(lst)
		snap.RequestsIncoming[t] = lst
	}
	for f, set := range friendReqsOutgoing {
		var lst []string
		for t := range set {
			lst = append(lst, t)
		}
		sort.Strings(lst)
		snap.RequestsOutgoing[f] = lst
	}
	return snap
}

func saveFriendsLocked() error {
	if friendsFilePath == "" {
		return nil
	}
	snap := snapshotLocked()
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(friendsFilePath)
	tmp, err := os.CreateTemp(dir, "friends-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, werr := tmp.Write(b)
	syncErr := tmp.Sync()
	cerr := tmp.Close()
	if werr != nil {
		os.Remove(tmpName)
		return werr
	}
	if syncErr != nil {
		os.Remove(tmpName)
		return syncErr
	}
	if cerr != nil {
		os.Remove(tmpName)
		return cerr
	}
	_ = os.Remove(friendsFilePath)
	if err := os.Rename(tmpName, friendsFilePath); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
func norm(name string) string { return strings.TrimSpace(strings.ToLower(name)) }

func ensureSet(m map[string]map[string]struct{}, k string) map[string]struct{} {
	s, ok := m[k]
	if !ok {
		s = map[string]struct{}{}
		m[k] = s
	}
	return s
}

func isFriends(a, b string) bool {
	a, b = norm(a), norm(b)
	as := friendEdges[a]
	if as == nil {
		return false
	}
	_, ok := as[b]
	return ok
}

func addFriendship(a, b string) {
	a0, b0 := norm(a), norm(b)
	as := ensureSet(friendEdges, a0)
	bs := ensureSet(friendEdges, b0)
	as[b0] = struct{}{}
	bs[a0] = struct{}{}
	if s := friendReqsIncoming[a0]; s != nil {
		delete(s, b0)
	}
	if s := friendReqsIncoming[b0]; s != nil {
		delete(s, a0)
	}
	if s := friendReqsOutgoing[a0]; s != nil {
		delete(s, b0)
	}
	if s := friendReqsOutgoing[b0]; s != nil {
		delete(s, a0)
	}
	_ = saveFriendsLocked()
}

func friendsIncomingCount(user string) int {
	fmu.Lock()
	defer fmu.Unlock()
	u := norm(user)
	return len(friendReqsIncoming[u])
}

func invitesIncomingCount(user string) int {
	fmu.Lock()
	defer fmu.Unlock()
	u := norm(user)
	return len(invitesByTarget[u])
}

func listFriends(user string) []string {
	fmu.Lock()
	defer fmu.Unlock()
	u := norm(user)
	set := friendEdges[u]
	var out []string
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func listIncomingRequests(user string) []string {
	fmu.Lock()
	defer fmu.Unlock()
	u := norm(user)
	set := friendReqsIncoming[u]
	var out []string
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func sendFriendRequest(from, to string) error {
	fmu.Lock()
	defer fmu.Unlock()
	f, t := norm(from), norm(to)
	if f == "" || t == "" || f == t {
		return errors.New("invalid request")
	}
	if isFriends(f, t) {
		return errors.New("already friends")
	}
	if _, ok := ensureSet(friendReqsIncoming, t)[f]; ok {
		return nil
	}
	ensureSet(friendReqsIncoming, t)[f] = struct{}{}
	ensureSet(friendReqsOutgoing, f)[t] = struct{}{}
	return saveFriendsLocked()
}

func acceptFriendRequest(user, from string) error {
	fmu.Lock()
	defer fmu.Unlock()
	u, f := norm(user), norm(from)
	if u == "" || f == "" || u == f {
		return errors.New("invalid")
	}
	if _, ok := friendReqsIncoming[u][f]; !ok {
		return errors.New("no such request")
	}
	addFriendship(u, f)
	return saveFriendsLocked()
}

func declineFriendRequest(user, from string) error {
	fmu.Lock()
	defer fmu.Unlock()
	u, f := norm(user), norm(from)
	if u == "" || f == "" {
		return errors.New("invalid")
	}
	if s := friendReqsIncoming[u]; s != nil {
		delete(s, f)
	}
	if s := friendReqsOutgoing[f]; s != nil {
		delete(s, u)
	}
	return saveFriendsLocked()
}

func sendChallenge(challenger, challengerPID, target, ticket string) (*challengeInvite, error) {
	fmu.Lock()
	defer fmu.Unlock()
	c, t := norm(challenger), norm(target)
	if c == "" || t == "" || c == t {
		return nil, errors.New("invalid")
	}
	if !isFriends(c, t) {
		return nil, errors.New("not friends")
	}
	if _, exists := invitesByTicket[ticket]; exists {
		return nil, errors.New("duplicate")
	}
	inv := &challengeInvite{
		Ticket:        ticket,
		Challenger:    c,
		ChallengerPID: challengerPID,
		Target:        t,
		Ch:            make(chan string, 1),
		Created:       time.Now(),
	}
	invitesByTicket[ticket] = inv
	if invitesByTarget[t] == nil {
		invitesByTarget[t] = map[string]*challengeInvite{}
	}
	invitesByTarget[t][ticket] = inv
	return inv, nil
}

func cancelChallenge(ticket, by string) {
	fmu.Lock()
	defer fmu.Unlock()
	inv := invitesByTicket[ticket]
	if inv == nil {
		return
	}
	delete(invitesByTicket, ticket)
	if invs := invitesByTarget[inv.Target]; invs != nil {
		delete(invs, ticket)
		if len(invs) == 0 {
			delete(invitesByTarget, inv.Target)
		}
	}
}

func listIncomingInvites(user string) []*challengeInvite {
	fmu.Lock()
	defer fmu.Unlock()
	u := norm(user)
	var out []*challengeInvite
	for _, inv := range invitesByTarget[u] {
		out = append(out, inv)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.Before(out[j].Created) })
	return out
}

func acceptChallenge(ticket, byUser, byPID string) (string, error) {
	fmu.Lock()
	inv := invitesByTicket[ticket]
	if inv == nil {
		fmu.Unlock()
		return "", errors.New("no such invite")
	}
	if norm(byUser) != inv.Target {
		fmu.Unlock()
		return "", errors.New("not your invite")
	}
	delete(invitesByTicket, ticket)
	if invs := invitesByTarget[inv.Target]; invs != nil {
		delete(invs, ticket)
		if len(invs) == 0 {
			delete(invitesByTarget, inv.Target)
		}
	}
	fmu.Unlock()
	code := genCode()
	now := time.Now()
	rm := &Room{
		Code:         code,
		Game:         game.NewGame(),
		Player1ID:    inv.ChallengerPID,
		Player2ID:    byPID,
		Player1User:  inv.Challenger,
		Player2User:  byUser,
		CreatedAt:    now,
		Rev:          1,
		subs:         make(map[chan struct{}]struct{}),
		Random:       false,
		TurnDeadline: now.Add(2 * time.Minute),
		StartNext:    game.Player2,
	}
	rm.Game.Player1Name = inv.Challenger
	rm.Game.Player2Name = byUser

	roomsMu.Lock()
	rooms[code] = rm
	roomsMu.Unlock()
	select {
	case inv.Ch <- code:
	default:
	}
	return code, nil
}

func declineChallenge(ticket, by string) error {
	fmu.Lock()
	defer fmu.Unlock()
	inv := invitesByTicket[ticket]
	if inv == nil {
		return errors.New("no such invite")
	}
	if norm(by) != inv.Target {
		return errors.New("not your invite")
	}
	delete(invitesByTicket, ticket)
	if invs := invitesByTarget[inv.Target]; invs != nil {
		delete(invs, ticket)
		if len(invs) == 0 {
			delete(invitesByTarget, inv.Target)
		}
	}
	return nil
}

func getInvite(ticket string) *challengeInvite {
	fmu.Lock()
	defer fmu.Unlock()
	return invitesByTicket[ticket]
}

func friendsState(user, other string) (areFriends bool, outPending bool, inPending bool) {
	fmu.Lock()
	defer fmu.Unlock()
	u, o := norm(user), norm(other)
	_, areFriends = friendEdges[u][o]
	_, outPending = friendReqsOutgoing[u][o]
	_, inPending = friendReqsIncoming[u][o]
	return
}
