package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned an empty hash")
	}

	if hash == password {
		t.Fatal("HashPassword() returned the plain-text password")
	}
}


func TestCheckPassword(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		if err := CheckPassword(password, hash); err != nil {
			t.Errorf("CheckPassword() returned error: %v", err)
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		if err := CheckPassword("wrong-password", hash); err == nil {
			t.Error("CheckPassword() expected error for incorrect password, got nil")
		}
	})

	t.Run("invalid hash", func(t *testing.T) {
		if err := CheckPassword(password, "invalid-hash"); err == nil {
			t.Error("CheckPassword() expected error for invalid hash, got nil")
		}
	})
}