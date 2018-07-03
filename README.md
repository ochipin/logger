ロギングライブラリ
===========================================================

```go
package main

import "github.com/ochipin/logger"
import "os"

func main() {
    log := logger.Log{
        Path:    "log/access.log",
        Lotate:  "log/%Y%m/access-%Y%m%d.log",
        Timing:  "00:00",
        Perm:    0644,
    }

    l, err := log.MakeLog(os.Stdout)
    if err != nil {
        panic(err)
    }

    // ログ情報の保存と出力
    l.Print("Hello World")
    l.Printf("%s %d", "Hello World", 200)

    // 保存されたログファイルをローテーションする
    log.Keeping()
}
```

## logger.Logger インターフェース
Loggerインターフェースは、`Print`, `Printf`の2つのログ出力関数を所持するインターフェースである。
このインターフェースは、Log.MakeLog関数から生成される(後述)

## logger.Log 構造体のパラメータ

| パラメータ | 説明 |
|:--|:--|
| Path      | ログファイルの保存場所                           |
| Lotate    | ログローテーションがされた際に、移動する場所        |
| Timing    | ログローテーションする時刻。時:分で指定する         |
| Newline   | デフォルト false。true の場合、改行コードを削除する |
| Tabspace  | デフォルト false。true の場合、タブを空白に置き換える |
| Trim      | デフォルト false。true の場合、Trimを行う |
| Overwrite | デフォルト false。true の場合、ログローテーション時に、すでにあるファイルに対して、上書きを実施。falseの場合は、追加書き込みを実施する。 |
| Perm      | 保存するログのパーミッション |

上記パラメータで、`Lotate`パラメータに関しては、以下のフォーマット指定子を使用することができる。

| フォーマット指定子 | 説明 |
|:-- |:-- |
| %Y | YYYY |
| %m | MM |
| %d | DD |
| %w | 週名 |
| %h | HH |

ログローテーションは、`Timing`で指定された時刻に実施される。

`log/%Y%m/access-%Y%m%d.log` と指定した場合は、下記のようにログローテションされる。
```
2017/01/01 ---> log/201701/access-20170101.log
2017/01/02 ---> log/201701/access-20170102.log
2017/01/03 ---> log/201701/access-20170103.log
```
また、1週間単位でのログローテーションを実施する場合は、`log/access-%w.log`とする。

```
2017/01/01(Sun) ---> log/access-Sun.log
2017/01/02(Man) ---> log/access-Man.log
2017/01/03(Tue) ---> log/access-Tue.log
2017/01/04(Wed) ---> log/access-Wed.log
2017/01/05(Thu) ---> log/access-Thu.log
2017/01/06(Fri) ---> log/access-Fri.log
2017/01/07(Sat) ---> log/access-Sat.log
2017/01/08(Sun) ---> log/access-Sun.log <--- 上書き or 追加書き込み
                                             Overwrite のフラグにより挙動が異なる
```

## logger.Log.MakeLog()
Loggerインターフェースを生成する関数。

## logger.Log.Keeping()
保存されたログファイルのローテーションを実施する関数。