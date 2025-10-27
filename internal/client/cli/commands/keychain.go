package commands

import (
	"context"
	"fmt"
	"github.com/thxhix/passKeeper/internal/client/client_services"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"gopkg.in/urfave/cli.v1"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

type KeychainCLICommands struct {
	s *client_services.KeychainClientService
}

func NewKeychainCLICommands(s *client_services.KeychainClientService) *KeychainCLICommands {
	return &KeychainCLICommands{s: s}
}

func (cmd *KeychainCLICommands) Add() cli.Command {
	return cli.Command{
		Name:  "add",
		Usage: "Добавить новую запись в хранилище паролей",
		Subcommands: []cli.Command{
			{
				Name:      "credential",
				Usage:     "passKeeper add credential [title] [login] [password] [site] [note]",
				ArgsUsage: "[title] [login] [password] [site(optional)] [note(optional)]",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if c.NArg() < 3 {
						fmt.Println("Ошибка: нужно указать как минимум title, login и password")
						return nil
					}

					title := c.Args().Get(0)
					login := c.Args().Get(1)
					password := c.Args().Get(2)
					site := c.Args().Get(3)
					note := c.Args().Get(4)

					if err := cmd.s.AddCredential(ctx, title, login, password, site, note); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}

					fmt.Println("✅ Успешно добавлено!")
					return nil
				},
			},

			{
				Name:      "card",
				Usage:     "passKeeper add card [title] [number] [expDate] [cvv] [holder] [bank] [note]",
				ArgsUsage: "[title] [number] [expDate] [cvv] [holder(optional)] [bank(optional)] [note(optional)]",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if c.NArg() < 4 {
						fmt.Println("Ошибка: нужно указать как минимум title, number, expDate и cvv")
						return nil
					}

					title := c.Args().Get(0)
					number := c.Args().Get(1)
					expDate := c.Args().Get(2)
					cvv := c.Args().Get(3)
					holder := c.Args().Get(4)
					bank := c.Args().Get(5)
					note := c.Args().Get(6)

					if err := cmd.s.AddCard(ctx, title, number, expDate, cvv, holder, bank, note); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}

					fmt.Println("✅ Успешно добавлено!")
					return nil
				},
			},

			{
				Name:      "text",
				Usage:     "passKeeper add text [title] [text] [note]",
				ArgsUsage: "[title] [text] [note(optional)]",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if c.NArg() < 2 {
						fmt.Println("Ошибка: нужно указать как минимум title и text")
						return nil
					}

					title := c.Args().Get(0)
					text := c.Args().Get(1)
					note := c.Args().Get(2)

					if err := cmd.s.AddText(ctx, title, text, note); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}

					fmt.Println("✅ Успешно добавлено!")
					return nil
				},
			},

			{
				Name:      "file",
				Usage:     "passKeeper add file [title] [filePath] [note]",
				ArgsUsage: "[title] [filePath] [note(optional)]",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()

					if c.NArg() < 2 {
						fmt.Println("Ошибка: нужно указать как минимум title и путь к файлу")
						return nil
					}

					title := c.Args().Get(0)
					filePath := c.Args().Get(1)
					note := ""
					if c.NArg() >= 3 {
						note = c.Args().Get(2)
					}

					if err := cmd.s.AddFile(ctx, title, filePath, note); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}

					fmt.Println("✅ Файл успешно загружен!")
					return nil
				},
			},
		},
	}
}

func (cmd *KeychainCLICommands) List() cli.Command {
	allowedTypes := keychain.AllKeyTypes

	types := make([]string, len(allowedTypes))
	for i, t := range allowedTypes {
		types[i] = string(t)
	}

	return cli.Command{
		Name:      "list",
		Usage:     "passKeeper list [type]",
		ArgsUsage: "[type:" + strings.Join(types, "|") + "]",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if c.NArg() > 1 {
				return cli.NewExitError("usage: passKeeper list [type]", 2)
			}
			keyType := c.Args().Get(0)

			resp, err := cmd.s.GetList(ctx, keyType)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			if len(resp.Keys) == 0 {
				fmt.Println("Нет сохранённых элементов.")
				return nil
			}

			// Настраиваем табличный вывод
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, err = fmt.Fprintf(w, "UUID\tTYPE\tTITLE\tCREATED_AT")
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			for _, rec := range resp.Keys {
				_, err = fmt.Fprintf(
					w,
					"%s\t%s\t%s\t%s\n",
					rec.KeyUUID,
					rec.KeyType,
					rec.Title,
					rec.CreatedAt.Format("2006-01-02 15:04:05"),
				)
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
			}

			_ = w.Flush()
			return nil
		},
	}
}

func (cmd *KeychainCLICommands) Get() cli.Command {
	return cli.Command{
		Name:      "get",
		Usage:     "get [key_uuid]",
		ArgsUsage: "[key_uuid]",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if c.NArg() > 1 {
				return cli.NewExitError("usage: passKeeper get [key_uuid]", 2)
			}
			keyUUID := c.Args().Get(0)

			resp, err := cmd.s.Get(ctx, keyUUID)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, err = fmt.Fprintln(w, "UUID\tTYPE\tTITLE\tDATA\tCREATED_AT\tUPDATED_AT")
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			_, err = fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\t%s\n",
				resp.KeyUUID,
				resp.KeyType,
				resp.Title,
				resp.Data,
				resp.CreatedAt.Format("2006-01-02 15:04:05"),
				resp.UpdatedAt.Format("2006-01-02 15:04:05"),
			)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			_ = w.Flush()
			return nil
		},
	}
}

func (cmd *KeychainCLICommands) Delete() cli.Command {
	return cli.Command{
		Name:      "delete",
		Usage:     "delete [key_uuid]",
		ArgsUsage: "[key_uuid]",

		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if c.NArg() > 1 {
				return cli.NewExitError("usage: passKeeper delete [key_uuid]", 2)
			}
			keyUUID := c.Args().Get(0)

			err := cmd.s.Delete(ctx, keyUUID)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			fmt.Println("✅ Успешно удалено!")
			return nil
		},
	}
}
