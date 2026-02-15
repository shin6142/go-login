package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// 秘密鍵（本番環境では環境変数等から取得）
var secretKey = []byte("my-super-secret-key-12345")

// JWTのHeader
type Header struct {
	Alg string `json:"alg"` // アルゴリズム
	Typ string `json:"typ"` // タイプ
}

// JWTのPayload
type Payload struct {
	Username string `json:"username"` // ユーザー名
	Exp      int64  `json:"exp"`      // 有効期限（Unix時間）
	Iat      int64  `json:"iat"`      // 発行日時（Unix時間）
}

// Base64URLエンコード（JWTはBase64URLを使用）
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLデコード
func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// HMAC-SHA256で署名を生成
func createSignature(header, payload string) string {
	message := header + "." + payload
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	return base64URLEncode(h.Sum(nil))
}

// JWTを生成
func generateJWT(username string, duration time.Duration) (string, error) {
	// 1. Headerを作成
	header := Header{
		Alg: "HS256",
		Typ: "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64URLEncode(headerJSON)

	// 2. Payloadを作成
	now := time.Now()
	payload := Payload{
		Username: username,
		Exp:      now.Add(duration).Unix(),
		Iat:      now.Unix(),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadEncoded := base64URLEncode(payloadJSON)

	// 3. 署名を生成
	signature := createSignature(headerEncoded, payloadEncoded)

	// 4. JWTを組み立て
	jwt := headerEncoded + "." + payloadEncoded + "." + signature

	return jwt, nil
}

// JWTを検証してPayloadを取得
func verifyJWT(token string) (*Payload, error) {
	// 1. JWTを3つの部分に分割
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("無効なトークン形式")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureReceived := parts[2]

	// 2. 署名を検証
	signatureExpected := createSignature(headerEncoded, payloadEncoded)
	if signatureReceived != signatureExpected {
		return nil, fmt.Errorf("署名が一致しません（改ざんの可能性）")
	}

	// 3. Payloadをデコード
	payloadJSON, err := base64URLDecode(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("Payloadのデコードに失敗: %w", err)
	}

	var payload Payload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("Payloadのパースに失敗: %w", err)
	}

	// 4. 有効期限をチェック
	if time.Now().Unix() > payload.Exp {
		return nil, fmt.Errorf("トークンが期限切れです")
	}

	return &payload, nil
}

func main() {
	fmt.Println("=== JWT デモ ===")
	fmt.Println()

	// 1. JWTを生成
	token, _ := generateJWT("taro", 1*time.Hour)
	fmt.Println("【生成されたJWT】")
	fmt.Println(token)
	fmt.Println()

	// 2. JWTの構造を確認
	parts := strings.Split(token, ".")
	fmt.Println("【JWTの構造】")

	headerJSON, _ := base64URLDecode(parts[0])
	fmt.Printf("Header:    %s\n", string(headerJSON))

	payloadJSON, _ := base64URLDecode(parts[1])
	fmt.Printf("Payload:   %s\n", string(payloadJSON))

	fmt.Printf("Signature: %s\n", parts[2])
	fmt.Println()

	// 3. 正常なトークンを検証
	fmt.Println("【正常なトークンを検証】")
	payload, err := verifyJWT(token)
	if err != nil {
		fmt.Printf("検証失敗: %s\n", err)
	} else {
		fmt.Printf("検証成功！ ユーザー: %s\n", payload.Username)
	}
	fmt.Println()

	// 4. 改ざんされたトークンを検証
	fmt.Println("【改ざんされたトークンを検証】")
	// Payloadを改ざん（taroをadminに変更）
	tamperedPayload := `{"username":"admin","exp":9999999999,"iat":1234567890}`
	tamperedPayloadEncoded := base64URLEncode([]byte(tamperedPayload))
	tamperedToken := parts[0] + "." + tamperedPayloadEncoded + "." + parts[2]

	fmt.Printf("改ざんしたPayload: %s\n", tamperedPayload)
	_, err = verifyJWT(tamperedToken)
	if err != nil {
		fmt.Printf("検証失敗: %s ← 改ざんを検出！\n", err)
	}
	fmt.Println()

	// 5. 期限切れトークンを検証
	fmt.Println("【期限切れトークンを検証】")
	expiredToken, _ := generateJWT("taro", -1*time.Hour) // 1時間前に期限切れ
	_, err = verifyJWT(expiredToken)
	if err != nil {
		fmt.Printf("検証失敗: %s\n", err)
	}
}
