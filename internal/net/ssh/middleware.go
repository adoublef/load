package ssh

import (
	"fmt"

	"charm.land/ssh"
	"charm.land/wish/v2"
)

func Middleware( /* args */ ) wish.Middleware {
	return func(f ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			fmt.Fprintf(s, "Hello, world!\n")
			f(s)
		}
	}
}

type (
	Context   = ssh.Context
	PublicKey = ssh.PublicKey
	Option    = ssh.Option
)

var ErrServerClosed = ssh.ErrServerClosed

func Equal(a, b ssh.PublicKey) bool {
	return ssh.KeysEqual(a, b)
}

func Key(in []byte) ssh.PublicKey {
	parsed, _, _, _, _ := ssh.ParseAuthorizedKey(in)
	return parsed
}
