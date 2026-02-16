# Go ログイン機構 学習リポジトリ

ログイン認証の仕組みを、できるだけ低レベルから自分で実装して理解するための学習リポジトリ。

## 学習の進め方

各フェーズごとにディレクトリを分けて、段階的に学習を進める。

## 進捗

- [x] Phase 1: 基礎
- [x] Phase 2: セッション認証
- [x] Phase 3-1, 3-2: JWT認証
- [x] Phase 3-3: ブラウザのストレージ管理
- [ ] Phase 4: 実践（DB連携、セキュリティ強化）

---

## ディレクトリ構成

```
go-login/
├── 01_hash_and_salt/        # Phase 1-1: パスワードのハッシュ化
│   ├── hash_demo.go         # bcryptのデモ
│   └── salt_demo.go         # ソルトの効果デモ
│
├── 02_inmemory_auth/        # Phase 1-2: インメモリ認証
│   └── auth_memory.go       # 登録・ログインサーバー
│
├── 03_session_auth/         # Phase 2: セッション認証
│   ├── 01_cookie_demo/      # Cookieの仕組み
│   └── 02_session_server/   # セッション管理サーバー
│
├── 04_jwt_auth/             # Phase 3-1, 3-2: JWT認証
│   ├── 01_jwt_demo/         # JWTの構造理解
│   └── 02_jwt_server/       # JWT認証サーバー
│
├── 05_browser_storage/      # Phase 3-3: ブラウザストレージ
│   └── storage_demo.go      # Cookie/localStorage/sessionStorageのデモ
│
└── 06_with_db/              # Phase 4: DB連携（予定）
```

---

## 学んだこと

### Phase 1: 基礎

| トピック | 学んだこと |
|----------|-----------|
| ハッシュ化 | パスワードは暗号化ではなくハッシュ化。一方向で復元不可能 |
| ソルト | 同じパスワードでも違うハッシュ値に。レインボーテーブル攻撃を防ぐ |
| bcrypt | パスワード用ハッシュ関数。わざと遅い設計で総当たり攻撃に強い |

### Phase 2: セッション認証

| トピック | 学んだこと |
|----------|-----------|
| Cookie | サーバーがブラウザに保存させるデータ。自動で送られる |
| HttpOnly | JSからアクセス不可にしてXSS攻撃を防ぐ |
| セッション | サーバー側でユーザー状態を管理。セッションIDをCookieで送る |
| crypto/rand | セキュリティ用の乱数生成。math/randは予測可能なので危険 |

### Phase 3: JWT認証

| トピック | 学んだこと |
|----------|-----------|
| JWT構造 | Header.Payload.Signature の3部構成 |
| 署名 | HMAC-SHA256で改ざん検出。秘密鍵がないと正しい署名を作れない |
| Base64URL | URLで安全に使える文字だけを使うエンコード |
| ステートレス | サーバーはセッションを保存しない。トークン自体に情報がある |

### Phase 3-3: ブラウザストレージ

| トピック | 学んだこと |
|----------|-----------|
| Cookie | サーバーに自動送信。認証系（セッションID、CSRF）に使う |
| localStorage | 永続保存、タブ間共有。JWT、ユーザー設定に使う |
| sessionStorage | タブを閉じると消える、タブ間共有なし。一時的な状態に使う |
| タブ間共有 | Cookie/localStorageは全タブで共有される |
| storageイベント | 他タブでのlocalStorage変更を検知できる |

#### 3つのストレージ比較

| | Cookie | localStorage | sessionStorage |
|--|--------|--------------|----------------|
| 有効期限 | 指定可能 | 永続 | タブを閉じるまで |
| タブ間共有 | ✅ | ✅ | ❌ |
| サーバー送信 | ✅ 自動 | ❌ | ❌ |
| 容量 | 約4KB | 約5MB | 約5MB |

#### なぜ複数タブでログインが維持される？
Cookie/localStorageが同一オリジンの全タブで共有されるから。

#### ログアウトで全タブがログアウトする？
**実装次第**。storageイベントで検知して画面更新が必要。

---

## 実行方法

各ディレクトリに移動してサーバーを起動：

```bash
# Phase 1-1: ハッシュ化デモ
cd 01_hash_and_salt
go run hash_demo.go

# Phase 2-2: セッション認証サーバー
cd 03_session_auth/02_session_server
go run session_server.go
```

詳細は各ディレクトリの `README.md` を参照。

---

## 技術スタック

- Go 1.21+
- golang.org/x/crypto/bcrypt
