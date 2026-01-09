package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptor(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		key := make([]byte, 32)
		encryptor, err := NewEncryptor(string(key))
		require.NoError(t, err)
		assert.NotNil(t, encryptor)
	})

	t.Run("invalid key length", func(t *testing.T) {
		key := "short-key"
		encryptor, err := NewEncryptor(key)
		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.Contains(t, err.Error(), "must be exactly 32 bytes")
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	t.Run("encrypt and decrypt string", func(t *testing.T) {
		plaintext := "sensitive-tax-id-12345"

		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("encrypt empty string", func(t *testing.T) {
		encrypted, err := encryptor.Encrypt("")
		require.NoError(t, err)
		assert.Empty(t, encrypted)
	})

	t.Run("decrypt empty string", func(t *testing.T) {
		decrypted, err := encryptor.Decrypt("")
		require.NoError(t, err)
		assert.Empty(t, decrypted)
	})

	t.Run("different plaintexts produce different ciphertexts", func(t *testing.T) {
		encrypted1, err := encryptor.Encrypt("plaintext1")
		require.NoError(t, err)

		encrypted2, err := encryptor.Encrypt("plaintext2")
		require.NoError(t, err)

		assert.NotEqual(t, encrypted1, encrypted2)
	})

	t.Run("same plaintext produces different ciphertexts (due to random nonce)", func(t *testing.T) {
		encrypted1, err := encryptor.Encrypt("same-plaintext")
		require.NoError(t, err)

		encrypted2, err := encryptor.Encrypt("same-plaintext")
		require.NoError(t, err)

		// Even with the same plaintext, ciphertexts should differ due to random nonce
		assert.NotEqual(t, encrypted1, encrypted2)

		// But both should decrypt to the same value
		decrypted1, err := encryptor.Decrypt(encrypted1)
		require.NoError(t, err)

		decrypted2, err := encryptor.Decrypt(encrypted2)
		require.NoError(t, err)

		assert.Equal(t, decrypted1, decrypted2)
	})

	t.Run("decrypt with invalid base64", func(t *testing.T) {
		_, err := encryptor.Decrypt("not-valid-base64!!!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode base64")
	})

	t.Run("decrypt with invalid ciphertext", func(t *testing.T) {
		// Create a valid base64 string that is too short
		_, err := encryptor.Decrypt("YWJjZA==") // "abcd" in base64
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})
}

func TestEncryptBytes(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	t.Run("encrypt and decrypt bytes", func(t *testing.T) {
		data := []byte("sensitive document number")

		encrypted, err := encryptor.EncryptBytes(data)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := encryptor.DecryptBytes(encrypted)
		require.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})

	t.Run("encrypt empty bytes", func(t *testing.T) {
		encrypted, err := encryptor.EncryptBytes(nil)
		require.NoError(t, err)
		assert.Nil(t, encrypted)
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("generate random key", func(t *testing.T) {
		key1, err := GenerateKey()
		require.NoError(t, err)
		assert.Len(t, key1, 32)

		key2, err := GenerateKey()
		require.NoError(t, err)
		assert.Len(t, key2, 32)

		// Keys should be different (random)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("generated key works with encryptor", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		encryptor, err := NewEncryptor(string(key))
		require.NoError(t, err)

		plaintext := "test data"
		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestMustGenerateKey(t *testing.T) {
	t.Run("must generate key panics on error", func(t *testing.T) {
		// This should not panic because GenerateKey should always succeed
		key := MustGenerateKey()
		assert.Len(t, key, 32)
	})
}

func TestEncryptorWithDifferentKeys(t *testing.T) {
	key1 := "12345678901234567890123456789011"
	key2 := "12345678901234567890123456789012"

	encryptor1, err := NewEncryptor(key1)
	require.NoError(t, err)

	encryptor2, err := NewEncryptor(key2)
	require.NoError(t, err)

	t.Run("data encrypted with key1 cannot be decrypted with key2", func(t *testing.T) {
		plaintext := "secret data"

		encrypted, err := encryptor1.Encrypt(plaintext)
		require.NoError(t, err)

		_, err = encryptor2.Decrypt(encrypted)
		assert.Error(t, err)
	})
}
