package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// ユーザー情報を保存する構造体
type User struct {
	Username     string
	PasswordHash []byte // ハッシュ化されたパスワードを保存（平文は保存しない！）
}

// インメモリのユーザーストア
type UserStore struct {
	users map[string]*User // key: username, value: User
}

// 新しいUserStoreを作成
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*User),
	}
}

// ユーザー登録
func (s *UserStore) Register(username, password string) error {
	// 既にユーザーが存在するかチェック
	if _, exists := s.users[username]; exists {
		return fmt.Errorf("ユーザー '%s' は既に存在します", username)
	}

	// パスワードをハッシュ化（Phase 1-1で学んだ！）
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	// ユーザーを保存
	s.users[username] = &User{
		Username:     username,
		PasswordHash: hash,
	}

	return nil
}

// ログイン（パスワード検証）
func (s *UserStore) Login(username, password string) error {
	// ユーザーを検索
	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("ユーザーが見つかりません")
	}

	// パスワードを検証（Phase 1-1で学んだ！）
	err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		return fmt.Errorf("パスワードが間違っています")
	}

	return nil
}

// --- HTTPハンドラー ---

// リクエストボディの構造体
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// レスポンスの構造体
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// JSONレスポンスを返すヘルパー関数
func jsonResponse(w http.ResponseWriter, status int, resp AuthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// 登録ハンドラー
func (s *UserStore) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "POSTメソッドを使用してください",
		})
		return
	}

	// リクエストボディをパース
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "無効なJSONです",
		})
		return
	}

	// バリデーション
	if req.Username == "" || req.Password == "" {
		jsonResponse(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "ユーザー名とパスワードは必須です",
		})
		return
	}

	// ユーザー登録
	if err := s.Register(req.Username, req.Password); err != nil {
		jsonResponse(w, http.StatusConflict, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	log.Printf("ユーザー登録成功: %s", req.Username)
	jsonResponse(w, http.StatusCreated, AuthResponse{
		Success: true,
		Message: fmt.Sprintf("ユーザー '%s' を登録しました", req.Username),
	})
}

// ログインハンドラー
func (s *UserStore) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "POSTメソッドを使用してください",
		})
		return
	}

	// リクエストボディをパース
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "無効なJSONです",
		})
		return
	}

	// ログイン
	if err := s.Login(req.Username, req.Password); err != nil {
		jsonResponse(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	log.Printf("ログイン成功: %s", req.Username)
	jsonResponse(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: fmt.Sprintf("ようこそ、%s さん！", req.Username),
	})
}

func main() {
	store := NewUserStore()

	http.HandleFunc("/register", store.HandleRegister)
	http.HandleFunc("/login", store.HandleLogin)

	fmt.Println("=== インメモリ認証サーバー ===")
	fmt.Println("http://localhost:3000 で起動中...")
	fmt.Println()
	fmt.Println("使い方:")
	fmt.Println("  登録: curl -X POST http://localhost:3000/register -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println("  ログイン: curl -X POST http://localhost:3000/login -d '{\"username\":\"testuser\",\"password\":\"secret123\"}'")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":3000", nil))
}
