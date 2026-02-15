# Phase 3: JWT認証

## 学習目標
- JWTの構造（Header, Payload, Signature）を理解する
- 署名による改ざん検出の仕組みを理解する
- JWT認証サーバーを自作する

---

## ファイル構成

```
04_jwt_auth/
├── README.md
├── 01_jwt_demo/           # Phase 3-1: JWTの構造理解
│   └── jwt_demo.go        # JWT生成・検証のデモ
└── 02_jwt_server/         # Phase 3-2: JWT認証サーバー
    └── jwt_server.go      # 登録・ログイン・認証API
```

---

## Phase 3-1: JWTの構造

### JWTとは

**JSON Web Token** の略。認証情報を安全にやり取りするための規格（RFC 7519）。

3つの部分が `.`（ドット）で区切られている：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRhcm8iLCJleHAiOjE3MDAwMDAwMDB9.X7pP2kQrVmZ8K9sF...
└─────────── Header ───────────┘.└──────────────── Payload ──────────────────┘.└─── Signature ───┘
```

---

### 1. Header（ヘッダー）

**署名アルゴリズムとトークンタイプを指定**

```json
{
  "alg": "HS256",    // 署名アルゴリズム
  "typ": "JWT"       // トークンタイプ
}
```

#### 主な署名アルゴリズム

| アルゴリズム | 種類 | 特徴 |
|-------------|------|------|
| HS256 | 共通鍵（HMAC） | シンプル、同じ鍵で署名・検証 |
| RS256 | 公開鍵（RSA） | 秘密鍵で署名、公開鍵で検証 |
| ES256 | 公開鍵（ECDSA） | RS256より高速・短い鍵長 |

**HS256の仕組み:**
```
サーバーA（署名）: 秘密鍵で署名 → JWT発行
サーバーA（検証）: 同じ秘密鍵で検証
→ 秘密鍵を知っているサーバーだけが署名・検証できる
```

**RS256の仕組み:**
```
認証サーバー: 秘密鍵で署名 → JWT発行
APIサーバー: 公開鍵で検証（秘密鍵不要）
→ 公開鍵は公開できるので、外部サービスでも検証可能
```

---

### 2. Payload（ペイロード）

**実際のデータ（クレーム）を格納**

```json
{
  "sub": "user123",           // Subject: ユーザーID
  "username": "taro",         // カスタムクレーム
  "role": "admin",            // カスタムクレーム
  "iat": 1700000000,          // Issued At: 発行日時
  "exp": 1700003600           // Expiration: 有効期限
}
```

#### 標準クレーム（Registered Claims）

| クレーム | 名前 | 説明 |
|----------|------|------|
| `iss` | Issuer | 発行者（例: "https://auth.example.com"） |
| `sub` | Subject | 対象者（通常はユーザーID） |
| `aud` | Audience | 受信者（このトークンを使うサービス） |
| `exp` | Expiration | 有効期限（Unix時間） |
| `iat` | Issued At | 発行日時（Unix時間） |
| `nbf` | Not Before | この時刻以前は無効 |
| `jti` | JWT ID | トークンの一意識別子 |

#### Unix時間とは？

1970年1月1日 00:00:00 UTCからの経過秒数

```
1700000000 = 2023年11月14日 22:13:20 UTC
```

```go
// Goでの変換
time.Unix(1700000000, 0)  // → time.Time型

// 現在時刻をUnix時間に
time.Now().Unix()  // → int64
```

#### Payloadに入れるもの・入れないもの

| 入れる | 入れない |
|--------|----------|
| ユーザーID（sub） | パスワード |
| 有効期限（exp） | 機密情報（クレジットカード番号等） |
| 発行日時（iat） | 大きなデータ（画像等） |
| 権限（role） | 頻繁に変わる情報 |
| ユーザー名 | |

**なぜ機密情報を入れない？** → PayloadはBase64URLデコードで誰でも読める

---

### 3. Signature（署名）

**改ざん検出のための署名**

```
Signature = HMAC-SHA256(
    Base64URL(Header) + "." + Base64URL(Payload),
    秘密鍵
)
```

#### 署名生成の流れ

```
1. Header JSON → Base64URLエンコード → eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
2. Payload JSON → Base64URLエンコード → eyJ1c2VybmFtZSI6InRhcm8ifQ
3. 上記を "." で連結 → eyJhbGci...eyJ1c2Vy...
4. 秘密鍵でHMAC-SHA256 → バイナリデータ
5. Base64URLエンコード → X7pP2kQrVmZ8K9sF...
6. 全部連結 → 完成したJWT
```

---

### Base64URL エンコードとは

**バイナリデータを安全な文字列に変換**

通常のBase64との違い：

| | Base64 | Base64URL |
|--|--------|-----------|
| 62番目の文字 | `+` | `-` |
| 63番目の文字 | `/` | `_` |
| パディング | `=` あり | `=` なし |

**なぜBase64URLを使う？**
- `+`, `/`, `=` はURLで特別な意味を持つ
- URLやHTTPヘッダーで安全に使えるように

```go
// Goでの実装
import "encoding/base64"

// Base64URL エンコード（パディングなし）
encoded := base64.RawURLEncoding.EncodeToString(data)

// Base64URL デコード
decoded, _ := base64.RawURLEncoding.DecodeString(encoded)
```

---

### HMAC-SHA256 とは

**メッセージ認証コード（MAC）の一種**

```
HMAC = Hash-based Message Authentication Code
SHA256 = Secure Hash Algorithm（256ビット出力）
```

#### HMAC-SHA256 と秘密鍵の関係

**HMAC-SHA256は「アルゴリズム（関数）」であり、秘密鍵ではない**

```
HMAC-SHA256(メッセージ, 秘密鍵) → 署名
            ↑入力1      ↑入力2    ↑出力
```

| 用語 | 役割 | 例 |
|------|------|-----|
| **HMAC-SHA256** | 計算方法（関数） | 「足し算」のようなもの |
| **秘密鍵** | 関数への入力値 | `"my-secret-key-123"` |
| **署名** | 関数の出力値 | `X7pP2kQrVmZ8K9sF...` |

```
┌─────────────────────────────────┐
│         HMAC-SHA256             │  ← アルゴリズム（計算方法）
│  ┌─────────┐  ┌─────────────┐   │
│  │メッセージ │  │  秘密鍵     │   │  ← 2つの入力
│  │ "Hello" │  │ "my-secret" │   │
│  └────┬────┘  └──────┬──────┘   │
│       └──────┬───────┘          │
│              ▼                  │
│      ┌─────────────┐            │
│      │   署名出力   │            │  ← 出力
│      │ X7pP2kQr... │            │
│      └─────────────┘            │
└─────────────────────────────────┘
```

**秘密鍵は「材料」、HMAC-SHA256は「レシピ」**

#### HMACの仕組み

```
HMAC(メッセージ, 秘密鍵) → 固定長のハッシュ値（256ビット = 32バイト）
```

**特徴:**
- 同じメッセージ + 同じ鍵 = 同じハッシュ値
- メッセージが1ビットでも変わる → 全く違うハッシュ値
- 秘密鍵がないと正しいハッシュ値を計算できない

```go
import (
    "crypto/hmac"
    "crypto/sha256"
)

func sign(message, secretKey string) []byte {
    h := hmac.New(sha256.New, []byte(secretKey))
    h.Write([]byte(message))
    return h.Sum(nil)  // 32バイトのハッシュ値
}
```

---

### 署名検証の流れ

```
【受信したJWT】
eyJhbGci...eyJ1c2Vy...X7pP2kQr...
    ↓
【分解】
Header: eyJhbGci...
Payload: eyJ1c2Vy...
Signature: X7pP2kQr...
    ↓
【再計算】
HMAC-SHA256(Header + "." + Payload, 秘密鍵) → 新しいSignature
    ↓
【比較】
新しいSignature == 受信したSignature ?
  → 一致: 有効なJWT（改ざんされていない）
  → 不一致: 無効なJWT（改ざんされた or 秘密鍵が違う）
```

---

### 改ざん検出の具体例

```
【正規のJWT】
Header.Payload.Signature
eyJhbGciOi...eyJ1c2VybmFtZSI6InRhcm8ifQ.validSignature

【攻撃者がPayloadを改ざん】
{"username":"taro"} → {"username":"admin"}
                            ↓
新しいPayload: eyJ1c2VybmFtZSI6ImFkbWluIn0

【攻撃者が送るJWT】
eyJhbGciOi...eyJ1c2VybmFtZSI6ImFkbWluIn0.validSignature
                    ↑改ざん                  ↑古い署名

【サーバーで検証】
HMAC(Header + "." + 改ざんPayload, 秘密鍵) → newSignature
newSignature != validSignature → 検証失敗！
```

**攻撃者は秘密鍵を知らないので、正しい署名を作れない**

---

## Phase 3-2: JWT認証サーバー

### セッション方式との違い

| | セッション方式 | JWT方式 |
|--|---------------|---------|
| 認証情報の保存 | サーバーのメモリ/Redis | クライアント側（トークン自体） |
| トークン送信方法 | Cookie | Authorizationヘッダー |
| 検証方法 | DBで検索 | 署名を検証 |
| スケーラビリティ | セッション共有が必要 | サーバー間共有不要 |

### 使い方

```bash
cd 02_jwt_server
go run jwt_server.go

# 1. ユーザー登録
curl -X POST http://localhost:3000/register \
  -d '{"username":"testuser","password":"secret123"}'

# 2. ログイン → トークン取得
curl -X POST http://localhost:3000/login \
  -d '{"username":"testuser","password":"secret123"}'
# → {"token":"eyJhbGciOiJIUzI1NiIs..."} が返る

# 3. 認証が必要なAPIにアクセス
curl -H 'Authorization: Bearer <token>' http://localhost:3000/profile
```

---

## セッション方式 vs JWT方式の使い分け

### セッション方式が向いているケース

| ケース | 理由 |
|--------|------|
| 銀行・金融系 | 即座にログアウト（無効化）が必要 |
| 管理画面 | セキュリティ重視、少人数利用 |
| 同時ログイン制限 | サーバーでセッション数を管理できる |

### JWT方式が向いているケース

| ケース | 理由 |
|--------|------|
| マイクロサービス | サービス間で状態共有不要 |
| モバイルアプリ | Authorizationヘッダーが自然 |
| 大規模サービス | スケーラビリティが高い |
| OAuth/外部連携 | トークンで情報を渡せる |

### 図解

```
【セッション方式】
サーバーA ─┐
サーバーB ─┼──→ Redis（セッション保存）
サーバーC ─┘
→ 全サーバーがRedisにアクセスして検証

【JWT方式】
サーバーA → 署名検証（独立）
サーバーB → 署名検証（独立）
サーバーC → 署名検証（独立）
→ 各サーバーが独立して検証可能
```

---

## Q&A

### Q: JWTのPayloadは誰でも見れるのに、なぜ安全？
「見れる」と「改ざんできる」は違う。署名があるため、Payloadを書き換えると検証に失敗する。

### Q: 秘密鍵が漏洩したら？
攻撃者が有効なJWTを自由に作れる → なりすまし可能。秘密鍵の管理が超重要。

### Q: JWTを即座に無効化できない？
できない（ステートレスなので）。対策：
- 有効期限を短くする（15分〜1時間）
- リフレッシュトークンをDB管理（無効化可能）
- ブラックリスト方式（無効化したJWTをDBに記録）

### Q: リフレッシュトークンとは？
短期のアクセストークン（JWT）と、長期のリフレッシュトークンを組み合わせる方式。
```
アクセストークン: JWT、15分有効、APIアクセスに使用
リフレッシュトークン: DB管理、7日有効、新しいアクセストークン取得に使用
```
