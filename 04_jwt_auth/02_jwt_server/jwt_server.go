package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ===================
// 設定
// ===================

// 秘密鍵（本番環境では環境変数から取得すること！）
var secretKey = []byte("my-super-secret-key-12345")

// トークンの有効期限
const tokenExpiration = 1 * time.Hour

// ===================
// JWT関連
// ===================

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Payload struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func createSignature(header, payload string) string {
	message := header + "." + payload
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	return base64URLEncode(h.Sum(nil))
}

func generateJWT(username string) (string, error) {
	header := Header{Alg: "HS256", Typ: "JWT"}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64URLEncode(headerJSON)

	now := time.Now()
	payload := Payload{
		Username: username,
		Exp:      now.Add(tokenExpiration).Unix(),
		Iat:      now.Unix(),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadEncoded := base64URLEncode(payloadJSON)

	signature := createSignature(headerEncoded, payloadEncoded)

	return headerEncoded + "." + payloadEncoded + "." + signature, nil
}

func verifyJWT(token string) (*Payload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("無効なトークン形式")
	}

	// 署名を検証
	signatureExpected := createSignature(parts[0], parts[1])
	if parts[2] != signatureExpected {
		return nil, fmt.Errorf("署名が無効です")
	}

	// Payloadをデコード
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Payloadのデコードに失敗")
	}

	var payload Payload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("Payloadのパースに失敗")
	}

	// 有効期限をチェック
	if time.Now().Unix() > payload.Exp {
		return nil, fmt.Errorf("トークンが期限切れです")
	}

	return &payload, nil
}

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
// HTTPハンドラー
// ===================

type Server struct {
	users *UserStore
	// セッション方式と違い、セッションストアがない！
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // JWTトークンを返す
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Authorizationヘッダーからトークンを取得
func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorizationヘッダーがありません")
	}

	// "Bearer <token>" の形式をパース
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("無効なAuthorizationヘッダー形式")
	}

	return parts[1], nil
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
	jsonResponse(w, http.StatusCreated, Response{true, "登録しました"})
}

// ログイン → JWTを発行
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

	// JWT生成（セッション方式と違い、サーバーに保存しない！）
	token, err := generateJWT(user.Username)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{false, "トークン生成に失敗"})
		return
	}

	log.Printf("ログイン成功: %s", user.Username)
	jsonResponse(w, http.StatusOK, LoginResponse{
		Success: true,
		Message: fmt.Sprintf("ようこそ、%s さん！", user.Username),
		Token:   token,
	})
}

// プロフィール（認証が必要）
func (s *Server) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Authorizationヘッダーからトークンを取得
	token, err := extractToken(r)
	if err != nil {
		jsonResponse(w, http.StatusUnauthorized, Response{false, err.Error()})
		return
	}

	// トークンを検証
	payload, err := verifyJWT(token)
	if err != nil {
		jsonResponse(w, http.StatusUnauthorized, Response{false, err.Error()})
		return
	}

	// 認証成功！
	jsonResponse(w, http.StatusOK, Response{true, fmt.Sprintf("こんにちは、%s さん！", payload.Username)})
}

func main() {
	server := &Server{
		users: NewUserStore(),
		// セッションストアがない！ステートレス！
	}

	http.HandleFunc("/register", server.HandleRegister)
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/profile", server.HandleProfile)

	fmt.Println("=== JWT認証サーバー ===")
	fmt.Println("http://localhost:3000 で起動中...")
	fmt.Println()
	fmt.Println("使い方 (04_jwt_auth ディレクトリから実行):")
	fmt.Println("  1. curl -X POST http://localhost:3000/register -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println("  2. curl -X POST http://localhost:3000/login -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println("     → 返ってきた token をコピー")
	fmt.Println("  3. curl -H 'Authorization: Bearer <token>' http://localhost:3000/profile")
	fmt.Println()
	fmt.Println("【セッション方式との違い】")
	fmt.Println("  - Cookieを使わない → Authorizationヘッダーで送信")
	fmt.Println("  - サーバーにセッション情報を保存しない → ステートレス")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":3000", nil))
}
