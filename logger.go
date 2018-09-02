package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Log 構造体は、ログ情報を取り扱う構造体
type Log struct {
	Path      string     // ログ保存パス
	Lotate    string     // ログローテーションファイル名
	Timing    string     // ログローテーションタイミング
	Newline   bool       // ログ保存時に、改行を含めるか否か
	Tabspace  bool       // ログ保存時に、タブを空白に置き換えるか
	Trim      bool       // ログ保存時に、Trimする
	Perm      int        // ログファイル作成時のパーミッション
	Overwrite bool       // ログローテーション時に、既にあるファイルに対して上書きする
	mu        sync.Mutex // 同時書き込み制御を行うMutex
	out       *os.File   // 標準出力/標準エラー出力先
	lotate    bool       // ログローテーションするか否か
	hour      int        // ログローテーションする時刻
	minute    int        // ログローテーションする時刻
}

// Logger : ログ管理インタフェース
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Write([]byte) (int, error)
}

// Initializer : ログ管理構造体にセットされたパラメータが適切かチェックし、パラメータを初期化する
func (l *Log) Initializer(out *os.File) error {
	// ログローテーションが有効か否かをチェックする
	if !(l.Lotate == "" || l.Path == "" || l.Timing == "") {
		l.lotate = true
		// 時:分指定が正しいかチェックする
		if !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(l.Timing) {
			return fmt.Errorf("logger: log lotation time is invalid")
		}
		// 時:分指定が正しい場合は、時, 分を数字に変換する
		times := strings.Split(l.Timing, ":")
		hour, _ := strconv.Atoi(times[0])
		minute, _ := strconv.Atoi(times[1])
		// 時(0-23), 分(0-59) の範囲の時刻であれば、正とみなし、その逆は負とみなす
		if hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59 {
			l.hour = hour
			l.minute = minute
		} else {
			return fmt.Errorf("logger: log lotation time is invalid")
		}
		// ローテーション後のパス命名が正しいかチェックする
		_, filename := filepath.Split(l.Lotate)
		if filename == "" {
			return fmt.Errorf("logger: log lotate filename is invalid")
		}
	}
	// パーミッションを検証する
	if l.Perm == 0 {
		l.Perm = 0644
	}

	return nil
}

// MakeLog : ログ管理構造体を初期化する
func (l *Log) MakeLog(out *os.File) (Logger, error) {
	if err := l.Initializer(out); err != nil {
		return nil, err
	}
	l.out = out
	return l, nil
}

// Print : ログを出力する
func (l *Log) Print(v ...interface{}) {
	if err := l.output(fmt.Sprint(v...)); err != nil {
		l.alert("logger: " + err.Error())
	}
}

// Printf : ログを出力する
func (l *Log) Printf(format string, v ...interface{}) {
	if err := l.output(fmt.Sprintf(format, v...)); err != nil {
		l.alert("logger: " + err.Error())
	}
}

// Write : ログを出力する
func (l *Log) Write(b []byte) (int, error) {
	var out []byte = b
	if l := len(b); l != 0 && b[l-1] == '\n' {
		out = b[:l-1]
	}
	if err := l.output(string(out)); err != nil {
		l.alert("logger: " + err.Error())
		return 0, err
	}
	return len(b), nil
}

// シスログへ出力する
func (l *Log) alert(message string) {
	// シスログに出力する設定を実施
	logger, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_USER, "go-logger")
	if err != nil {
		return
	}
	// シスログへ出力する
	log.SetOutput(logger)
	log.Println(message)
}

// タブを半角空白へ置き換える
func (l *Log) tabToBlank(str string) string {
	return strings.Replace(str, "\t", "    ", -1)
}

// 改行を取り除く
func (l *Log) lineDelete(str string) string {
	rep := strings.NewReplacer("\r\n", "", "\n", "", "\r", "")
	return rep.Replace(str)
}

// 先頭の空白を取り除く
func (l *Log) trim(str string) string {
	return strings.Trim(str, " ")
}

// ログにメッセージを出力する
func (l *Log) output(s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// タブ ---> 空白置き換え
	if l.Tabspace {
		s = l.tabToBlank(s)
	}
	// 改行コードの削除を行う
	if l.Newline {
		s = l.lineDelete(s)
	}
	// Trimを行う
	if l.Trim {
		s = l.trim(s)
	}
	// 標準出力/標準エラー出力のどちらかが設定されている場合、出力する
	if l.out != nil {
		fmt.Fprint(l.out, s+"\n")
	}
	// ログファイル保存パスが未設定の場合、ファイルにはログ情報を保存しない
	if l.Path == "" {
		return nil
	}

	// ログ情報をファイルへ書き込む
	return l.savefile(s)
}

// ログ情報をファイルへ書き込む
func (l *Log) savefile(s string) error {
	// ログ保存先のパスから、ディレクトリ名のみ抜き出し、ディレクトリを作成する
	dir, _ := filepath.Split(l.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// ファイルオープンをする
	fp, err := os.OpenFile(l.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(l.Perm))
	if err != nil {
		return err
	}
	defer fp.Close()
	// ログファイルへ書き込む
	fmt.Fprint(fp, s+"\n")
	return nil
}

// Keeping : ログファイルをローテーションする
func (l *Log) Keeping() {
	// ログローテーションを実施いない場合は何もせず関数を抜ける
	if l.lotate == false {
		return
	}

	go func() {
		tick := time.NewTicker(time.Duration(1) * time.Second)
		ok := false
		for {
			select {
			// 1秒置きにログローテーションを実施する
			case <-tick.C:
				now := time.Now()
				if l.hour == now.Hour() && l.minute == now.Minute() {
					// 指定時刻になったら、1度だけログローテーションを実施する
					if ok == false {
						l.logReplace(now)
						ok = true
					}
				} else {
					// 指定時刻以外の場合、処理フラグをfalseにする
					ok = false
				}
			}
		}
	}()
}

// ログローテーションするファイル名を返却する
func (l *Log) getLotateName(now time.Time) string {
	var week = now.Weekday().String()[:3]
	rep := strings.NewReplacer(
		"%Y", fmt.Sprint(now.Year()),
		"%m", fmt.Sprintf("%02d", int(now.Month())),
		"%d", fmt.Sprintf("%02d", now.Day()),
		"%H", fmt.Sprintf("%02d", now.Hour()),
		"%w", week)
	return rep.Replace(l.Lotate)
}

// ログファイルを置き換える
func (l *Log) logReplace(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// ログローテーションするファイル名を変数へ格納
	// ex) log/%Y%m/app-%Y%m%d.log ---> log/201803/app-20180322.log
	lotatepath := l.getLotateName(now)
	// ex) log/201803/, app-20180322.log
	dirname, filename := filepath.Split(lotatepath)
	// ログローテーションするファイル名が不正の場合、関数を抜ける
	if filename == "" {
		l.alert("logger: log lotate filename is invalid")
		return
	}
	// ディレクトリが存在しない場合、作成する
	if dirname != "" {
		if err := os.MkdirAll(dirname, 0755); err != nil {
			l.alert("logger: " + err.Error())
			return
		}
	}

	// 1. ローテーションするファイルをオープン
	var fp *os.File
	var err error
	if l.Overwrite {
		// 上書きの場合は、ローテーションするファイルを作り直し
		os.Remove(lotatepath)
		fp, err = os.OpenFile(lotatepath, os.O_WRONLY|os.O_CREATE, os.FileMode(l.Perm))
	} else {
		// 追加書き込みの場合は、既存のローテーションファイルを読み込み
		fp, err = os.OpenFile(lotatepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(l.Perm))
	}
	if err != nil {
		l.alert("logger: " + err.Error())
		return
	}
	defer fp.Close()
	// 2. ログファイルを読み込む
	buf, err := ioutil.ReadFile(l.Path)
	if err != nil {
		buf = []byte{}
	}
	// 3. ログファイルの内容をローテーション後のログファイルへ書き込む
	fmt.Fprint(fp, string(buf))
	// 4. ログファイルの中身を0バイトにする
	ioutil.WriteFile(l.Path, []byte(""), os.FileMode(l.Perm))
}
