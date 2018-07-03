package accesslog

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ochipin/logger"
)

var environRegex = regexp.MustCompile(`%\{(.+?)\}[iec]`)

// Log 構造体は、ログ情報を取り扱う構造体
type Log struct {
	logger.Log
	Format   string // ログフォーマット
	Modename string // 起動モード名
	ServerIP string // サーバのIPアドレス
	hostname string // ホスト名
}

// Logger : ログ管理インタフェース
type Logger interface {
	Print(int, time.Time, *http.Request)
}

// MakeLog 関数はログ管理構造体を初期化する
func (l *Log) MakeLog(out *os.File) (Logger, error) {
	if err := l.Initializer(out); err != nil {
		return nil, err
	}
	// 全て1行で出力することを強制する
	l.Newline = true
	l.Trim = true
	l.Tabspace = true
	// ホスト名を取得する
	l.hostname = "-"
	if hostname, err := os.Hostname(); err == nil {
		l.hostname = hostname
	}

	return l, nil
}

// Print 関数はログを出力する
func (l *Log) Print(status int, start time.Time, r *http.Request) {
	info := l.message(status, start, r)
	l.Log.Print(info, "\n")
}

// 認証ユーザ名を取得する
func (l *Log) username(au string) string {
	if au == "" {
		return "-"
	}

	strs := strings.Split(au, " ")
	if len(strs) != 2 || strs[0] != "Basic" {
		return "-"
	}

	if buf, err := base64.StdEncoding.DecodeString(strs[1]); err == nil {
		if idx := strings.Index(string(buf), ":"); idx != -1 {
			return string(buf)[:idx]
		}
	}

	return "-"
}

// 出力するログ情報を選定する
func (l *Log) message(status int, start time.Time, r *http.Request) string {
	formats := strings.Split(l.Format, "%%")
	for i := 0; i < len(formats); i++ {
		formats[i] = l.getinfo(formats[i], status, start, r)
	}
	return strings.Join(formats, "%")
}

// 出力するログ情報を整理する
func (l *Log) getinfo(format string, status int, start time.Time, r *http.Request) string {
	// アクセス元のIPアドレスを取得する
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	// サーバのIPアドレスを取得する
	sa := l.ServerIP
	if sa == "" {
		sa = "-"
	}
	// Content-Length を取得する
	cl := fmt.Sprint(r.ContentLength)
	// モード名の取得
	mn := l.Modename
	// リクエストホスト名:ポートを取得する
	rh := r.Host
	// リクエストプロトコルを取得する
	rp := r.Proto
	// リクエストメソッドを取得する
	rm := r.Method
	// クエリパラメータを取得する
	qp := r.URL.RawQuery
	if qp != "" {
		qp = "?" + qp
	}
	// アクセス時刻を取得する
	at := start.Format("2006-01-02 15:04:05")
	// 認証ユーザ名を取得する
	au := l.username(r.Header.Get("Authorization"))
	// クエリパスを取得する
	up := r.URL.Path
	// アップロードファイル名を取得する
	var fn = "-"
	if reader, err := r.MultipartReader(); err == nil {
		if form, err := reader.ReadForm(32 << 20); err == nil {
			var names []string
			for fname, _ := range form.File {
				names = append(names, fname)
			}
			if len(names) == 0 {
				fn = strings.Join(names, ";")
			}
		}
	}
	// 正式なホスト名を取得する
	vh := l.hostname
	// 正式なポート番号を取得する
	var vp = "-"
	if idx := strings.Index(rh, ":"); idx != -1 {
		vp = rh[idx+1:]
	}
	// ユーザエージェント情報を取得する
	ua := r.UserAgent()
	// プロトコル名を取得する
	pn := "http"
	if r.TLS != nil {
		pn = "https"
	}
	// 接続ステータスを取得する
	st := fmt.Sprint(status)
	// リバースプロキシ経由の場合、リアルIPを取得する
	xf := r.Header.Get("X-Forwarded-For")
	if xf == "" {
		xf = "-"
	}
	// 参照元URLの取得
	rf := r.Referer()
	if rf == "" {
		rf = "-"
	}
	// 処理時間を取得する
	et := fmt.Sprint(time.Since(start).String())
	// 取得した値で、フォーマットを置き換える
	rep := strings.NewReplacer(
		"%ra", ra, // 訪問者(ユーザ)のIPアドレス
		"%sa", sa, // サーバのIPアドレス
		"%cl", cl, // 送信バイト数(Byte)
		"%et", et, // レスポンスタイム
		"%fn", fn, // アップロードファイル名
		"%rh", rh, // リモートホスト:ポート
		"%rp", rp, // リクエストプロトコル
		"%rm", rm, // リクエストメソッド
		"%vp", vp, // 正式なサーバが使用するポート番号
		"%qp", qp, // クエリパラメータ
		"%st", st, // 接続ステータス
		"%at", at, // アクセス時刻
		"%au", au, // 認証ユーザ名
		"%up", up, // URLパス
		"%vh", vh, // 正式なサーバ名
		"%ua", ua, // ユーザエージェント
		"%mn", mn, // アプリケーション起動モード名
		"%pn", pn, // プロトコル名
		"%xf", xf, // リバースプロキシ使用時のリアルIP
		"%rf", rf, // リファラ
	)
	info := rep.Replace(format)

	// リクエストヘッダ/環境変数/クッキーから情報の取得を行う
	info = environRegex.ReplaceAllStringFunc(info, func(str string) string {
		r := r
		length := len(str)
		c := str[length-1]
		var result string
		if c == 'i' {
			// リクエストヘッダから情報を取得
			name := str[2 : length-2]
			if name == "Authorization" {
				result = l.username(r.Header.Get("Authorization"))
			} else {
				result = r.Header.Get(name)
			}
		} else if c == 'e' {
			// 環境変数から情報を取得
			result = os.Getenv(str[2 : length-2])
		} else if c == 'c' {
			// クッキーから情報を取得
			if v, err := r.Cookie(str[2 : length-2]); err == nil {
				result = v.Value
			}
		}
		// 取得した情報が空の場合、 "-" を格納する
		if result == "" {
			result = "-"
		}
		return result
	})
	// ログに出力する情報を返却する
	return info
}
