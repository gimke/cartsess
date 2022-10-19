package cartsess

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

type RedisStore struct {
	Options         *Options // default configuration
	SessionIDLength int
	Client          *redis.Client
	Prefix          string
}

var _ Store = &RedisStore{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Serialize to JSON. Will err if there are unmarshalable key values
func Serialize(s *Session) ([]byte, error) {
	m := make(map[string]interface{}, len(s.Values))
	for k, v := range s.Values {
		ks, ok := k.(string)
		if !ok {
			err := fmt.Errorf("Non-string key value, cannot serialize session to JSON: %v", k)
			fmt.Printf("redistore.JSONSerializer.serialize() Error: %v", err)
			return nil, err
		}
		m[ks] = v
	}
	return json.Marshal(m)
}

// Deserialize back to map[string]interface{}
func Deserialize(d []byte, s *Session) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(d, &m)
	if err != nil {
		fmt.Printf("redistore.JSONSerializer.deserialize() Error: %v", err)
		return err
	}
	for k, v := range m {
		s.Values[k] = v
	}
	return nil
}

func NewRedisStore() *RedisStore {
	s := &RedisStore{
		Options: &Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		SessionIDLength: 64,
		Prefix:          "",
		Client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
	return s
}

func (s *RedisStore) Get(r *http.Request, cookieName string) (session *Session, err error) {
	session, err = s.New(r, cookieName)
	session.cookieName = cookieName
	session.store = s
	return
}

func (s *RedisStore) New(r *http.Request, cookieName string) (*Session, error) {
	session := NewSession(s, cookieName)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if sid, errCookie := r.Cookie(cookieName); errCookie == nil {
		session.ID = sid.Value
		//get value
		val, err := s.Client.Get(s.Prefix + sid.Value).Result()
		if err == nil {
			err = Deserialize([]byte(val), session)
		}
	} else {
		newid, errId := s.generateID()
		err = errId
		session.ID = newid
		session.IsNew = true
	}
	return session, err
}

func (s *RedisStore) generateID() (string, error) {
	var err error = nil
	if s.SessionIDLength < 32 {
		err = errors.New("SessionIDLength is too short the value should >= 32")
		s.SessionIDLength = 32
	}
	b := make([]rune, s.SessionIDLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b), err

}

// Save adds a single session to the response.
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	sid := session.ID
	b, err := Serialize(session)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = s.Client.Set(s.Prefix+sid, string(b), time.Duration(s.Options.MaxAge)*time.Second).Err()
	if err != nil {
		return err
	}

	cookie := NewCookie(session.CookieName(), session.ID, session.Options)
	http.SetCookie(w, cookie)

	return nil
}

func (s *RedisStore) Destroy(r *http.Request, w http.ResponseWriter, session *Session) error {
	sid := session.ID
	s.Client.Del(s.Prefix + sid)
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
