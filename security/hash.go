package security

import (
	"crypto/rand"
	"golang.org/x/crypto/argon2"
)

type Params struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

func HashPassword(password string, p *Params) (hash []byte, err error) {
    // Generate a cryptographically secure random salt.
    salt, err := generateSalt(p.SaltLength)
    if err != nil {
        return nil, err
    }

    // Pass the plaintext password, salt and parameters to the argon2.IDKey
    // function. This will generate a hash of the password using the Argon2id
    // variant.
    hash = argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

    return hash, nil
}

func generateSalt( n uint32) ([]byte, error) {
	b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }

    return b, nil
}
