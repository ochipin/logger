package logger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// 通常のログ保存テストを実施
func TestLogger(t *testing.T) {
	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:     "test/logger.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   timing,
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	l.Print("    Hello			World   \n	Hello World")
	l.Printf("%s %d", "Hello World", 200)

	log.Keeping()
	time.Sleep(2 * time.Second)
}

// ログ上書き保存のテスト
func TestLoggerOverwrite(t *testing.T) {
	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:      "test/logger.log",
		Lotate:    "test/%Y%m/logger.%Y%m%d.log",
		Timing:    timing,
		Tabspace:  true,
		Newline:   true,
		Trim:      true,
		Perm:      0,
		Overwrite: true,
	}
	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	l.Print("    Hello			World   \n	Hello World")
	l.Printf("%s %d", "Hello World", 200)

	log.Keeping()
	time.Sleep(2 * time.Second)
}

// ログファイルへ保存しない
func TestLoggerNoWrite(t *testing.T) {
	log := Log{
		Path:      "",
		Lotate:    "test/%Y%m/logger.%Y%m%d.log",
		Timing:    "00:00",
		Tabspace:  true,
		Newline:   true,
		Trim:      true,
		Perm:      0,
		Overwrite: true,
	}
	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	l.Print("    Hello			World   \n	Hello World")
	l.Printf("%s %d", "Hello World", 200)

	// log.Keeping は何もせず復帰する
	log.Keeping()
}

// Timing : 時刻間違い
func TestLoggerTiming(t *testing.T) {
	log := Log{
		Path:     "test/logger.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   "32:90",
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err == nil {
		t.Fatal(err)
	}
	log.Timing = "01:O9"
	_, err = log.MakeLog(os.Stdout)
	if err == nil {
		t.Fatal(err)
	}
}

// ローテーション後のファイル名が不正
func TestLoggerLotate(t *testing.T) {
	log := Log{
		Path:     "test/logger.log",
		Lotate:   "test/%Y%m/",
		Timing:   "18:00",
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err == nil {
		t.Fatal(err)
	}

	log.Lotate = "test/%Y%m/logger.%Y%m%d.log"
	log.Keeping()
	time.Sleep(2 * time.Second)
}

// アクセスが1回もなく、ログファイルが生成されていない状態からのローテーション
func TestLoggerLotation(t *testing.T) {
	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:     "test/loggertemp.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   timing,
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	log.Keeping()
	time.Sleep(2 * time.Second)
}

// ログローテーション最中にローテーション名が不正になった場合
func TestLoggerInvalidLotate(t *testing.T) {
	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:     "test/loggertemp.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   timing,
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	log.Keeping()
	log.Lotate = "test/%Y%m/"
	time.Sleep(2 * time.Second)
}

// ディレクトリ作成エラー
func TestLoggerDirError(t *testing.T) {
	os.MkdirAll("test/dir", os.FileMode(0400))
	log := Log{
		Path:     "test/dir/sample/loggertemp.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   "00:00",
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	l.Print("Hello World")
	l.Printf("%s", "Hello World")
}

// 書き込みエラーテスト
func TestLoggerWriteError(t *testing.T) {
	log := Log{
		Path:     "test/loggertemp.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   "00:00",
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	os.Chmod("test/loggertemp.log", os.FileMode(0400))
	l.Print("Hello World")
	l.Printf("%s", "Hello World")
}

// ログローテーション時のファイル書き込みエラー
func TestLoggerWriteLotateError(t *testing.T) {
	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:     "test/loggertemp.log",
		Lotate:   "test/%Y%m/logger.%Y%m%d.log",
		Timing:   timing,
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	os.Chmod(log.getLotateName(now), os.FileMode(0100))
	log.Keeping()
	time.Sleep(2 * time.Second)
}

// ディレクトリ作成失敗時の挙動テスト
func TestLoggerMkdirLotateError(t *testing.T) {
	os.MkdirAll("test/dir2", os.FileMode(0400))

	now := time.Now()
	timing := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	log := Log{
		Path:     "test/loggertemp.log",
		Lotate:   "test/dir2/sample/logger.log",
		Timing:   timing,
		Tabspace: true,
		Newline:  true,
		Trim:     true,
		Perm:     0,
	}
	_, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	os.Chmod(log.getLotateName(now), os.FileMode(0100))
	log.Keeping()
	time.Sleep(2 * time.Second)
}
