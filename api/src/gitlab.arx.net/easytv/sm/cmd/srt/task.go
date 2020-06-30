package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"gitlab.arx.net/easytv/sm"
	cli "gopkg.in/urfave/cli.v1"
)

func NewTaskCommand(service sm.TaskService, repository sm.TaskRepository) cli.Command {
	return cli.Command{
		Name:    "task",
		Aliases: []string{"t"},
		Usage:   "Actions about the task",
		Subcommands: []cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new task",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "service",
						Usage: "The id of the service this task should belong to",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "The name of the task",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "The description of the task",
					},
					cli.StringFlag{
						Name:  "start-url",
						Usage: "The url to use to start the task",
					},
					cli.StringFlag{
						Name:  "cancel-url",
						Usage: "The url to use to cancel the task",
					},
					cli.StringFlag{
						Name:  "input",
						Usage: "The input of the task encoded in JSON",
					},
					cli.StringFlag{
						Name:  "output",
						Usage: "The output of the task encoded in JSON",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("name") ||
						!c.IsSet("description") ||
						!c.IsSet("service") ||
						!c.IsSet("start-url") ||
						!c.IsSet("cancel-url") ||
						!c.IsSet("input") ||
						!c.IsSet("output") {
						return cli.ShowSubcommandHelp(c)
					}

					var input_json map[string]interface{}
					err := json.Unmarshal([]byte(c.String("input")), &input_json)
					if err != nil {
						fmt.Printf("Failed to parse 'input' err='%v'\n", err)
						return nil
					}

					input := make(map[string]sm.ParamType)

					for name, value := range input_json {
						str_val, ok := value.(string)
						if !ok {
							fmt.Printf("\"%s\" should have a string value\n", name)
							return nil
						}
						input[name] = sm.GetParamTypeFromSting(str_val)
					}

					var output_json map[string]interface{}
					err = json.Unmarshal([]byte(c.String("output")), &output_json)
					if err != nil {
						fmt.Printf("Failed to parse 'output' err='%v'\n", err)
						return nil
					}

					output := make(map[string]sm.ParamType)

					for name, value := range output_json {
						str_val, ok := value.(string)
						if !ok {
							fmt.Printf("\"%s\" should have a string value\n", name)
							return nil
						}
						output[name] = sm.GetParamTypeFromSting(str_val)
					}

					task, err := service.RegisterTask(
						c.Int64("service"),
						c.String("name"),
						c.String("description"),
						c.String("start-url"),
						c.String("cancel-url"),
						input,
						output,
					)

					if err != nil {
						fmt.Printf("Failed to create task err='%v'\n", err)
						return nil
					} else {
						fmt.Printf("Created task with id=%v\n", task.ID)
					}

					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List the tasks of a service",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "service",
						Usage: "The id of the service this task should belong to",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("service") {
						return cli.ShowSubcommandHelp(c)
					}

					tasks, err := repository.GetTasks(c.Int64("service"), false)

					if err != nil {
						fmt.Printf("Failed to get tasks err='%v'", err)
						return nil
					}

					table := tablewriter.NewWriter(os.Stdout)

					table.SetHeader([]string{"ID", "Name", "Description", "StartUrl", "CancelUrl", "Enabled", "Input", "Output"})
					table.SetFooter([]string{"", "", "", "", "", "", "Total", strconv.Itoa(len(tasks))})
					table.SetBorder(false)
					for _, task := range tasks {
						input_json, _ := json.Marshal(task.Input)
						output_json, _ := json.Marshal(task.Output)

						table.Append([]string{
							strconv.FormatInt(task.ID, 10),
							task.Name,
							task.Description,
							task.StartUrl,
							task.CancelUrl,
							strconv.FormatBool(task.Enabled),
							string(input_json),
							string(output_json)})
					}
					table.Render()

					return nil
				},
			},
			{
				Name:    "set-availability",
				Aliases: []string{"sa"},
				Usage:   "Change the availability of the task",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the task",
					},
					cli.BoolFlag{
						Name:  "enable",
						Usage: "Enable the task",
					},
					cli.BoolFlag{
						Name:  "disable",
						Usage: "Disable the task",
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
						fmt.Printf("Failed to update task err='%v'\n", err)
					} else if enable {
						fmt.Println("Task enabled")
					} else {
						fmt.Println("Task disabled")
					}

					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "Update the task",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the task",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "Change the name",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "Change the description",
					},
					cli.StringFlag{
						Name:  "start-url",
						Usage: "Change the start url",
					},
					cli.StringFlag{
						Name:  "cancel-url",
						Usage: "Change the cancel url",
					},
					cli.StringFlag{
						Name:  "cancel-rest-url",
						Usage: "Set the cancel url to a REST endpoint",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") ||
						(!c.IsSet("name") && !c.IsSet("description") && !c.IsSet("start-url") && !c.IsSet("cancel-url")) {
						return cli.ShowSubcommandHelp(c)
					}

					fields := make(map[string]string)

					if c.IsSet("name") {
						fields["Name"] = c.String("name")
					}

					if c.IsSet("description") {
						fields["Description"] = c.String("description")
					}

					if c.IsSet("start-url") {
						fields["StartUrl"] = c.String("start-url")
					}

					if c.IsSet("cancel-url") {
						fields["CancelUrl"] = c.String("cancel-url")
					}

					if c.IsSet("cancel-rest-url") {
						fields["CancelUrl"] = fmt.Sprintf("REST %v", c.String("cancel-rest-url"))
					}

					fmt.Println(fields)

					err := service.Update(c.Int64("id"), fields)
					if err != nil {
						fmt.Printf("Failed to update task err='%v'\n", err)
					} else {
						fmt.Println("Task was updated")
					}
					return nil
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove the task",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the task",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") {
						return cli.ShowSubcommandHelp(c)
					}

					err := service.DeleteTask(c.Int64("id"))

					if err != nil {
						fmt.Printf("Failed to delete task err='%v'\n", err)
					} else {
						fmt.Println("The task was deleted")
					}

					return nil
				},
			},
			{
				Name:  "set-vars",
				Usage: "Change the input or output of the task",
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "id",
						Usage: "The id of the task",
					},
					cli.StringFlag{
						Name:  "input",
						Usage: "The input of the task",
					},
					cli.StringFlag{
						Name:  "output",
						Usage: "The output of the task",
					},
				},
				Action: func(c *cli.Context) error {
					if !c.IsSet("id") || (!c.IsSet("input") && !c.IsSet("output")) {
						return cli.ShowSubcommandHelp(c)
					}

					if c.IsSet("input") {
						var input_json map[string]interface{}
						err := json.Unmarshal([]byte(c.String("input")), &input_json)
						if err != nil {
							fmt.Printf("Failed to parse 'input' err='%v'\n", err)
							return nil
						}

						input := make(map[string]sm.ParamType)

						for name, value := range input_json {
							str_val, ok := value.(string)
							if !ok {
								fmt.Printf("\"%s\" should have a string value\n", name)
								return nil
							}
							input[name] = sm.GetParamTypeFromSting(str_val)
						}

						err = service.UpdateVars(c.Int64("id"), input, true)

						if err != nil {
							fmt.Printf("Failed to update input err='%v'\n", err)
							return nil
						}
					}

					if c.IsSet("output") {
						var output_json map[string]interface{}
						err := json.Unmarshal([]byte(c.String("output")), &output_json)
						if err != nil {
							fmt.Printf("Failed to parse 'output' err='%v'\n", err)
							return nil
						}

						output := make(map[string]sm.ParamType)

						for name, value := range output_json {
							str_val, ok := value.(string)
							if !ok {
								fmt.Printf("\"%s\" should have a string value\n", name)
								return nil
							}
							output[name] = sm.GetParamTypeFromSting(str_val)
						}

						err = service.UpdateVars(c.Int64("id"), output, false)

						if err != nil {
							fmt.Printf("Failed to update output err='%v'\n", err)
							return nil
						}
					}

					fmt.Println("The task was updated")

					return nil
				},
			},
		},
	}
}
