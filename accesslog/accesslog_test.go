package accesslog

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAccessLog(t *testing.T) {
	log := &Log{}
	log.Format = "%ra %qp %sa"
	log.ServerIP = ""
	log.Modename = "development"

	l, err := log.MakeLog(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		l.Print(200, now, r)
	}))
	defer testServer.Close()

	_, err = http.Get(testServer.URL + "?query=200")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}
}
