package cartsess

import (
	"net/http"
)

type Options struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

type Store interface {
	Get(r *http.Request, name string) (*Session, error)
	New(r *http.Request, name string) (*Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *Session) error
	Destroy(r *http.Request, w http.ResponseWriter, s *Session) error
}

func NewCookie(name, value string, options *Options) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}
	//if options.MaxAge > 0 {
	//	d := time.Duration(options.MaxAge) * time.Second
	//	cookie.Expires = time.Now().Add(d)
	//} else if options.MaxAge < 0 {
	//	// Set it to the past to expire now.
	//	cookie.Expires = time.Unix(1, 0)
	//}
	return cookie
}
