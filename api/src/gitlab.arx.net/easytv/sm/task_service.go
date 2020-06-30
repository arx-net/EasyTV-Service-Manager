package sm

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type task_service struct {
	repository     TaskRepository
	job_repository JobRepository
}

func NewTaskService(
	repository TaskRepository, job_repository JobRepository) TaskService {
	return &task_service{
		repository:     repository,
		job_repository: job_repository,
	}
}

func (this *task_service) RegisterTask(
	module_id int64,
	name, description, start_url, cancel_url string,
	input, output map[string]ParamType) (*Task, error) {

	if len(name) <= 1 {
		return nil, ErrTaskNameTooShort
	} else if len(description) <= 1 {
		return nil, ErrTaskDescTooShort
	} else if !(strings.HasPrefix(start_url, "http://") ||
		strings.HasPrefix(start_url, "https://")) {
		return nil, ErrInvalidStartUrl
	} else if !(strings.HasPrefix(cancel_url, "http://") ||
		strings.HasPrefix(cancel_url, "https://") ||
		strings.HasPrefix(cancel_url, "REST http://") ||
		strings.HasPrefix(cancel_url, "REST https://")) {
		return nil, ErrInvalidCancelUrl
	} else if len(input) == 0 {
		return nil, ErrEmptyInput
	} else if len(output) == 0 {
		return nil, ErrEmptyOutput
	}

	exists, err := this.repository.NameExists(name)

	if err != nil {
		return nil, err
	} else if exists {
		return nil, ErrTaskAlreadyExists
	}

	task := Task{
		ModuleID:    module_id,
		Name:        name,
		Description: description,
		StartUrl:    start_url,
		CancelUrl:   cancel_url,
		Input:       make(map[string]ParamType),
		Output:      make(map[string]ParamType),
	}

	for name, value := range input {
		if value != StringParam &&
			value != IntParam &&
			value != DoubleParam {
			return nil, ErrInvalidInputType
		}
		task.Input[name] = value
	}

	for name, value := range output {
		if value != StringParam &&
			value != IntParam &&
			value != DoubleParam {
			return nil, ErrInvalidOutputType
		}
		task.Output[name] = value
	}

	log.Infof("register task name=%v", task.Name)

	err = this.repository.CreateTask(&task)
	if err != nil {
		return nil, err
	}
	log.Infof("task name=%v id=%v registed", task.Name, task.ID)

	return &task, nil
}

func (this *task_service) SetAvailability(id int64, enabled bool) error {
	task, err := this.repository.GetTask(id)

	if err != nil {
		return err
	} else if task == nil || task.Deleted {
		return ErrNotFound
	}

	return this.repository.SetAvailability(id, enabled)
}

func (this *task_service) DeleteTask(id int64) error {
	task, err := this.repository.GetTask(id)

	if err != nil {
		return err
	} else if task == nil || task.Deleted {
		return ErrNotFound
	} else if task.Enabled {
		return ErrTaskIsEnabled
	}

	has_jobs, err := this.job_repository.TaskHasActiveJobs(task.ID)

	if err != nil {
		return err
	} else if has_jobs {
		return ErrTaskHasJobStepsInProgress
	}

	return this.repository.DeleteTask(task.ID)
}

func (this *task_service) Update(id int64, fields map[string]string) error {
	task, err := this.repository.GetTask(id)

	if err != nil {
		return err
	} else if task == nil || task.Deleted {
		return ErrNotFound
	}

	is_active, err := this.job_repository.TaskHasActiveJobs(task.ID)

	if err != nil {
		return err
	}

	for name, value := range fields {
		if name == "Name" {
			task.Name = value
		} else if name == "Description" {
			task.Description = value
		} else if task.Enabled {
			// For the rest of the field only allow updates if the task is disabled
			return ErrTaskIsEnabled
		} else if is_active {
			// For the rest of the fields only allows update if the task is not currently used.
			return ErrTaskHasJobStepsInProgress
		} else if name == "StartUrl" {
			task.StartUrl = value
		} else if name == "CancelUrl" {
			task.CancelUrl = value
		}
	}

	return this.repository.Save(task)
}

func (this *task_service) UpdateVars(id int64, data map[string]ParamType, is_input bool) error {
	task, err := this.repository.GetTask(id)

	if err != nil {
		return nil
	} else if task == nil || task.Deleted {
		return ErrNotFound
	} else if task.Enabled {
		return ErrTaskIsEnabled
	}

	is_active, err := this.job_repository.TaskHasActiveJobs(task.ID)

	if err != nil {
		return err
	} else if is_active {
		return ErrTaskHasJobStepsInProgress
	}

	if is_input {
		task.Input = make(map[string]ParamType)
	} else {
		task.Output = make(map[string]ParamType)
	}

	for name, value := range data {
		if value != StringParam &&
			value != IntParam &&
			value != DoubleParam {
			if is_input {
				return ErrInvalidInputType
			} else {
				return ErrInvalidOutputType
			}
		}
		if is_input {
			task.Input[name] = value
		} else {
			task.Output[name] = value
		}
	}

	return this.repository.SaveVars(task, is_input)
}
