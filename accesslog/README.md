アクセスログ出力ライブラリ
===========================================================
accesslogライブラリは、loggerライブラリを継承したライブラリ。

```go
package main

import (
    "github.com/ochipin/logger/accesslog"
    "os"
    "net/http"
)

func main() {
    log := &accesslog.Log{}
    log.Format = "%ra - %au [%at] \"%rm %up %rp\" %st %cl \"%rf\" \"%ua\" %et"
    log.ServerIP = "192.168.1.3"
    log.Modename = "development"

    l, err := log.MakeLog(os.Stdout)
    if err != nil {
        panic(err)
    }

    ...
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ...
    // ログ出力
    l.Print(200, start, r)
    // 192.168.1.2 - - [YYYY-MM-DD HH:MI:SS] "GET / HTTP/1.1" 200 0 "-" "Mozilla/5.0 (...) like Gecko" 154.232ms
}
```

## logger.Logger インターフェース
Loggerインターフェースは、アクセスログ出力関数を所持するインターフェースである。
このインターフェースは、Log.MakeLog関数から生成される(後述)

引数は、以下の値を受け取る。

| 引数 | 説明 |
|:--   |:-- |
| status int      | HTTPステータス  |
| start time.Time | アクセス開始時刻 |
| r *http.Request | リクエスト情報 |

## logger.Log 構造体のパラメータ

logger ライブラリに加え、以下のパラメータを追加している。

| パラメータ | 説明 |
|:--|:--|
| Format   | ログフォーマット指定子                   |
| Modename | 起動したアプリケーションのモードを格納する |
| ServerIP | サーバのIPアドレスを登録する              |

ログフォーマット指定子を設定する`Format`は、下記指定子を利用可能。

| フォーマット指定子 | 説明 |
|:--  |:-- |
| %ra | 訪問者(ユーザ)のIPアドレス |
| %sa | 訪問者(ユーザ)のIPアドレス |
| %cl | 送信バイト数(Byte) |
| %et | レスポンスタイム |
| %fn | アップロードファイル名 |
| %rh | リモートホスト:ポート |
| %rp | リクエストプロトコル |
| %rm | リクエストメソッド |
| %vp | 正式なサーバが使用するポート番号 |
| %qp | クエリパラメータ |
| %st | 接続ステータス |
| %at | アクセス時刻 |
| %au | 認証ユーザ名 |
| %up | URLパス |
| %vh | 正式なサーバ名 |
| %ua | ユーザエージェント |
| %mn | アプリケーション起動モード名 |
| %pn | プロトコル名 |
| %xf | リバースプロキシ使用時のリアルIP |
| %rf | 参照元URL |
| %{envname}e | 環境変数の値をログに出力する |
| %{header}i | リクエストヘッダの中身を出力する |
| %{cookie}c | クッキーの情報を出力する |

## logger.Log.MakeLog()
上で述べた、Loggerインターフェースを生成する関数。