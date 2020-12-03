package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/gabrielmorenobrc/go-lib/util"
	"strings"
	"sync"
	"time"
)

type DataProvider interface {
	LoadSnapshot() map[string]*SessionEntry
	CreateSession(entry *SessionEntry) int64
	UpdateSessionTime(id int64, expirationTime time.Time, lastTime time.Time)
	RemoveSession(entry *SessionEntry)
	Shrink()
}

type SessionsConfig struct {
	Secret       *string `json:"secret"`
	TokenTimeout *int    `json:"tokenTimeout"`
}

type JwtTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JwtTokenPayload struct {
	UserId         int64     `json:"userId"`
	MinutesTimeout int       `json:"minutesTimeout"`
	CreationTime   time.Time `json:"creationTime"`
}

type SessionEntry struct {
	UserId         int64
	CreationTime   time.Time
	ExpirationTime time.Time
	LastTime       time.Time
	TokenString    string
	Id             *int64
}

type SessionManager struct {
	dataProvider DataProvider
	jwtConfig    SessionsConfig
	sessionMap   map[string]*SessionEntry
	mux          sync.Mutex
}

func (o *SessionsConfig) Validate() {
	if o.Secret == nil {
		panic("Invalid secret")
	}
	if o.TokenTimeout == nil {
		panic("Invalid tokenTimeout")
	}
}

func (o *SessionManager) EvictToken(tokenString string) {
	entry := o.doEvictToken(tokenString)
	o.dataProvider.RemoveSession(entry)
}

func (o *SessionManager) doEvictToken(value string) *SessionEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	entry := o.sessionMap[value]
	delete(o.sessionMap, value)
	return entry
}

func (o *SessionManager) CreateToken(userId int64) string {

	header := JwtTokenHeader{Alg: "HS256", Typ: "JWT"}

	payload := JwtTokenPayload{UserId: userId, MinutesTimeout: *o.jwtConfig.TokenTimeout, CreationTime: time.Now()}

	b, err := json.Marshal(header)
	util.CheckErr(err)
	content1 := base64.RawURLEncoding.EncodeToString(b)

	b, err = json.Marshal(payload)
	util.CheckErr(err)
	content2 := base64.RawURLEncoding.EncodeToString(b)

	content := content1 + "." + content2

	keyString := *o.jwtConfig.Secret
	key := []byte(keyString)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(content))
	value := string(content)

	tokenEntry := o.registerToken(&payload, value)
	id := o.dataProvider.CreateSession(tokenEntry)
	tokenEntry.Id = &id
	return value
}

func (o *SessionManager) ValidateToken(token string) *SessionEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		panic("Invalid token")
	}
	payload := JwtTokenPayload{}
	decodeTokenPart(&payload, parts[1])
	entry, ok := o.sessionMap[token]
	if !ok {
		return nil
	}
	if entry.ExpirationTime.Before(time.Now()) {
		delete(o.sessionMap, token)
		return nil
	}
	if entry.UserId != payload.UserId {
		return nil
	}
	entry.LastTime = time.Now()
	entry.ExpirationTime = entry.LastTime.Add(time.Minute * time.Duration(*o.jwtConfig.TokenTimeout))
	tokenCopy := *entry
	return &tokenCopy
}

func (o *SessionManager) registerToken(payload *JwtTokenPayload, token string) *SessionEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	expiration := time.Now().Add(time.Minute * time.Duration(*o.jwtConfig.TokenTimeout))
	te := SessionEntry{UserId: payload.UserId, CreationTime: time.Now(), ExpirationTime: expiration, LastTime: time.Now(), TokenString: token}
	o.sessionMap[token] = &te
	return &te
}

func (o *SessionManager) Shrink() {
	o.dataProvider.Shrink()
}

func (o *SessionManager) Load() {
	o.sessionMap = o.dataProvider.LoadSnapshot()
}

func NewSessionManager(dataProvider DataProvider, jwtConfig SessionsConfig) *SessionManager {
	tm := SessionManager{dataProvider: dataProvider, jwtConfig: jwtConfig, sessionMap: make(map[string]*SessionEntry), mux: sync.Mutex{}}
	return &tm
}

func decodeTokenPart(i interface{}, part string) {
	jsonBytes, err := base64.RawURLEncoding.DecodeString(part)
	util.CheckErr(err)
	util.JsonDecode(i, bytes.NewReader(jsonBytes))
}

type MockDataProvider struct {
	DataProvider
}

func (o *MockDataProvider) LoadSnapshot() map[string]*SessionEntry {
	return nil
}

func (o *MockDataProvider) CreateSession(entry *SessionEntry) int64 {
	return 0
}
func (o *MockDataProvider) UpdateSessionTime(id int64, expirationTime time.Time, lastTime time.Time) {

}
func (o *MockDataProvider) RemoveSession(entry *SessionEntry) {

}
func (o *MockDataProvider) Shrink() {

}
