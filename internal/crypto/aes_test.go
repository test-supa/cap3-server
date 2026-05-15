package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key1 := DeriveKey("device123", "secret")
	key2 := DeriveKey("device123", "secret")
	key3 := DeriveKey("device456", "secret")

	if len(key1) != KeySize {
		t.Errorf("expected key size %d, got %d", KeySize, len(key1))
	}

	if !bytes.Equal(key1, key2) {
		t.Error("same inputs should produce same key")
	}

	if bytes.Equal(key1, key3) {
		t.Error("different device IDs should produce different keys")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey("test-device", "test-secret")

	plaintext := []byte("hello world this is a secret message")
	encoded, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(encoded, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("plaintext mismatch: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptEmpty(t *testing.T) {
	key := DeriveKey("test", "secret")

	encoded, err := Encrypt([]byte{}, key)
	if err != nil {
		t.Fatalf("encrypt empty failed: %v", err)
	}

	decrypted, err := Decrypt(encoded, key)
	if err != nil {
		t.Fatalf("decrypt empty failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Error("decrypted empty should be empty")
	}
}

func TestEncryptLarge(t *testing.T) {
	key := DeriveKey("large-test", "secret")

	plaintext := bytes.Repeat([]byte("A"), 1024*1024) // 1MB
	encoded, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt large failed: %v", err)
	}

	decrypted, err := Decrypt(encoded, key)
	if err != nil {
		t.Fatalf("decrypt large failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("large plaintext mismatch")
	}
}

func TestDecryptInvalid(t *testing.T) {
	key := DeriveKey("test", "secret")

	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"too short", "too"},
		{"garbage", "dGhpcyBpcyBub3QgdmFsaWQgZW5jcnlwdGVkIGRhdGE="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.data, key)
			if err == nil {
				t.Error("expected error for invalid data")
			}
		})
	}
}

func TestWrongKey(t *testing.T) {
	key1 := DeriveKey("device1", "secret")
	key2 := DeriveKey("device2", "secret")

	plaintext := []byte("sensitive data")
	encoded, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	_, err = Decrypt(encoded, key2)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestGenerateMasterSecret(t *testing.T) {
	secret, err := GenerateMasterSecret()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if len(secret) == 0 {
		t.Error("master secret should not be empty")
	}

	// Should be base64 encoded (at least 44 chars for 32 bytes)
	if len(secret) < 44 {
		t.Errorf("master secret too short: %d", len(secret))
	}
}
