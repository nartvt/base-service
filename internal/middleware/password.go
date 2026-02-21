package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// =============================================================================
// Password Hasher Implementation
// clean-arch: Implements PasswordHasher interface with Argon2id algorithm
// =============================================================================

// Argon2id parameters - based on OWASP recommendations for 2024
// These provide ~50-100ms hash time on modern hardware
const (
	argon2Time       = 3         // Number of iterations
	argon2Memory     = 64 * 1024 // 64 MB of memory
	argon2Threads    = 2         // Number of parallel threads
	argon2KeyLength  = 32        // Length of the derived key
	argon2SaltLength = 16        // Length of the random salt
)

// Compile-time interface compliance check
var _ PasswordHasher = (*Argon2PasswordHasher)(nil)

// Argon2PasswordHasher implements PasswordHasher using Argon2id algorithm.
type Argon2PasswordHasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen int
}

// NewArgon2PasswordHasher creates a new Argon2id password hasher with default settings.
func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{
		time:    argon2Time,
		memory:  argon2Memory,
		threads: argon2Threads,
		keyLen:  argon2KeyLength,
		saltLen: argon2SaltLength,
	}
}

// NewArgon2PasswordHasherWithConfig creates a new Argon2id password hasher with custom settings.
func NewArgon2PasswordHasherWithConfig(time, memory uint32, threads uint8, keyLen uint32, saltLen int) *Argon2PasswordHasher {
	return &Argon2PasswordHasher{
		time:    time,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
		saltLen: saltLen,
	}
}

// Hash generates an argon2id hash of the password.
// Returns: base64-encoded string in format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate a cryptographically secure random salt
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.time,
		h.memory,
		h.threads,
		h.keyLen,
	)

	// Encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return in PHC string format for easy parsing and future-proofing
	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.memory,
		h.time,
		h.threads,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify checks if a password matches the hash.
// Returns true if the password matches, false otherwise.
func (h *Argon2PasswordHasher) Verify(password, encodedHash string) (bool, error) {
	if password == "" {
		return false, errors.New("password cannot be empty")
	}
	if encodedHash == "" {
		return false, errors.New("hash cannot be empty")
	}

	// Parse the encoded hash to extract salt and parameters
	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	// Verify algorithm
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported algorithm")
	}

	// Parse version
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}

	// Parse parameters: m=memory,t=time,p=threads
	var memory, time, threads uint32
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode stored hash
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Hash the input password with the SAME salt and parameters
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		uint8(threads),
		uint32(len(storedHash)),
	)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(storedHash, newHash) == 1, nil
}
