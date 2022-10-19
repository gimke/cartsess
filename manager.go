package cartsess

import (
	"log"
	"net/http"
	"time"

	"github.com/gimke/cart"
)

var firstKey string = ""

const Version = "v1.0.6"

const (
	prefixKey   = "github.com/gimke/cartsess:"
	errorFormat = "[CART-SESS] %v ERROR! %s\n"
	infoFormat  = "[CART-SESS] %v INFO %s\n"
)

// NewSession is called by session stores to create a new session instance.
func NewManager(cookieName string, store Store) cart.Handler {
	return func(c *cart.Context, next cart.Next) {
		s := &SessionManager{
			cookieName: cookieName,
			store:      store,
			request:    c.Request,
			written:    false,
			response:   c.Response,
		}
		c.Set(prefixKey+cookieName, s)
		if firstKey == "" {
			//save firstKey as De
			firstKey = prefixKey + cookieName
		}
		next()
	}
}

// shortcut to get session
func Default(c *cart.Context) *SessionManager {
	return c.MustGet(firstKey).(*SessionManager)
}

func GetByName(c *cart.Context, cookieName string) *SessionManager {
	return c.MustGet(prefixKey + cookieName).(*SessionManager)
}

type SessionManager struct {
	cookieName string
	store      Store
	id         string
	request    *http.Request
	response   http.ResponseWriter
	session    *Session
	written    bool
}

func (s *SessionManager) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

func (s *SessionManager) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
	s.Save()
}

func (s *SessionManager) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
	s.Save()
}

func (s *SessionManager) Destroy() {
	//for key := range s.Session().Values {
	//	s.Delete(key)
	//}
	s.written = true
	//clear values
	s.Session().Values = make(map[interface{}]interface{})
	s.Session().Destroy(s.request, s.response)
}
func (s *SessionManager) Session() *Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.cookieName)
		if err != nil {
			now := time.Now().Format("2006-01-02 15:04:05")
			log.Printf(errorFormat, now, err)
		}
	}
	return s.session
}

// Save is a convenience method to save this session. It is the same as calling
// store.Save(request, response, session). You should call Save before writing to
// the response or returning from the handler.
func (s *SessionManager) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.response)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

// Name returns the name used to register the session.
func (s *SessionManager) Name() string {
	return s.cookieName
}

// Store returns the session store used to register the session.
func (s *SessionManager) Store() Store {
	return s.store
}

func (s *SessionManager) Written() bool {
	return s.written
}
