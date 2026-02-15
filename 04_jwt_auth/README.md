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

**JSON Web Token** の略。3つの部分からなる：

```
eyJhbGciOiJIUzI1NiIs...eyJ1c2VybmFtZSI6InRhcm8i...X7pP2kQrVmZ8K9sF...
└────── Header ──────┘└────── Payload ──────────┘└─── Signature ───┘
```

| 部分 | 内容 | エンコード |
|------|------|-----------|
| Header | アルゴリズム情報 | Base64URL |
| Payload | ユーザー情報、有効期限など | Base64URL |
| Signature | 改ざん防止の署名 | HMAC-SHA256 |

### 署名の仕組み

```
【生成】
Header + Payload + 秘密鍵 → HMAC-SHA256 → Signature

【検証】
受け取ったHeader + Payload + 秘密鍵 → 署名を再計算 → 一致すれば有効
```

**Payloadは誰でも見れるが、改ざんはできない**（署名が合わなくなる）

### Payloadに入れるもの

| 入れる | 入れない |
|--------|----------|
| ユーザーID | パスワード |
| 有効期限（exp） | 機密情報 |
| 発行日時（iat） | |
| 権限（role） | |

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
