# Go ログイン機構 学習リポジトリ

ログイン認証の仕組みを、できるだけ低レベルから自分で実装して理解するための学習リポジトリ。

## 学習の進め方

各フェーズごとにディレクトリを分けて、段階的に学習を進める。

## 進捗

- [x] Phase 1: 基礎
- [x] Phase 2: セッション認証
- [ ] Phase 3: JWT認証
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
├── 04_jwt_auth/             # Phase 3: JWT認証（予定）
│
└── 05_with_db/              # Phase 4: DB連携（予定）
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
