package gosession

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

func generateSessionID() string {
	bytes := make([]byte, 40)

	rand.Read(bytes[:32])

	binary.LittleEndian.PutUint64(bytes[32:], uint64(time.Now().UnixNano()))

	hash := sha256.Sum256(bytes)

	idHex := make([]byte, len(hash)*2)
	hex.Encode(idHex, hash[:32])
	return string(idHex)
}

// SessionProvider is the one that stores the session somewhere
type SessionProvider interface {
	Save(id string, data map[string]interface{}) error

	Delete(id string) error

	Get(id string, data *map[string]interface{}) (bool, error)
}

// Session contain the session data
type Session struct {
	// public
	Data map[string]interface{}

	ID string

	// private
	store *SessionStore
}

// Save the changes made to the session
func (s *Session) Save() error {
	return s.store.provider.Save(s.ID, s.Data)
}

// RenewID creates a new session id and saves the session
func (s *Session) RenewID(w http.ResponseWriter) {
	newID := generateSessionID()

	s.store.provider.Save(newID, s.Data)
	s.store.provider.Delete(s.ID)

	s.ID = newID

	if s.store.isCookieBased {
		http.SetCookie(w, &http.Cookie{
			Name:     s.store.name,
			Value:    s.ID,
			MaxAge:   0,
			HttpOnly: true})
	}
}

// Destroy the session
func (s *Session) Destroy(w http.ResponseWriter) error {
	s.store.provider.Delete(s.ID)

	if s.store.isCookieBased {
		// When MaxAge is negative the cookie gets deleted
		http.SetCookie(w, &http.Cookie{
			Name:     s.store.name,
			Value:    s.ID,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true})
	}

	return nil
}

// The SessionStore manages the retrieval of a session
type SessionStore struct {
	provider SessionProvider

	isCookieBased bool

	autoCreateWhenMissing bool

	// If the store is cookie based it will be the cookie name
	// if the store is header based it will be the header name
	name string
}

// New ... creates a session
func (s *SessionStore) New(w http.ResponseWriter) (*Session, error) {
	sessionID := generateSessionID()

	if s.isCookieBased {
		http.SetCookie(w, &http.Cookie{
			Name:     s.name,
			Value:    sessionID,
			MaxAge:   0,
			HttpOnly: true})
	}

	session := Session{
		ID:    sessionID,
		store: s,
		Data:  make(map[string]interface{})}
	session.Save()
	return &session, nil
}

// Get the session for this request
func (s *SessionStore) Get(r *http.Request,
	w http.ResponseWriter) (*Session, error) {
	var sessionID string

	if s.isCookieBased {
		cookie, err := r.Cookie(s.name)

		if err == http.ErrNoCookie && s.autoCreateWhenMissing {
			// Create new Session
			sessionID = generateSessionID()

			http.SetCookie(w, &http.Cookie{
				Name:     s.name,
				Value:    sessionID,
				MaxAge:   0,
				HttpOnly: true})

			session := Session{
				ID:    sessionID,
				store: s,
				Data:  make(map[string]interface{})}
			session.Save()
			return &session, nil
		} else if err == http.ErrNoCookie {
			return nil, nil
		}

		sessionID = cookie.Value
	} else {
		if headers, ok := r.Header[s.name]; ok {
			// In cases there are multiple headers with s name
			// It will only try the first one
			sessionID = headers[0]
		} else {
			if s.autoCreateWhenMissing {
				session := Session{
					ID:    generateSessionID(),
					store: s,
					Data:  make(map[string]interface{})}

				session.Save()
				return &session, nil
			} else {
				return nil, nil
			}
		}
	}

	return s.GetByID(sessionID)
}

// GetByID returnes the session that corresponds the the ID
func (s *SessionStore) GetByID(id string) (*Session, error) {
	session := Session{
		ID:    id,
		store: s,
		Data:  make(map[string]interface{})}

	ok, err := s.provider.Get(id, &session.Data)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !ok && s.autoCreateWhenMissing {
		// Session doesn't exist in memcache.
		// Generate a new ID (New session)
		session.ID = generateSessionID()
		session.Save()
	} else if !ok {
		return nil, nil
	}

	return &session, nil
}

// NewCookieBasedSessionStore creates a Cookie based sesion store
func NewCookieBasedSessionStore(
	provider SessionProvider,
	cookie string,
	createWhenMissing bool) *SessionStore {
	return &SessionStore{
		provider:              provider,
		isCookieBased:         true,
		name:                  cookie,
		autoCreateWhenMissing: createWhenMissing}
}

// NewHeaderBasedSessionStore create a new SessionStore based with id passed
// as header values
func NewHeaderBasedSessionStore(
	provider SessionProvider,
	header string,
	createWhenMissing bool) *SessionStore {

	return &SessionStore{
		provider:              provider,
		isCookieBased:         false,
		name:                  header,
		autoCreateWhenMissing: createWhenMissing}
}
