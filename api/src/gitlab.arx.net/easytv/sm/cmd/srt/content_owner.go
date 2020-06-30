package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"gitlab.arx.net/easytv/sm"
	cli "gopkg.in/urfave/cli.v1"
)

func NewContentOwnerCommand(
	service sm.ContentOwnerService, repository sm.ContentOwnerRepository) cli.Command {

	return cli.Command{
		Name:    "content-owner",
		Aliases: []string{"co"},
		Usage:   "actions about the content-owner user",
		Subcommands: []cli.Command{
			{
				Name:    "reset-password",
				Aliases: []string{"rp"},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the user",
					},
					cli.StringFlag{
						Name:  "password",
						Usage: "(Optional) The new password",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") {
						return cli.ShowSubcommandHelp(c)
					}

					var password string
					if c.IsSet("password") {
						password = c.String("password")
					} else {
						password = service.GenerateRandomPassword(12)
					}

					err := service.ResetPassword(c.Int64("id"), password)

					if err != nil {
						fmt.Printf("Failed to reset password err='%v'\n", err)
					} else {
						fmt.Printf("Password changed successfully to %v\n", password)
					}
					return nil
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "The name of the content owner",
					},
					cli.StringFlag{
						Name:  "username",
						Usage: "The username for the content owner",
					},
					cli.StringFlag{
						Name:  "email",
						Usage: "The email of the content owner",
					},
					cli.StringFlag{
						Name:  "password",
						Usage: "(Optional) The new password",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("name") || !c.IsSet("username") || !c.IsSet("email") {
						return cli.ShowSubcommandHelp(c)
					}

					var password string
					if c.IsSet("password") {
						password = c.String("password")
					} else {
						password = service.GenerateRandomPassword(12)
					}

					user, err := service.CreateContentOwner(
						c.String("name"),
						c.String("username"),
						c.String("email"),
						password)

					if err != nil {
						fmt.Printf("Failed to create user err='%v'\n", err)
					} else {
						fmt.Printf("Create user id=%v username=%v with password=%v (Last time you're going to see this as plain text)\n",
							user.ID, user.Username, password)
					}
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Action: func(c *cli.Context) error {
					users, err := repository.GetAll()

					if err != nil {
						fmt.Println(err)
						return nil
					}

					table := tablewriter.NewWriter(os.Stdout)

					table.SetHeader([]string{"ID", "Name", "Username", "Email"})
					table.SetFooter([]string{"", "", "Total", strconv.Itoa(len(users))})
					table.SetBorder(false)
					for _, user := range users {
						table.Append([]string{
							strconv.FormatInt(user.ID, 10),
							user.Name,
							user.Username,
							user.Email})
					}
					table.Render()

					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "the id of the content owner",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "update the name of the content owner",
					},
					cli.StringFlag{
						Name:  "username",
						Usage: "update the username of the content owner",
					},
					cli.StringFlag{
						Name:  "email",
						Usage: "update the email of the content owner",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") ||
						(!c.IsSet("name") && !c.IsSet("username") && !c.IsSet("email")) {
						cli.ShowSubcommandHelp(c)
					}

					fields := make(map[string]string)

					if c.IsSet("name") {
						fields["Name"] = c.String("name")
					}

					if c.IsSet("username") {
						fields["Username"] = c.String("username")
					}

					if c.IsSet("email") {
						fields["Email"] = c.String("email")
					}

					err := service.Update(c.Int64("id"), fields)
					if err != nil {
						fmt.Printf("Failed to update task err='%v'\n", err)
					} else {
						fmt.Println("Task was updated")
					}

					return nil
				},
			},
		},
	}
}
