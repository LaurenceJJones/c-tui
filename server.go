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
	if config.authorizedKeys != "" {
		opts = append(opts, wish.WithAuthorizedKeys(config.authorizedKeys))
	}
	return wish.NewServer(
		opts...,
	)
}

// Wish cannot handle a list of public keys, so we need to create a middleware to handle the public keys
func generatePublicKeyMiddleware(config *Config) ssh.Option {
	return func(s *ssh.Server) error {
		return wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			return isKeyAuthorized(config.TrustedKeys, func(k ssh.PublicKey) bool {
				return ssh.KeysEqual(key, k)
			})
		})(s)
	}
}

func isKeyAuthorized(paths []string, checker func(k ssh.PublicKey) bool) bool {
	for _, path := range paths {
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
			break // Public keys should be single line in the file so we break on first non empty and non comment line
		}
	}
	return false
}
