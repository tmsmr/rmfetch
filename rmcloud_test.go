package rmfetch

import (
	"errors"
	"testing"
)

// TestNewApi requires RMAPI_DEVICE_CODE to be set
func TestNewApi(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Fatal(err)
	}
}

// TestGetDocs requires RMAPI_DEVICE_CODE to be set
func TestGetDocs(t *testing.T) {
	rmcloud, err := New()
	if err != nil {
		t.Fatal(err)
	}
	_ = rmcloud.Docs()
}

// TestFetchDoc requires RMAPI_DEVICE_CODE to be set and any available doc in the RM cloud
func TestFetchDoc(t *testing.T) {
	rmcloud, err := New()
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

// TestGenPDF requires RMAPI_DEVICE_CODE to be set, a running instance of https://github.com/tmsmr/rmrl-aas and any available doc in the RM cloud
func TestGenPDF(t *testing.T) {
	rmcloud, err := New()
	if err != nil {
		t.Fatal(err)
	}
	docs := rmcloud.Docs()
	if len(docs) == 0 {
		t.Fatal(errors.New("no docs available"))
	}
	_, err = rmcloud.GenPDF(docs[0], "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
}
