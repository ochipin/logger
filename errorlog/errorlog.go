package errorlog

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/ochipin/logger"
)

var matchSource = regexp.MustCompile(`%[fml]`)

// Log 構造体は、ログ情報を取り扱う構造体
type Log struct {
	logger.Log
	Format string // ログフォーマット
	Level  int    // ログレベル
	Depth  int    // 実行された関数、行番号等を取得する際に使用する階層
}

// Logger : ログ管理インタフェース
type Logger interface {
	Emerg(...interface{})
	Emergf(string, ...interface{})
	Alert(...interface{})
	Alertf(string, ...interface{})
	Crit(...interface{})
	Critf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Notice(...interface{})
	Noticef(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
}

// MakeLog : ログ管理構造体を初期化する
func (l *Log) MakeLog(out *os.File) (Logger, error) {
	if !(l.Level >= 0 && l.Level <= 7) {
		return nil, fmt.Errorf("please log level set 0-7")
	}
	// ソースコードの情報を取得するDepth値を検証
	if l.Depth == 0 {
		l.Depth = 1
	}
	if err := l.Initializer(out); err != nil {
		return nil, err
	}
	return l, nil
}

// 指定されたログレベル名を取得する
func (l *Log) logLevel(level int) (string, error) {
	if l.Level >= level {
		switch level {
		case 0:
			return "emerg", nil
		case 1:
			return "alert", nil
		case 2:
			return "crit", nil
		case 3:
			return "error", nil
		case 4:
			return "warn", nil
		case 5:
			return "notice", nil
		case 6:
			return "info", nil
		case 7:
			return "debug", nil
		}
	}
	return "", fmt.Errorf("\"%d\" log level is not found", level)
}

// 実行中のソースファイル名、関数名、行番号を取得する
func (l *Log) source() (string, string, int) {
	// runtime.Callerで実行中の関数名やソースファイル名を取得し返却する
	if pc, filename, linenum, ok := runtime.Caller(l.Depth); ok {
		// 関数名を取得
		funcname := runtime.FuncForPC(pc).Name()
		// ソースファイル名、関数名をショートネーム化する
		funcname = funcname[strings.LastIndex(funcname, ".")+1:]
		filename = filename[strings.LastIndex(filename, "/")+1:]
		return filename, funcname, linenum
	}
	// 実行中の情報を取得できなかった場合は、適当な値を返却する
	return "???", "???", 0
}

// 出力するメッセージをフォーマットに沿った形式に変換する
func (l *Log) message(level int) (string, error) {
	// 返却する値をフォーマット文字列で初期化 ex) %D %T %f(%m:%l) %M
	var result = l.Format

	// ログレベルを取得する
	levelname, err := l.logLevel(level)
	if err != nil {
		return "", err
	}

	// %f, %l, %m 等のフォーマットが存在する場合、関数名、ファイル名、行番号等を取得し埋め込む
	if matchSource.MatchString(l.Format) {
		filename, funcname, linenum := l.source()
		rep := strings.NewReplacer(
			"%f", filename,
			"%m", funcname,
			"%l", fmt.Sprint(linenum))
		// ex) %D %T main.go(main:11) %M
		result = rep.Replace(l.Format)
	}

	// %D, %T を日時に置き換え、%Mを出力するメッセージに置き換える
	now := time.Now()
	rep := strings.NewReplacer(
		"%D", now.Format("2006-01-02"),
		"%T", now.Format("15:04:05"),
		"%L", levelname)
	// 2018-03-21 21:22:02 main.go(main:11) error: a.out not found...
	return rep.Replace(result), nil
}

// Emerg : ログレベル0。呼び出されるとリターンコード127でプログラムを強制終了する
func (l *Log) Emerg(v ...interface{}) {
	mes, _ := l.message(0)
	l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	os.Exit(127)
}

// Emergf : ログレベル0。呼び出されるとリターンコード127でプログラムを強制終了する
func (l *Log) Emergf(format string, v ...interface{}) {
	mes, _ := l.message(0)
	l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	os.Exit(127)
}

// Alert : ログレベル1。緊急でかつ重大なエラーに使用する
func (l *Log) Alert(v ...interface{}) {
	if mes, err := l.message(1); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Alertf : ログレベル1。緊急でかつ重大なエラーに使用する
func (l *Log) Alertf(format string, v ...interface{}) {
	if mes, err := l.message(1); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Crit : ログレベル2。重大エラーに使用する
func (l *Log) Crit(v ...interface{}) {
	if mes, err := l.message(2); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Critf : ログレベル2。重大エラーに使用する
func (l *Log) Critf(format string, v ...interface{}) {
	if mes, err := l.message(2); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Error : ログレベル3。エラー発生時に使用する
func (l *Log) Error(v ...interface{}) {
	if mes, err := l.message(3); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Errorf : ログレベル3。エラー発生時に使用する
func (l *Log) Errorf(format string, v ...interface{}) {
	if mes, err := l.message(3); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Warn : ログレベル4。警告メッセージ
func (l *Log) Warn(v ...interface{}) {
	if mes, err := l.message(4); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Warnf : ログレベル4。警告メッセージ
func (l *Log) Warnf(format string, v ...interface{}) {
	if mes, err := l.message(4); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Notice : ログレベル5。通知メッセージ
func (l *Log) Notice(v ...interface{}) {
	if mes, err := l.message(5); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Noticef : ログレベル5。通知メッセージ
func (l *Log) Noticef(format string, v ...interface{}) {
	if mes, err := l.message(5); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Info : ログレベル6。通常メッセージ
func (l *Log) Info(v ...interface{}) {
	if mes, err := l.message(6); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Infof : ログレベル6。通常メッセージ
func (l *Log) Infof(format string, v ...interface{}) {
	if mes, err := l.message(6); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}

// Debug : ログレベル7。デバッグメッセージ
func (l *Log) Debug(v ...interface{}) {
	if mes, err := l.message(7); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprint(v...), -1))
	}
}

// Debugf : ログレベル7。デバッグメッセージ
func (l *Log) Debugf(format string, v ...interface{}) {
	if mes, err := l.message(7); err == nil {
		l.Print(strings.Replace(mes, "%M", fmt.Sprintf(format, v...), -1))
	}
}
