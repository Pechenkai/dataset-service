package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Ключ для шифрования cookie. В продакшене храните этот ключ в env,
// но для MVP можно в коде (строку нужно сделать 32 или 64 байта).
var sessionKey = []byte("very-secret-key-32-bytes-----")
var store = sessions.NewCookieStore(sessionKey)

// Get получает сессию по запросу
func Get(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "session-name")
}

// Save сохраняет сессию (cookie)
func Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return s.Save(r, w)
}
