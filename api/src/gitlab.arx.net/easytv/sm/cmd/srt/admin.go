package main

import (
	"fmt"

	"gitlab.arx.net/easytv/sm"
	cli "gopkg.in/urfave/cli.v1"
)

func NewAdminCommand(service sm.AdminService) cli.Command {
	return cli.Command{
		Name:    "admin",
		Aliases: []string{"a"},
		Usage:   "actions about the admin user",
		Subcommands: []cli.Command{
			{
				Name:    "change-password",
				Aliases: []string{"cw"},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the user",
					},
					cli.StringFlag{
						Name:  "new-password",
						Usage: "The new password",
					},
					cli.StringFlag{
						Name:  "old-password",
						Usage: "The old password",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") ||
						!c.IsSet("new-password") ||
						!c.IsSet("old-password") {
						return cli.ShowSubcommandHelp(c)
					}

					user_id := c.Int64("id")
					old_password := c.String("old-password")
					new_password := c.String("new-password")

					err := service.ChangePassword(user_id, old_password, new_password)

					if err != nil {
						fmt.Printf("Failed: %v\n", err)
					} else {
						fmt.Println("Password changed successfully")
					}
					return nil
				},
			},
		},
	}
}
