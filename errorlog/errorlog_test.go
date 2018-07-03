package errorlog

import (
	"os"
	"testing"
)

func TestErrorLog(t *testing.T) {
	log := Log{}
	log.Format = "%D %T %f(%m:%l) %L: %M"
	log.Depth = 3
	log.Level = 7

	l, err := log.MakeLog(os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	l.Alert("Hello World")
	l.Alertf("%s %d", "Hello World", 200)
	l.Crit("Hello World")
	l.Critf("%s %d", "Hello World", 200)
	l.Error("Hello World")
	l.Errorf("%s %d", "Hello World", 200)
	l.Warn("Hello World")
	l.Warnf("%s %d", "Hello World", 200)
	l.Notice("Hello World")
	l.Noticef("%s %d", "Hello World", 200)
	l.Info("Hello World")
	l.Infof("%s %d", "Hello World", 200)
	l.Debug("Hello World")
	l.Debugf("%s %d", "Hello World", 200)
}
