# Phase 3-3: ブラウザストレージ

## 学習目標
- Cookie / localStorage / sessionStorage の違いを理解する
- タブ間でのデータ共有の仕組みを理解する
- 実際のアプリケーションでの使い分けを学ぶ

---

## ファイル構成

```
05_browser_storage/
├── README.md
└── storage_demo.go    # ブラウザで動作確認できるデモ
```

---

## 3つのストレージの比較

| | Cookie | localStorage | sessionStorage |
|--|--------|--------------|----------------|
| **有効期限** | 指定可能（Expires） | 永続（削除するまで） | タブを閉じるまで |
| **タブ間共有** | ✅ する | ✅ する | ❌ しない |
| **サーバーに送信** | ✅ 自動送信 | ❌ しない | ❌ しない |
| **容量** | 約4KB | 約5MB | 約5MB |
| **アクセス** | JS / サーバー両方 | JSのみ | JSのみ |

---

## 実際に保管する情報

### Cookie に保管するもの

```
【認証系】
- セッションID: session_id=abc123...
- CSRFトークン: csrf_token=xyz789...

【トラッキング系】
- Google Analytics: _ga=GA1.2.123456789
- 広告ID: _fbp=fb.1.1234567890

【ユーザー設定】
- 言語設定: lang=ja
- 同意フラグ: cookie_consent=true
```

**特徴**: サーバーに自動送信される → 認証に使う

---

### localStorage に保管するもの

```
【認証系】
- JWTトークン: token=eyJhbGciOiJIUzI1NiIs...
- リフレッシュトークン: refresh_token=...

【ユーザー設定】
- ダークモード: theme=dark
- 言語: locale=ja
- サイドバー開閉: sidebar=collapsed

【キャッシュ】
- APIレスポンス: cache_users=[{...}, {...}]
- 最終取得時刻: cache_timestamp=1700000000
```

**特徴**: 永続保存、容量大きい → 設定やキャッシュに使う

---

### sessionStorage に保管するもの

```
【一時的なデータ】
- フォーム入力途中: form_data={"name":"田中",...}
- ウィザードのステップ: wizard_step=2
- 検索フィルター: filters={"category":"electronics"}

【タブ固有の状態】
- スクロール位置: scroll_position=1500
- モーダルの開閉状態: modal_open=true
```

**特徴**: タブを閉じると消える → 一時的な状態に使う

---

## タブ間共有の図解

```
┌─────────────────────────────────────────────────┐
│                 ブラウザ                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │  タブA   │ │  タブB   │ │  タブC   │        │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘        │
│       │            │            │              │
│       └────────────┼────────────┘              │
│                    ▼                           │
│  ┌─────────────────────────────────────┐       │
│  │ Cookie / localStorage（共有）        │       │
│  │ → 全タブで同じデータを参照           │       │
│  └─────────────────────────────────────┘       │
│                                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │sessionA │ │sessionB │ │sessionC │          │
│  │(独立)   │ │(独立)   │ │(独立)   │          │
│  └─────────┘ └─────────┘ └─────────┘          │
└─────────────────────────────────────────────────┘
```

---

## JWT/セッションの保管場所

| 方式 | 保管場所 | メリット | デメリット |
|------|----------|----------|-----------|
| **セッション方式** | Cookie | XSS耐性（HttpOnly） | CSRF対策が必要 |
| **JWT方式（Cookie）** | Cookie | XSS耐性（HttpOnly） | CSRF対策が必要 |
| **JWT方式（localStorage）** | localStorage | CSRF影響なし | XSSで盗まれる可能性 |

### よくある構成

```
【セッション方式】
Cookie: session_id=abc123 (HttpOnly)

【JWT方式（推奨）】
Cookie: access_token=eyJ... (HttpOnly)

【JWT方式（SPA向け）】
localStorage: token=eyJ...
→ XSS対策が重要
```

---

## デモの使い方

```bash
cd 05_browser_storage
go run storage_demo.go
```

http://localhost:3000 を2つのタブで開いて実験：

| 操作 | 期待する結果 |
|------|-------------|
| タブAで **localStorage** に保存 | タブBで「読み取り」→ 見える ✅ |
| タブAで **sessionStorage** に保存 | タブBで「読み取り」→ 見えない ❌ |
| タブAで **Cookie** に保存 | タブBで「読み取り」→ 見える ✅ |

---

## Q&A

### Q: なぜ複数タブでログイン状態が維持される？

Cookie/localStorageが同一オリジンの全タブで共有されるから。

```
タブA: localStorage.getItem('token') → "eyJ..."
タブB: localStorage.getItem('token') → "eyJ..." （同じ値）
```

### Q: タブAでログアウトしたら、タブBも自動でログアウトする？

**実装次第**。ストレージからトークンは削除されるが、タブBの画面は更新されない。

```javascript
// storageイベントで他タブの変更を検知
window.addEventListener('storage', (event) => {
  if (event.key === 'token' && event.newValue === null) {
    // トークン削除を検知 → ログイン画面へ
    window.location.href = '/login';
  }
});
```

### Q: プロファイル変更が他タブに反映されないのはなぜ？

JWTのPayloadは発行時に固定される。プロファイル変更後もJWTは古いまま。

```
【解決策】
1. JWTを再発行する（サーバーから新しいトークンを取得）
2. APIから最新情報を取得する（JWTは認証のみに使用）
```

### Q: XSSとCSRFの違いは？

| 攻撃 | 手法 | 狙い |
|------|------|------|
| **XSS** | 悪意あるJSを実行させる | localStorage/Cookieを盗む |
| **CSRF** | 偽のリクエストを送らせる | Cookieが自動送信されることを悪用 |

```
【対策】
XSS対策: Cookie に HttpOnly を設定（JSからアクセス不可）
CSRF対策: CSRFトークンを使う（Phase 4で学習予定）
```
