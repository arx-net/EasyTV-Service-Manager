package main

import (
	"fmt"
	"os"

	"gitlab.arx.net/easytv/sm"
	"gitlab.arx.net/easytv/sm/db"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/sirupsen/logrus"
)

type CommandHandler interface {
	Handle()

	Help() string
}

func main() {
	logrus.SetLevel(logrus.FatalLevel)

	// Setup database pool
	pool, err := db.Open()

	if err != nil {
		fmt.Printf("Can't connect to the database! err=%v", err)
		os.Exit(-1)
	}

	// Setup repositories and services
	admin_repo := &db.AdminRepository{Pool: pool}
	module_repo := &db.ModuleRepository{Pool: pool}
	task_repo := &db.TaskRepository{Pool: pool}
	job_repo := &db.JobRepository{Pool: pool}
	owner_repo := &db.ContentOwnerRepository{Pool: pool}

	admin_service := sm.NewAdminService(admin_repo)
	module_service := sm.NewModuleService(module_repo)
	task_service := sm.NewTaskService(task_repo, job_repo)
	owner_service := sm.NewContentOwnerService(owner_repo)

	app := cli.NewApp()
	app.Name = "srt"
	app.Usage = "Service Registration Tool"
	app.Version = "1.0.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lang",
			Value: "english",
			Usage: "language for the greeting",
		},
	}

	app.Commands = []cli.Command{
		NewAdminCommand(admin_service),
		NewServiceCommand(module_repo, module_service),
		NewTaskCommand(task_service, task_repo),
		NewContentOwnerCommand(owner_service, owner_repo),
	}

	app.Run(os.Args)
}
