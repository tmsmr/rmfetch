package rmfetch

import (
	"errors"
	"testing"
)

// requires a valid rmapi.conf or a valid OneTimeCode
func TestNewApi(t *testing.T) {
	conf := Config{OneTimeCode: nil}
	_, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
}

// requires a valid rmapi.conf or a valid OneTimeCode
func TestGetDocs(t *testing.T) {
	conf := Config{OneTimeCode: nil}
	rmcloud, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	_ = rmcloud.Docs()
}

// requires a valid rmapi.conf or a valid OneTimeCode + any available doc in the RM cloud
func TestFetchDoc(t *testing.T) {
	conf := Config{OneTimeCode: nil}
	rmcloud, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	docs := rmcloud.Docs()
	if len(docs) == 0 {
		t.Fatal(errors.New("no docs available"))
	}
	_, err = rmcloud.Fetch(docs[0])
	if err != nil {
		t.Fatal(err)
	}
}
