package apitest

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tsu-self/internal/pkg/xerrors"
)

// Session 表示缓存的会话信息。
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionKey 唯一标识一个会话缓存文件。
type SessionKey struct {
	Service  string
	Username string
	BaseURL  string
}

// SessionFetcher 登录回调。
type SessionFetcher func(context.Context) (*Session, error)

// SessionManager 在磁盘/内存中保存可复用的 token。
type SessionManager struct {
	dir    string
	maxAge time.Duration
	mu     sync.Mutex
	cache  map[string]*Session
}

var (
	defaultSessionsOnce sync.Once
	defaultSessions     *SessionManager
)

// DefaultSessionManager 返回全局实例。
func DefaultSessionManager() *SessionManager {
	defaultSessionsOnce.Do(func() {
		dir := sessionDirFromEnv()
		_ = os.MkdirAll(dir, 0o755)
		defaultSessions = NewSessionManager(dir, sessionMaxAgeFromEnv())
	})
	return defaultSessions
}

// NewSessionManager 创建自定义实例。
func NewSessionManager(dir string, maxAge time.Duration) *SessionManager {
	if dir == "" {
		dir = sessionDirFromEnv()
	}
	if maxAge <= 0 {
		maxAge = 30 * time.Minute
	}
	return &SessionManager{
		dir:    dir,
		maxAge: maxAge,
		cache:  make(map[string]*Session),
	}
}

// GetOrLogin 返回缓存 Session,不存在时调用 fetch 登录。
func (m *SessionManager) GetOrLogin(ctx context.Context, key SessionKey, fetch SessionFetcher) (*Session, error) {
	id := key.fileID()
	if id == "" || fetch == nil {
		return fetch(ctx)
	}

	if sess := m.loadFromMemory(id); sess != nil {
		return cloneSession(sess), nil
	}

	if sess, err := m.loadFromDisk(id); err != nil {
		return nil, err
	} else if sess != nil {
		m.saveToMemory(id, sess)
		return cloneSession(sess), nil
	}

	fresh, err := fetch(ctx)
	if err != nil {
		return nil, err
	}
	freshCopy := *fresh
	freshCopy.UpdatedAt = time.Now()
	m.saveToMemory(id, &freshCopy)
	if err := m.writeToDisk(id, &freshCopy); err != nil {
		return nil, err
	}
	return cloneSession(&freshCopy), nil
}

// Invalidate 删除缓存文件,迫使下次登录。
func (m *SessionManager) Invalidate(key SessionKey) {
	id := key.fileID()
	if id == "" {
		return
	}
	m.mu.Lock()
	delete(m.cache, id)
	m.mu.Unlock()
	os.Remove(m.filePath(id)) // nolint:errcheck
}

func (m *SessionManager) loadFromMemory(id string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sess, ok := m.cache[id]; ok {
		if !m.isExpired(sess) {
			return sess
		}
		delete(m.cache, id)
	}
	return nil
}

func (m *SessionManager) saveToMemory(id string, sess *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[id] = sess
}

func (m *SessionManager) isExpired(sess *Session) bool {
	if m.maxAge <= 0 {
		return false
	}
	if sess.UpdatedAt.IsZero() {
		return true
	}
	return time.Since(sess.UpdatedAt) > m.maxAge
}

func (m *SessionManager) loadFromDisk(id string) (*Session, error) {
	path := m.filePath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	if m.isExpired(&sess) {
		os.Remove(path) // nolint:errcheck
		return nil, nil
	}
	return &sess, nil
}

func (m *SessionManager) writeToDisk(id string, sess *Session) error {
	path := m.filePath(id)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0o600)
}

func (m *SessionManager) filePath(id string) string {
	return filepath.Join(m.dir, id+".json")
}

func cloneSession(s *Session) *Session {
	if s == nil {
		return nil
	}
	copy := *s
	return &copy
}

func (key SessionKey) fileID() string {
	service := sanitizeKey(key.Service)
	user := sanitizeKey(key.Username)
	if service == "" || user == "" {
		return ""
	}
	base := strings.TrimSpace(strings.ToLower(key.BaseURL))
	if base == "" {
		base = "local"
	}
	h := sha1.Sum([]byte(base))
	hash := hex.EncodeToString(h[:8])
	return fmt.Sprintf("%s_%s_%s", service, user, hash)
}

func sanitizeKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ToLower(value)
	return sanitizeFilename(value)
}

func sanitizeFilename(value string) string {
	replacer := func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}
	return strings.Map(replacer, value)
}

func sessionDirFromEnv() string {
	if dir := os.Getenv("TSU_SESSION_DIR"); dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".cache", "tsu-sessions")
	}
	return filepath.Join(os.TempDir(), "tsu-sessions")
}

func sessionMaxAgeFromEnv() time.Duration {
	if raw := os.Getenv("TSU_SESSION_MAX_AGE"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			return d
		}
	}
	return 30 * time.Minute
}

// EnsureAdminSession 复用 root 登录态。
func EnsureAdminSession(ctx context.Context, client *Client, cfg Config) (Session, error) {
	manager := DefaultSessionManager()
	key := SessionKey{Service: "admin", Username: cfg.AdminUsername, BaseURL: cfg.BaseURL}
	sess, err := manager.GetOrLogin(ctx, key, func(ctx context.Context) (*Session, error) {
		loginReq := LoginRequest{Identifier: cfg.AdminUsername, Password: cfg.AdminPassword}
		resp, httpResp, raw, err := PostJSON[LoginRequest, LoginResponse](ctx, client, "/api/v1/admin/auth/login", loginReq, "")
		if err != nil {
			return nil, fmt.Errorf("admin login failed: %w", err)
		}
		if httpResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("admin login status=%d body=%s", httpResp.StatusCode, string(raw))
		}
		if resp.Code != int(xerrors.CodeSuccess) || resp.Data == nil {
			return nil, fmt.Errorf("admin login unexpected body=%s", string(raw))
		}
		data := resp.Data
		return &Session{
			Token:    data.SessionToken,
			UserID:   data.UserID,
			Email:    data.Email,
			Username: data.Username,
		}, nil
	})
	if err != nil {
		return Session{}, err
	}
	return *sess, nil
}

// EnsureGameSession 复用已存在玩家登录态。
func EnsureGameSession(ctx context.Context, client *Client, baseURL, username, password string) (Session, error) {
	if username == "" || password == "" {
		return Session{}, fmt.Errorf("username/password required for session reuse")
	}
	manager := DefaultSessionManager()
	key := SessionKey{Service: "game", Username: username, BaseURL: baseURL}
	sess, err := manager.GetOrLogin(ctx, key, func(ctx context.Context) (*Session, error) {
		loginReq := LoginRequest{Identifier: username, Password: password}
		resp, httpResp, raw, err := PostJSON[LoginRequest, LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
		if err != nil {
			return nil, fmt.Errorf("game login failed: %w", err)
		}
		if httpResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("game login status=%d body=%s", httpResp.StatusCode, string(raw))
		}
		if resp.Code != int(xerrors.CodeSuccess) || resp.Data == nil {
			return nil, fmt.Errorf("game login unexpected body=%s", string(raw))
		}
		data := resp.Data
		return &Session{
			Token:    data.SessionToken,
			UserID:   data.UserID,
			Email:    data.Email,
			Username: data.Username,
		}, nil
	})
	if err != nil {
		return Session{}, err
	}
	return *sess, nil
}
