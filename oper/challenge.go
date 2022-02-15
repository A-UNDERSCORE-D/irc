package oper

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"awesome-dragon.science/go/irc/numerics"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/youmark/pkcs8"
)

// Challenge is a nice wrapper around DoChallenge that manages storing data and decoding b64 for you
type Challenge struct {
	dataBuffer  bytes.Buffer
	keypath     string
	keypassword string
}

// NewChallenge creates a new Challenge instance with the given password and path
func NewChallenge(path, password string) (*Challenge, error) {
	if path == "" || password == "" {
		return nil, errors.New("cannot have empty path or password")
	}

	return &Challenge{
		keypath:     path,
		keypassword: password,
	}, nil
}

// OnChallengeMessage is a helper to push data when RPL_RSACHALLENGE2 is received
func (c *Challenge) OnChallengeMessage(data *ircmsg.Message) error {
	if data.Command != numerics.RPL_RSACHALLENGE2 {
		return nil
	}

	return c.PushData(data.Params[1])
}

// PushData takes base64 encoded bytes and pushes the contained value onto the Challenge instance's buffer
func (c *Challenge) PushData(data string) error {
	decoded, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("could not decode b64: %w", err)
	}

	// This can never error, it only panics if you... use all the memory
	_, _ = c.dataBuffer.Write(decoded)

	return nil
}

// DoChallenge executes the below DoChallenge method with the contents of the buffer
func (c *Challenge) DoChallenge() (string, error) {
	return DoChallenge(c.keypath, c.keypassword, c.dataBuffer.Bytes())
}

// DoChallenge decrypts the ciphertext using the given password and key, returns the b64 encoded sha1 hash of the data
func DoChallenge(keypath, password string, ciphertext []byte) (string, error) {
	key, err := getChallengeKey(keypath, password)
	if err != nil {
		return "", fmt.Errorf("failed to do challenge: %w", err)
	}

	data, err := key.Decrypt(nil, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("unable to hash result: %w", err)
	}

	out := sha1.Sum(data) //nolint:gosec // Its whats required by the format protocol

	return base64.RawStdEncoding.EncodeToString(out[:]), nil
}

func getChallengeKey(path, password string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read key file: %w", err)
	}

	return decodeChallengeKey(data, password)
}

func decodeChallengeKey(pemFile []byte, password string) (*rsa.PrivateKey, error) {
	// data is PEM format, decode it
	p, rest := pem.Decode(pemFile)
	if len(rest) > 0 {
		//nolint:goerr113 // Yeah, no
		return nil, fmt.Errorf("leftover data after decoding of length %d", len(rest))
	}

	key, err := pkcs8.ParsePKCS8PrivateKeyRSA(p.Bytes, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("could not decrypt key, or key was invalid: %w", err)
	}

	return key, nil
}
