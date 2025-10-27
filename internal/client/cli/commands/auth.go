package commands

import (
	"context"
	"fmt"
	"github.com/thxhix/passKeeper/internal/client/client_services"
	"gopkg.in/urfave/cli.v1"
	"time"
)

type AuthCLICommands struct {
	s *client_services.AuthClientService
}

func NewAuthCLICommands(s *client_services.AuthClientService) *AuthCLICommands {
	return &AuthCLICommands{s: s}
}

func (cmd *AuthCLICommands) RegisterCmd() cli.Command {
	return cli.Command{
		Name:      "register",
		Usage:     "register [login] [password] — register new user",
		ArgsUsage: "[login] [password]",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if c.NArg() < 2 {
				return cli.NewExitError("usage: passKeeper register [login] [password]", 2)
			}
			login := c.Args().Get(0)
			password := c.Args().Get(1)

			if err := cmd.s.Register(ctx, login, password); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			fmt.Println("Registered successfully.")

			return nil
		},
	}
}

func (cmd *AuthCLICommands) LoginCmd() cli.Command {
	return cli.Command{
		Name:      "login",
		Usage:     "login [login] [password] — login user by credentials",
		ArgsUsage: "[login] [password]",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if c.NArg() < 2 {
				return cli.NewExitError("usage: passKeeper login [login] [password]", 2)
			}
			login := c.Args().Get(0)
			password := c.Args().Get(1)

			if err := cmd.s.Login(ctx, login, password); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			fmt.Println("Log-in successfully.")

			return nil
		},
	}
}

func (cmd *AuthCLICommands) RefreshTokenCmd() cli.Command {
	return cli.Command{
		Name:  "refresh",
		Usage: "refresh — refresh access token",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := cmd.s.RefreshToken(ctx); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			fmt.Println("Refresh successfully.")

			return nil
		},
	}
}
