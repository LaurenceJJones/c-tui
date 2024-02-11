package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	lm "github.com/charmbracelet/wish/logging"
)

func NewServer(config *Config) (*ssh.Server, error) {
	opts := []ssh.Option{
		wish.WithAddress(config.Addr),
		generatePublicKeyMiddleware(config),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			AppMiddleware(),
			lm.Middleware(),
		),
	}
	return wish.NewServer(
		opts...,
	)
}

/*
Wish cannot handle multiple public keys parameters, so we need to create a middleware to handle the public keys
and the authorized keys
*/
func generatePublicKeyMiddleware(config *Config) ssh.Option {
	return func(s *ssh.Server) error {
		return wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			return isKeyAuthorized(config.TrustedKeys, config.authorizedKeys, func(k ssh.PublicKey) bool {
				return ssh.KeysEqual(key, k)
			})
		})(s)
	}
}

func isKeyAuthorized(keyPaths []string, authorizedKeysPath string, checker func(k ssh.PublicKey) bool) bool {
	if authorizedKeysPath != "" {
		if PathContentsAuthorizesKey(authorizedKeysPath, false, checker) {
			return true
		}
	}
	for _, path := range keyPaths {
		if PathContentsAuthorizesKey(path, true, checker) {
			return true
		}
	}
	return false
}

func PathContentsAuthorizesKey(path string, firstLineOnly bool, checker func(k ssh.PublicKey) bool) bool {
	f, err := os.Open(path)
	if err != nil {
		log.Warn("failed to parse", "path", path, "error", err)
		return false
	}
	defer f.Close() // nolint: errcheck

	rd := bufio.NewReader(f)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			break
		}
		if strings.TrimSpace(string(line)) == "" {
			continue
		}
		if bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}
		upk, _, _, _, err := ssh.ParseAuthorizedKey(line)
		if err != nil {
			log.Warn("failed to parse", "path", path, "error", err)
			return false
		}
		if checker(upk) {
			return true
		}
		if firstLineOnly {
			break
		}
	}
	return false
}
