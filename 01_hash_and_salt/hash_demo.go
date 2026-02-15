package main

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"

	// パスワードをハッシュ化
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== ハッシュ化のデモ ===")
	fmt.Println("元のパスワード:", password)
	fmt.Println("ハッシュ値:", string(hash))
	fmt.Println()

	// 同じパスワードでもう一度ハッシュ化（ソルトが違うので結果も違う）
	hash2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Println("=== ソルトの効果 ===")
	fmt.Println("同じパスワードを再ハッシュ:", string(hash2))
	fmt.Println("2つのハッシュは違う？:", string(hash) != string(hash2))
	fmt.Println()

	// パスワード検証（ログイン時に使う）
	fmt.Println("=== パスワード検証 ===")

	// 正しいパスワード
	err = bcrypt.CompareHashAndPassword(hash, []byte("password123"))
	fmt.Println("正しいパスワードで検証 hash:", err == nil) // true
	err = bcrypt.CompareHashAndPassword(hash2, []byte("password123"))
	fmt.Println("正しいパスワードで検証 hash2:", err == nil) // true

	// 間違ったパスワード
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword"))
	fmt.Println("間違ったパスワードで検証:", err == nil) // false
	err = bcrypt.CompareHashAndPassword(hash2, []byte("wrongpassword"))
	fmt.Println("間違ったパスワードで検証 hash2:", err == nil) // false
}


func salt() {
	password := "password123"

	fmt.Println("=== ソルトなしの場合（危険） ===")
	fmt.Println()

	// ソルトなし: 同じパスワードは同じハッシュになる
	hash1 := sha256.Sum256([]byte(password))
	hash2 := sha256.Sum256([]byte(password))
	hash3 := sha256.Sum256([]byte(password))

	fmt.Println("ユーザーA:", fmt.Sprintf("%x", hash1)[:20], "...")
	fmt.Println("ユーザーB:", fmt.Sprintf("%x", hash2)[:20], "...")
	fmt.Println("ユーザーC:", fmt.Sprintf("%x", hash3)[:20], "...")
	fmt.Println()
	fmt.Println("→ 全員同じハッシュ！攻撃者に「同じパスワード」とバレる")

	fmt.Println()
	fmt.Println("=== ソルトありの場合（安全） ===")
	fmt.Println()

	// ソルトあり: 各ユーザーにランダムな文字列を追加
	saltA := "random_salt_A"
	saltB := "random_salt_B"
	saltC := "random_salt_C"

	hashA := sha256.Sum256([]byte(password + saltA))
	hashB := sha256.Sum256([]byte(password + saltB))
	hashC := sha256.Sum256([]byte(password + saltC))

	fmt.Println("ユーザーA (ソルト:", saltA, ")")
	fmt.Println("  ハッシュ:", fmt.Sprintf("%x", hashA)[:20], "...")
	fmt.Println()
	fmt.Println("ユーザーB (ソルト:", saltB, ")")
	fmt.Println("  ハッシュ:", fmt.Sprintf("%x", hashB)[:20], "...")
	fmt.Println()
	fmt.Println("ユーザーC (ソルト:", saltC, ")")
	fmt.Println("  ハッシュ:", fmt.Sprintf("%x", hashC)[:20], "...")
	fmt.Println()
	fmt.Println("→ 全員違うハッシュ！同じパスワードでもわからない")

	fmt.Println()
	fmt.Println("=== 検証の仕組み ===")
	fmt.Println()

	// DBに保存されているもの（例: ユーザーA）
	savedSalt := saltA
	savedHash := hashA

	// ログイン時: ユーザーが入力したパスワード
	inputPassword := "password123"

	// 検証: 保存されたソルト + 入力パスワード でハッシュ化して比較
	checkHash := sha256.Sum256([]byte(inputPassword + savedSalt))

	fmt.Println("保存されたハッシュ:", fmt.Sprintf("%x", savedHash)[:20], "...")
	fmt.Println("検証用ハッシュ:    ", fmt.Sprintf("%x", checkHash)[:20], "...")
	fmt.Println("一致:", savedHash == checkHash)
}