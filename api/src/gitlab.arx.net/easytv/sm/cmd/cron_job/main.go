package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"gitlab.arx.net/easytv/sm"

	"gitlab.arx.net/easytv/sm/db"
)

const BATCH_LIMIT = 10000

func main() {
	// Setup logging
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)

	os.MkdirAll("/var/log/sm", os.ModePerm)
	f, err := os.OpenFile("/var/log/sm/sm.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	log.SetOutput(f)

	// Pool
	pool, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	job_repository := &db.JobRepository{Pool: pool}
	task_repository := &db.TaskRepository{Pool: pool}
	module_repository := &db.ModuleRepository{Pool: pool}
	asset_repository := &db.AssetRepository{Pool: pool}
	owner_repository := &db.ContentOwnerRepository{Pool: pool}

	job_service := sm.NewJobService(job_repository,
		task_repository,
		module_repository,
		owner_repository)

	asset_service := sm.NewAssetService(asset_repository, job_repository, task_repository)

	log.Print("Started periodic cleanup of jobs")
	log.Print("Cancel jobs exceeding publication date...")

	if err = job_service.CancelJobsWithExceedingPublicationDate(); err != nil {
		log.Fatal(err)
	}

	log.Print("Clean job resources of expired jobs...")
	if err = asset_service.GC(); err != nil {
		log.Fatal(err)
	}

	log.Print("Completed")
}
