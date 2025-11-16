package age

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"filippo.io/age"

	. "github.com/k4lipso/gokill/internal"
)

var (
	ageKeyFileName = "age.key"
)

type Key *age.X25519Identity

func GenerateAgeKey(filename string) (*age.X25519Identity, error) {
	// Generate a new X25519 identity (private key)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	// Convert the identity to its string representation
	privateKey := identity.String()

	// Write the private key to a file
	err = os.WriteFile(filename, []byte(privateKey), 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to save private key to file: %w", err)
	}

	Log.Infof("Private key saved to %s\n", filename)
	return identity, nil
}

func LoadAgeKey(filename string) (*age.X25519Identity, error) {
	// Read the private key from the file
	privateKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from file: %w", err)
	}

	// Parse the private key
	identity, err := age.ParseX25519Identity(string(privateKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return identity, nil
}

func LoadOrGenerateKeys(filename string) (*age.X25519Identity, error) {
	_, err := os.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		Log.Info("No Key found. Generating Key")
		return GenerateAgeKey(filename)
	}

	return LoadAgeKey(filename)
}

func Decrypt(encryptedData []byte, identity *age.X25519Identity) ([]byte, error) {
	// Create a new decryptor using the recipient's identity
	decryptor, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create decryptor: %w", err)
	}

	// Read the decrypted data
	decryptedData, err := io.ReadAll(decryptor)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return decryptedData, nil
}

func Encrypt(data []byte, recipientKeys []string) ([]byte, error) {
	var recipients []age.Recipient

	// Parse all recipient keys
	for _, key := range recipientKeys {
		recipient, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse recipient key: %w", err)
		}
		recipients = append(recipients, recipient)
	}

	// Create a new buffer to hold the encrypted data
	var encryptedData bytes.Buffer

	// Create a new age encryptor for the recipients
	encryptor, err := age.Encrypt(&encryptedData, recipients...)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Write the data to the encryptor
	_, err = encryptor.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Close the encryptor to finalize the encryption process
	if err := encryptor.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize encryption: %w", err)
	}

	return encryptedData.Bytes(), nil
}
