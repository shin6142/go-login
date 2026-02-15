package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Cookieを設定するエンドポイント
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{
			Name:     "my_cookie",
			Value:    "hello_from_server",
			Path:     "/",
			HttpOnly: true,                           // JavaScriptからアクセス不可
			Expires:  time.Now().Add(24 * time.Hour), // 24時間後に期限切れ
		}
		http.SetCookie(w, cookie)
		fmt.Fprintln(w, "Cookieを設定しました: my_cookie=hello_from_server")
	})

	// Cookieを読み取るエンドポイント
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("my_cookie")
		if err != nil {
			fmt.Fprintln(w, "Cookieが見つかりません")
			return
		}
		fmt.Fprintf(w, "受け取ったCookie: %s=%s\n", cookie.Name, cookie.Value)
	})

	// Cookieを削除するエンドポイント
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{
			Name:    "my_cookie",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0), // 過去の日時 = 削除
			MaxAge:  -1,
		}
		http.SetCookie(w, cookie)
		fmt.Fprintln(w, "Cookieを削除しました")
	})

	fmt.Println("=== Cookie デモサーバー ===")
	fmt.Println("http://localhost:3000 で起動中...")
	fmt.Println()
	fmt.Println("使い方 (01_cookie_demo ディレクトリから実行):")
	fmt.Println("  1. curl -c ./cookies.txt http://localhost:3000/set  # Cookieを保存")
	fmt.Println("  2. curl -b ./cookies.txt http://localhost:3000/get  # Cookieを送信")
	fmt.Println("  3. curl -b ./cookies.txt -c ./cookies.txt http://localhost:3000/delete  # Cookie削除")
	fmt.Println()

	http.ListenAndServe(":3000", nil)
}
