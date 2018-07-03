エラーログ専用ロギングライブラリ
===========================================================
errorlogライブラリは、loggerライブラリを継承したライブラリ。

```go
package main

import "github.com/ochipin/logger/errorlog"
import "os"

func main() {
    log := &errorlog.Log{}
    log.Format = "%D %T %f(%m:%l) %L: %M"
    log.Depth = 3
    log.Level = 7

    l, err := log.MakeLog(os.Stderr)
    if err != nil {
        panic(err)
    }

    // ログ情報の出力
    l.Debug("Hello World")
    l.Debugf("%s %d", "Hello World", 200)
    // YYYY-MM-DD HH:MI:SS main.go(main:24) debug: Hello World
    // YYYY-MM-DD HH:MI:SS main.go(main:25) debug: Hello World 200 
}
```

## logger.Logger インターフェース
Loggerインターフェースは、下記エラーログ出力関数を所持するインターフェースである。
このインターフェースは、Log.MakeLog関数から生成される(※1後述)

| メソッド  | 説明 |
|:-- |:-- |
| Debug[f] | ログレベル7。デバッグメッセージ |
| Info[f] | ログレベル6。通常メッセージ |
| Notice[f] | ログレベル5。通知メッセージ |
| Warn[f] | ログレベル4。警告メッセージ |
| Error[f] | ログレベル3。エラーメッセージ |
| Crit[f] | ログレベル2。重大エラーメッセージ |
| Alert[f] | ログレベル1。緊急でかつ重大なエラーメッセージ |
| Emerg[f] | ログレベル0。呼び出されるとリターンコード127でプログラムを強制終了する |

## logger.Log 構造体のパラメータ

logger ライブラリに加え、以下のパラメータを追加している。

| パラメータ | 説明 |
|:--|:--|
| Format | ログフォーマット指定子                           |
| Level  | ログレベル        |
| Depth  | ソースコード情報を取得する階層 |

ログフォーマット指定子を設定する`Format`は、下記指定子を利用可能。

| フォーマット指定子 | 説明 |
|:-- |:-- |
| %D | エラーログ出力関数がコールされた日付 |
| %T | エラーログ出力関数がコールされた時刻 |
| %L | ログレベル(emerg/alert/crit/error/warn/notice/info/debug) |
| %f | エラーログ出力関数がコールされた場所(ソースコードファイル名)  |
| %m | エラーログ出力関数がコールされた場所(関数名)  |
| %l | エラーログ出力関数がコールされた場所(行番号) |
| %M | エラーログ出力関数に渡したメッセージ内容 |

## logger.Log.MakeLog()
上で述べた、Loggerインターフェースを生成する関数。