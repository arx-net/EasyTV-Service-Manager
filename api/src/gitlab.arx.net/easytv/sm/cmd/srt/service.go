package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"gitlab.arx.net/easytv/sm"
	cli "gopkg.in/urfave/cli.v1"
)

func NewServiceCommand(
	repo sm.ModuleRepository, service sm.ModuleService) cli.Command {
	return cli.Command{
		Name:    "service",
		Aliases: []string{"s"},
		Usage:   "actions about the services",
		Subcommands: []cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Action: func(c *cli.Context) error {
					modules := make([]*sm.Module, 0)

					err := repo.GetModulesOBJ(&modules)

					if err != nil {
						fmt.Println(err)
						return nil
					}

					table := tablewriter.NewWriter(os.Stdout)

					table.SetHeader([]string{"ID", "Name", "Description", "ApiKey(hashed)", "Enabled"})
					table.SetFooter([]string{"", "", "", "Total", strconv.Itoa(len(modules))})
					table.SetBorder(false)
					for _, module := range modules {
						table.Append([]string{
							strconv.FormatInt(module.ID, 10),
							module.Name,
							module.Description,
							module.ApiKey,
							strconv.FormatBool(module.Enabled)})
					}
					table.Render()

					return nil
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "The name of the service",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "The description of the task",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("name") || !c.IsSet("description") {
						return cli.ShowSubcommandHelp(c)
					}

					module, err := service.CreateService(c.String("name"), c.String("description"))

					if err != nil {
						fmt.Printf("Failed to create service err='%v'\n", err)
					} else {
						fmt.Printf("Service created with id=%v and api_key=%v\n",
							module.ID, module.ApiKey)
					}
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the service",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "The name of the service",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "The description of the task",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") || (!c.IsSet("name") && !c.IsSet("description")) {
						return cli.ShowSubcommandHelp(c)
					}

					var err error
					if c.IsSet("name") && c.IsSet("description") {
						err = service.Update(c.Int64("id"), c.String("name"), c.String("description"))
					} else if c.IsSet("name") {
						err = service.UpdateName(c.Int64("id"), c.String("name"))
					} else if c.IsSet("description") {
						err = service.UpdateDescription(c.Int64("id"), c.String("description"))
					}

					if err != nil {
						fmt.Printf("Failed to update service err='%v'\n", err)
					} else {
						fmt.Println("Service updated")
					}
					return nil
				},
			},
			{
				Name:    "set-availability",
				Aliases: []string{"sa"},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the service",
					},
					cli.BoolFlag{
						Name:  "enable",
						Usage: "Enable the service",
					},
					cli.BoolFlag{
						Name:  "disable",
						Usage: "Disable the service",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") || (!c.IsSet("enable") && !c.IsSet("disable")) {
						return cli.ShowSubcommandHelp(c)
					}

					enable := true

					if c.IsSet("enable") {
						enable = c.Bool("enable")
					} else {
						enable = !c.Bool("disable")
					}

					err := service.SetAvailability(c.Int64("id"), enable)

					if err != nil {
						fmt.Printf("Failed to update service err='%v'\n", err)
					} else if enable {
						fmt.Println("Service enabled")
					} else {
						fmt.Println("Service disabled")
					}

					return nil
				},
			},
			{
				Name: "renew-api-key",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the service",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") {
						return cli.ShowSubcommandHelp(c)
					}

					module, err := service.RenewApiKey(c.Int64("id"))

					if err != nil {
						fmt.Printf("Failed to update service err='%v'\n", err)
					} else {
						fmt.Printf("Service id=%v api-key changed to '%v'\n",
							module.ID, module.ApiKey)
					}

					return nil
				},
			},
		},
	}
}
