package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"charm.land/wish/v2"
	"charm.land/wish/v2/logging"
	"golang.org/x/sync/errgroup"
	"load.adoublef.dev/internal/net/ssh"
)

type serveCmd struct {
	host               string
	port               string
	hostKeyPath        string
	authorizedKeysPath string
}

func (c *serveCmd) run(ctx context.Context) error {
	c.host = cmp.Or(c.host, os.Getenv("HOST"), "::")
	c.port = cmp.Or(c.port, os.Getenv("PORT"), "2222")
	c.hostKeyPath = cmp.Or(c.hostKeyPath, os.Getenv("SSH_HOST_KEY_PATH"), "ssh/id_ed25519")
	c.authorizedKeysPath = cmp.Or(c.authorizedKeysPath, os.Getenv("AUTHORIZED_KEYS_PATH"), "authorized_keys")
	// https://github.com/charmbracelet/wish/blob/main/examples/git/main.go
	// https://github.com/charmbracelet/wish/blob/main/examples/identity/main.go
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(c.host, c.port)),
		wish.WithHostKeyPath(c.hostKeyPath),
		// Reads and validates against all public keys in the authorized_keys file
		wish.WithAuthorizedKeys(c.authorizedKeysPath),
		wish.WithMiddleware(ssh.Middleware(), logging.Middleware()))
	if err != nil {
		return fmt.Errorf("new server: %v", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		ctx = context.WithoutCancel(ctx)
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			return fmt.Errorf("shutdown: %v", cmp.Or(s.Close(), err))
		}
		return nil
	})

	return g.Wait()
}
