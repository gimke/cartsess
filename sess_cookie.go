package cartsess

import (
	"net/http"
)

type CookieStore struct {
	Codecs  []Codec
	Options *Options // default configuration
}

var _ Store = &CookieStore{}

func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	cs := &CookieStore{
		Codecs: CodecsFromPairs(keyPairs...),
		Options: &Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
	}

	cs.MaxAge(cs.Options.MaxAge)
	return cs
}

func (s *CookieStore) Get(r *http.Request, cookieName string) (session *Session, err error) {
	session, err = s.New(r, cookieName)
	session.cookieName = cookieName
	session.store = s
	return
}

func (s *CookieStore) New(r *http.Request, cookieName string) (*Session, error) {
	session := NewSession(s, cookieName)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(cookieName); errCookie == nil {
		err = DecodeMulti(cookieName, c.Value, &session.Values,
			s.Codecs...)
		if err == nil {
			session.IsNew = false
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *CookieStore) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	encoded, err := EncodeMulti(session.CookieName(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}

	cookie := NewCookie(session.CookieName(), encoded, session.Options)
	http.SetCookie(w, cookie)
	return nil
}

func (s *CookieStore) Destroy(r *http.Request, w http.ResponseWriter, session *Session) error {
	opt := &Options{
		Path:     session.Options.Path,
		Domain:   session.Options.Domain,
		Secure:   session.Options.Secure,
		HttpOnly: session.Options.HttpOnly,
		SameSite: session.Options.SameSite,
		MaxAge:   -1,
	}
	http.SetCookie(w, NewCookie(session.CookieName(), "", opt))
	return nil
}

func (s *CookieStore) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}
