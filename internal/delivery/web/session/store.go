package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var sessionKey = []byte("very-secret-key-32-bytes-----")
var store = sessions.NewCookieStore(sessionKey)

func Get(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "session-name")
}

func Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return s.Save(r, w)
}
