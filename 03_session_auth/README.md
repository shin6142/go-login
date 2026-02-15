# Phase 2: セッション認証

## 学習目標
- Cookieの仕組みを理解する
- セッション管理を自作する
- ログイン状態の維持を実装する

---

## ファイル構成

```
03_session_auth/
├── README.md
├── 01_cookie_demo/          # Phase 2-1
│   ├── cookie_demo.go
│   └── cookies.txt          # curlで生成されるCookie保存ファイル
└── 02_session_server/       # Phase 2-2
    ├── session_server.go
    └── cookies.txt          # curlで生成されるCookie保存ファイル
```

### cookies.txt とは？

curlコマンドで **ブラウザのCookie保存領域を模擬** するためのファイル。

| curlオプション | 意味 | ブラウザでの動作 |
|----------------|------|-----------------|
| `-c cookies.txt` | Cookieを保存 | Set-Cookieを受け取り保存 |
| `-b cookies.txt` | Cookieを送信 | リクエスト時に自動でCookie送信 |

```bash
# ログイン → サーバーがSet-Cookieを返す → cookies.txtに保存
curl -c ./cookies.txt -X POST http://localhost:3000/login -d '...'

# 次のリクエスト → cookies.txtからCookieを読み取り送信
curl -b ./cookies.txt http://localhost:3000/profile
```

**注意**: cookies.txtは `.gitignore` に含まれているため、Gitには追跡されない。

---

## Phase 2-1: Cookieの仕組み

### Cookieとは
- サーバーがブラウザに保存させる小さなデータ
- ブラウザが自動で送ってくれる（開発者が毎回書く必要なし）

### Cookie vs Authorization: Bearer

| | Cookie | Authorization: Bearer |
|--|--------|----------------------|
| 送り方 | ブラウザが**自動で**送る | 開発者が**手動で**コードに書く |
| 設定場所 | サーバーが `Set-Cookie` で指定 | JavaScriptなどで明示的に設定 |

### Cookieの属性

| 属性 | 説明 |
|------|------|
| `HttpOnly` | JavaScriptからアクセス不可（XSS対策） |
| `Secure` | HTTPS通信でのみ送信 |
| `SameSite` | クロスサイトリクエストでの送信を制限（CSRF対策） |
| `Expires` | 有効期限 |
| `Path` | Cookieが送られるパス |

### HttpOnlyの仕組み

```
1. サーバー: Set-Cookie: session_id=xxx; HttpOnly
2. ブラウザ: 「HttpOnlyか。JSからアクセス禁止にしよう」
3. JavaScript: document.cookie → session_id は見えない
```

**ブラウザがルールを強制する**。サーバーは属性を指定するだけ。

---

## Phase 2-2: セッション管理

### セッション認証の流れ

```
1. POST /login (username, password)
2. サーバー: パスワード検証 → セッションID生成 → メモリに保存
3. レスポンス: Set-Cookie: session_id=xxx
4. 以降のリクエスト: Cookie: session_id=xxx → サーバーがメモリで検索
```

### セッションIDの生成

```go
func generateSessionID() (string, error) {
    b := make([]byte, 32)
    rand.Read(b)  // crypto/rand を使用
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**重要: `crypto/rand` を使う理由** → 下記Q&A参照

### エンドポイント

| エンドポイント | 説明 |
|----------------|------|
| POST /register | ユーザー登録 |
| POST /login | ログイン → セッション作成 → Cookie送信 |
| GET /profile | 認証が必要なページ |
| POST /logout | セッション削除 |

---

## Q&A

### Q: なぜ HttpOnly: true にする？
XSS攻撃でセッションIDを盗まれることを防ぐため。JavaScriptから `document.cookie` でアクセスできなくなる。

### Q: 盗まれた後は防げる？
防げない。だから「盗まれないようにする」のがHttpOnly。

### Q: なぜ `crypto/rand` を使う？ `math/rand` ではダメ？

| | `math/rand` | `crypto/rand` |
|--|-------------|---------------|
| 用途 | ゲーム、シミュレーション | セキュリティ |
| 予測可能性 | 予測可能 | 予測不可能 |

`math/rand` は「シード」から乱数列が決まるため、攻撃者がシードを推測できれば予測可能。

### Q: シードとは？
乱数を生成する「出発点」となる数値。

```go
rand.Seed(12345)
fmt.Println(rand.Int())  // 常に同じ値が出る！

rand.Seed(12345)
fmt.Println(rand.Int())  // また同じ値！
```

**同じシード = 同じ乱数列**。`time.Now()` をシードにすると、ログイン時刻から推測される危険がある。

### Q: サーバーが複数台あると何が問題？

```
ユーザー → サーバーA（ログイン、セッション保存）
ユーザー → サーバーB（次のリクエスト）
         → サーバーBにはセッション情報がない！認証失敗
```

**解決策:**
- スティッキーセッション: 同じユーザーを常に同じサーバーに振り分け
- 共有ストレージ: Redis等でセッション情報を共有
- JWT: サーバーが状態を持たない → この問題が起きない（Phase 3で学ぶ）
