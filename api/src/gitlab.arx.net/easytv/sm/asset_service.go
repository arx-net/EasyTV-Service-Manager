package sm

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type asset_service struct {
	task_repository TaskRepository
	job_repository  JobRepository
	repository      AssetRepository
}

func NewAssetService(repository AssetRepository,
	job_repository JobRepository,
	task_repository TaskRepository) AssetService {
	return &asset_service{
		repository:      repository,
		task_repository: task_repository,
		job_repository:  job_repository,
	}
}

func (this *asset_service) GC() error {
	now := time.Now()

	BATCH_LIMIT := int64(10000)

	// Get jobs in batches of <LIMIT>
	jobs, err := this.job_repository.GetNewExpiredJobsAt(now, BATCH_LIMIT, 0)

	if err != nil {
		return err
	}
	log.Infof("Executing GC for expired jobs")

	for len(jobs) > 0 {
		log.Infof("GC: Batch of %d jobs", len(jobs))
		for _, job := range jobs {
			log.Infof("GC: Clean job %d", job.ID)

			// Delete stored input and output params
			err = this.job_repository.DeleteParamsForJob(job.ID)
			if err != nil {
				log.Warningf("GC: Failed to delete params for job %d (%s)", job.ID, err)
			}

			// Delete asset records
			err = this.repository.DeleteAssetsForJob(job.ID)

			if err != nil {
				log.Warningf("GC: Failed to delete asset entries for job %d (%s)", job.ID, err)
			}

			// Delete asset files
			job_id_hash := md5.Sum([]byte(strconv.FormatInt(job.ID, 10)))

			asset_folder := fmt.Sprintf("/asset/%s/%d",
				hex.EncodeToString(job_id_hash[:1]),
				job.ID)

			err = os.RemoveAll(asset_folder)

			if err != nil {
				log.Warningf("GC: Failed to remove asset folder \"%s\" for job %d (%s)",
					asset_folder,
					job.ID,
					err)
			}
		}

		jobs, err = this.job_repository.GetNewExpiredJobsAt(
			now,
			BATCH_LIMIT,
			jobs[len(jobs)-1].ID)

		if err != nil {
			return err
		}
	}

	log.Infof("GC: Completed")

	return this.job_repository.ExpireJobsBefore(now)
}

func (this *asset_service) CreateAsset(step_id int64,
	module *Module,
	file multipart.File,
	filename string,
	filesize int64) (*Asset, error) {
	job, err := this.job_repository.GetJobByID(step_id)

	if err != nil {
		return nil, err
	} else if job == nil {
		return nil, ErrNotFound
	} else if job.IsCompleted {
		return nil, ErrJobIsCompleted
	} else if job.IsCanceled {
		return nil, ErrJobIsCanceled
	}

	if err = this.job_repository.GetJobSteps(job.ID, &job.Steps); err != nil {
		return nil, err
	}

	step := job.Steps[job.CurrentStep]
	task, err := this.task_repository.GetTask(step.TaskID)

	// Check that this module is responsible for this job step
	if err != nil {
		return nil, err
	} else if task == nil {
		return nil, fmt.Errorf("Task %d doesn't exist even though step %d claims so",
			step.TaskID, step.ID)
	} else if task.ModuleID != module.ID {
		// as far as this service is concerned this job doesn't exist
		return nil, ErrNotFound
	}

	// Create Asset object and save the file
	job_id_hash := md5.Sum([]byte(strconv.FormatInt(step_id, 10)))

	url_hex := md5.Sum([]byte(fmt.Sprintf("%d/%s", job.ID, filename)))

	log.Infof("saving asset filename=%v for step=%v", filename, step_id)

	asset := Asset{
		JobID: job.ID,
		Path: fmt.Sprintf("/asset/%s/%d/%s",
			hex.EncodeToString(job_id_hash[:1]),
			job.ID,
			filename),
		Size:     filesize,
		UrlParam: base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(url_hex[:]),
	}

	os.MkdirAll(filepath.Dir(asset.Path), os.ModePerm)
	f, err := os.OpenFile(asset.Path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(f, file)
	f.Close()

	// Save Asset to DB
	if err = this.repository.Create(&asset); err != nil {
		os.Remove(asset.Path)
		return nil, err
	}

	log.Infof("asset file=%v saved with id=Creating%v", filename, asset.ID)

	return &asset, nil
}
