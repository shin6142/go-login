package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ===================
// ユーザー管理
// ===================

type User struct {
	Username     string
	PasswordHash []byte
}

type UserStore struct {
	users map[string]*User
}

func NewUserStore() *UserStore {
	return &UserStore{users: make(map[string]*User)}
}

func (s *UserStore) Register(username, password string) error {
	if _, exists := s.users[username]; exists {
		return fmt.Errorf("ユーザー '%s' は既に存在します", username)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.users[username] = &User{Username: username, PasswordHash: hash}
	return nil
}

func (s *UserStore) Authenticate(username, password string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("ユーザーが見つかりません")
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, fmt.Errorf("パスワードが間違っています")
	}
	return user, nil
}

// ===================
// セッション管理
// ===================

type Session struct {
	ID        string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	sessions map[string]*Session // key: セッションID
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*Session)}
}

// ログ出力用
func (s *SessionStore) String() string {
	if len(s.sessions) == 0 {
		return "セッションなし"
	}
	result := fmt.Sprintf("セッション数: %d\n", len(s.sessions))
	for id, session := range s.sessions {
		result += fmt.Sprintf("  - ID: %s... / User: %s\n", id[:16], session.Username)
	}
	return result
}

// セッションIDを生成（32バイトのランダムな文字列）
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// 新しいセッションを作成
func (s *SessionStore) Create(username string) (*Session, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        id,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24時間有効
	}
	s.sessions[id] = session
	return session, nil
}

// セッションIDからセッションを取得
func (s *SessionStore) Get(id string) (*Session, error) {
	session, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("セッションが見つかりません")
	}
	// 有効期限チェック
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, id)
		return nil, fmt.Errorf("セッションが期限切れです")
	}
	return session, nil
}

// セッションを削除
func (s *SessionStore) Delete(id string) {
	delete(s.sessions, id)
}

// ===================
// HTTPハンドラー
// ===================

type Server struct {
	users    *UserStore
	sessions *SessionStore
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func jsonResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// ユーザー登録
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{false, "POSTメソッドを使用してください"})
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{false, "無効なJSONです"})
		return
	}

	if err := s.users.Register(req.Username, req.Password); err != nil {
		jsonResponse(w, http.StatusConflict, Response{false, err.Error()})
		return
	}

	log.Printf("ユーザー登録: %s", req.Username)
	log.Printf("セッション: %s", s.sessions)
	jsonResponse(w, http.StatusCreated, Response{true, "登録しました"})
}

// ログイン
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{false, "POSTメソッドを使用してください"})
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{false, "無効なJSONです"})
		return
	}

	// パスワード認証
	user, err := s.users.Authenticate(req.Username, req.Password)
	if err != nil {
		jsonResponse(w, http.StatusUnauthorized, Response{false, err.Error()})
		return
	}

	// セッション作成
	session, err := s.sessions.Create(user.Username)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{false, "セッション作成に失敗"})
		return
	}

	// CookieにセッションIDを設定
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Expires:  session.ExpiresAt,
	})

	log.Printf("ログイン成功: %s (セッションID: %s...)", user.Username, session.ID[:16])
	jsonResponse(w, http.StatusOK, Response{true, fmt.Sprintf("ようこそ、%s さん！", user.Username)})
}

// プロフィール（認証が必要）
func (s *Server) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// CookieからセッションIDを取得
	cookie, err := r.Cookie("session_id")
	if err != nil {
		jsonResponse(w, http.StatusUnauthorized, Response{false, "ログインしてください"})
		return
	}

	// セッションを検証
	session, err := s.sessions.Get(cookie.Value)
	if err != nil {
		jsonResponse(w, http.StatusUnauthorized, Response{false, err.Error()})
		return
	}

	// 認証成功！
	jsonResponse(w, http.StatusOK, Response{true, fmt.Sprintf("こんにちは、%s さん！", session.Username)})
}

// ログアウト
func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{false, "POSTメソッドを使用してください"})
		return
	}

	// CookieからセッションIDを取得
	cookie, err := r.Cookie("session_id")
	if err != nil {
		jsonResponse(w, http.StatusOK, Response{true, "既にログアウトしています"})
		return
	}

	// セッションを削除
	s.sessions.Delete(cookie.Value)

	// Cookieを削除
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})

	log.Printf("ログアウト: セッションID %s...", cookie.Value[:16])
	jsonResponse(w, http.StatusOK, Response{true, "ログアウトしました"})
}

func main() {
	server := &Server{
		users:    NewUserStore(),
		sessions: NewSessionStore(),
	}

	http.HandleFunc("/register", server.HandleRegister)
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/profile", server.HandleProfile)
	http.HandleFunc("/logout", server.HandleLogout)

	fmt.Println("=== セッション認証サーバー ===")
	fmt.Println("http://localhost:3000 で起動中...")
	fmt.Println()
	fmt.Println("使い方 (02_session_server ディレクトリから実行):")
	fmt.Println("  1. curl -X POST http://localhost:3000/register -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println("  2. curl -c ./cookies.txt -X POST http://localhost:3000/login -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println("  3. curl -b ./cookies.txt http://localhost:3000/profile")
	fmt.Println("  4. curl -b ./cookies.txt -X POST http://localhost:3000/logout")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":3000", nil))
}
