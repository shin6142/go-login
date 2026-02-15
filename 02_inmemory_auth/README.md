# Phase 1-2: インメモリ認証

## 学習目標
- ユーザー登録機能の実装
- ログイン機能（パスワード検証）の実装
- HTTPサーバーでの認証エンドポイント

---

## ファイル

| ファイル | 内容 |
|----------|------|
| `auth_memory.go` | 登録・ログイン機能を持つHTTPサーバー |

---

## 実行方法

```bash
# サーバー起動
go run auth_memory.go

# ユーザー登録
curl -X POST http://localhost:3000/register \
  -d '{"username":"testuser","password":"secret123"}'

# ログイン
curl -X POST http://localhost:3000/login \
  -d '{"username":"testuser","password":"secret123"}'
```

---

## 学んだこと

### アーキテクチャ

```
POST /register  →  パスワードをハッシュ化してメモリに保存
POST /login     →  ハッシュと比較して検証
```

### コードの流れ

```go
// 登録時
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
users[username] = &User{PasswordHash: hash}

// ログイン時
err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
```

### インメモリ vs DB

| | インメモリ | DB |
|--|-----------|-----|
| データ保存 | RAM | ディスク |
| 再起動時 | 消える | 残る |
| 用途 | 学習・テスト | 本番環境 |

---

## 現状の問題点

**ログイン状態を維持できない**

現在のコードでは：
- 毎回ユーザー名とパスワードを送る必要がある
- 「ログイン済み」という状態を覚えていない

→ **Phase 2（セッション）で解決する**

---

## Q&A

（学習中に追加）
