package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>ブラウザストレージ デモ</title>
    <style>
        body { font-family: sans-serif; padding: 20px; max-width: 900px; margin: 0 auto; }
        .container { display: flex; gap: 20px; flex-wrap: wrap; }
        .storage-box { border: 2px solid #333; padding: 15px; border-radius: 8px; flex: 1; min-width: 250px; }
        .cookie { border-color: #e74c3c; }
        .local { border-color: #3498db; }
        .session { border-color: #2ecc71; }
        h2 { margin-top: 0; }
        .cookie h2 { color: #e74c3c; }
        .local h2 { color: #3498db; }
        .session h2 { color: #2ecc71; }
        input { padding: 8px; margin: 5px 0; width: 100%; box-sizing: border-box; }
        button { padding: 8px 16px; margin: 5px 5px 5px 0; cursor: pointer; }
        .value { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px; word-break: break-all; min-height: 20px; }
        .log { background: #1a1a2e; color: #0f0; padding: 15px; border-radius: 8px; margin-top: 20px; font-family: monospace; max-height: 200px; overflow-y: auto; }
        .log-entry { margin: 5px 0; }
        .info { background: #fff3cd; padding: 15px; border-radius: 8px; margin: 20px 0; }
        .experiment { background: #e8f4f8; padding: 15px; border-radius: 8px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Phase 3-3: ブラウザストレージ デモ</h1>

    <div class="info">
        <strong>実験方法:</strong><br>
        1. このページを<strong>2つのタブ</strong>で開く（Cmd+T または Ctrl+T）<br>
        2. 片方のタブで値を保存して、もう片方で「読み取り」ボタンを押す<br>
        3. どのストレージがタブ間で共有されるか確認する
    </div>

    <div class="container">
        <div class="storage-box cookie">
            <h2>Cookie</h2>
            <p>サーバーに自動送信される</p>
            <input type="text" id="cookieInput" placeholder="保存する値">
            <div>
                <button onclick="setCookie()">保存</button>
                <button onclick="getCookie()">読み取り</button>
                <button onclick="deleteCookie()">削除</button>
            </div>
            <div class="value" id="cookieValue">-</div>
        </div>

        <div class="storage-box local">
            <h2>localStorage</h2>
            <p>永続保存（削除するまで残る）</p>
            <input type="text" id="localInput" placeholder="保存する値">
            <div>
                <button onclick="setLocal()">保存</button>
                <button onclick="getLocal()">読み取り</button>
                <button onclick="deleteLocal()">削除</button>
            </div>
            <div class="value" id="localValue">-</div>
        </div>

        <div class="storage-box session">
            <h2>sessionStorage</h2>
            <p>タブを閉じると消える</p>
            <input type="text" id="sessionInput" placeholder="保存する値">
            <div>
                <button onclick="setSession()">保存</button>
                <button onclick="getSession()">読み取り</button>
                <button onclick="deleteSession()">削除</button>
            </div>
            <div class="value" id="sessionValue">-</div>
        </div>
    </div>

    <div class="experiment">
        <h3>実験: タブ間同期（storageイベント）</h3>
        <p>他のタブで localStorage を変更すると、このログに表示されます</p>
        <button onclick="clearLog()">ログクリア</button>
    </div>

    <div class="log" id="log">
        <div class="log-entry">[起動] ストレージデモを開始しました</div>
    </div>

    <script>
        // ログ出力
        function log(message) {
            const logDiv = document.getElementById('log');
            const time = new Date().toLocaleTimeString();
            logDiv.innerHTML += '<div class="log-entry">[' + time + '] ' + message + '</div>';
            logDiv.scrollTop = logDiv.scrollHeight;
        }

        function clearLog() {
            document.getElementById('log').innerHTML = '<div class="log-entry">[クリア] ログをクリアしました</div>';
        }

        // Cookie操作
        function setCookie() {
            const value = document.getElementById('cookieInput').value;
            document.cookie = 'demo=' + encodeURIComponent(value) + '; path=/; max-age=3600';
            log('Cookie保存: ' + value);
            getCookie();
        }

        function getCookie() {
            const cookies = document.cookie.split(';');
            for (let c of cookies) {
                const [key, val] = c.trim().split('=');
                if (key === 'demo') {
                    document.getElementById('cookieValue').textContent = decodeURIComponent(val);
                    return;
                }
            }
            document.getElementById('cookieValue').textContent = '(未設定)';
        }

        function deleteCookie() {
            document.cookie = 'demo=; path=/; max-age=0';
            log('Cookie削除');
            getCookie();
        }

        // localStorage操作
        function setLocal() {
            const value = document.getElementById('localInput').value;
            localStorage.setItem('demo', value);
            log('localStorage保存: ' + value);
            getLocal();
        }

        function getLocal() {
            const value = localStorage.getItem('demo');
            document.getElementById('localValue').textContent = value || '(未設定)';
        }

        function deleteLocal() {
            localStorage.removeItem('demo');
            log('localStorage削除');
            getLocal();
        }

        // sessionStorage操作
        function setSession() {
            const value = document.getElementById('sessionInput').value;
            sessionStorage.setItem('demo', value);
            log('sessionStorage保存: ' + value);
            getSession();
        }

        function getSession() {
            const value = sessionStorage.getItem('demo');
            document.getElementById('sessionValue').textContent = value || '(未設定)';
        }

        function deleteSession() {
            sessionStorage.removeItem('demo');
            log('sessionStorage削除');
            getSession();
        }

        // 他タブでの変更を検知（storageイベント）
        window.addEventListener('storage', function(e) {
            log('<strong style="color: yellow;">他のタブで変更検知!</strong> key=' + e.key + ', 新しい値=' + e.newValue);
            // 画面を更新
            getLocal();
        });

        // 初期読み込み
        getCookie();
        getLocal();
        getSession();
        log('タブID: ' + Math.random().toString(36).substr(2, 5) + ' (このタブの識別用)');
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	})

	fmt.Println("ブラウザストレージ デモサーバー起動")
	fmt.Println("http://localhost:3000 を2つのタブで開いてください")
	http.ListenAndServe(":3000", nil)
}
