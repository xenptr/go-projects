package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := int64(123)

	req = SetUserID(req, userID)

	gotUserID, ok := GetUserID(req)
	if !ok {
		t.Fatal("GetUserID() returned false, want true")
	}

	if gotUserID != userID {
		t.Errorf("GetUserID() = %d, want %d", gotUserID, userID)
	}
}

func TestGetUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	gotUserID, ok := GetUserID(req)

	if ok {
		t.Fatal("GetUserID() returned true, want false")
	}

	if gotUserID != 0 {
		t.Errorf("GetUserID() = %d, want 0", gotUserID)
	}
}
