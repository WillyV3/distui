package main

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
)

func main() {
	k, err := keygen.New(
		filepath.Join(".wishlist", "server"),
		keygen.WithKeyType(keygen.Ed25519),
	)
	if err != nil {
		log.Fatal("Server keypair", "err", err)
	}
	if !k.KeyPairExists() {
		if err := k.WriteKeys(); err != nil {
			log.Fatal("Server keypair", "err", err)
		}
	}

	// Direct SSH server config - skip wishlist directory listing
	server, err := wish.NewServer(
		wish.WithAddress("localhost:2234"),
		wish.WithHostKeyPEM(k.RawPrivateKey()),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return initialAppModel(), []tea.ProgramOption{
					tea.WithAltScreen(),
				}
			}),
			lm.Middleware(),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Server setup", "err", err)
	}

	// Start the direct SSH server
	log.Info("Starting TUI Template SSH server", "port", 2234)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server", "err", err)
	}
}

// App model and views will be imported from view files
