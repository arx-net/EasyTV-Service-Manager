package sm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"

	"net/http"
	"time"
)

type jservice struct {
	repository        JobRepository
	task_repository   TaskRepository
	module_repository ModuleRepository
	owner_repository  ContentOwnerRepository
}

func NewJobService(repository JobRepository,
	task_repository TaskRepository,
	module_repository ModuleRepository,
	owner_repository ContentOwnerRepository) JobService {
	return &jservice{
		repository:        repository,
		task_repository:   task_repository,
		module_repository: module_repository,
		owner_repository:  owner_repository,
	}
}

func (this *jservice) SetJobStatusForStep(step_id int64, status string) error {
	log.Infof("step=%v set status='%v'", step_id, status)
	job, err := this.repository.GetJobByStepID(step_id)

	if err != nil {
		return err
	} else if job == nil {
		return ErrNotFound
	} else if job.IsCompleted || job.IsCanceled {
		return ErrJobStatusNotUpdatable
	}

	err = this.repository.GetJobSteps(job.ID, &job.Steps)
	if err != nil {
		return err
	}

	if job.Steps[job.CurrentStep].ID != step_id {
		return ErrJobStatusNotUpdatable
	}

	job.Status = status

	return this.repository.SaveStatus(job)
}

// Sends a cancel request for the current job step
func (this *jservice) SendCancelRequest(
	job *Job, task *Task, module *Module) error {
	log.Infof("job=%v step=%v send cancel request url=%v",
		job.ID,
		job.Steps[job.CurrentStep].ID,
		task.CancelUrl)
	json_data, _ := json.Marshal(map[string]interface{}{
		"job_id": job.Steps[job.CurrentStep].ID,
		"action": "cancel",
	})

	client := http.Client{}

	var req *http.Request
	var err error

	if strings.HasPrefix(task.CancelUrl, "REST") {
		// For rest endpoints send a DELETE request
		url := fmt.Sprintf("%v/%v", task.CancelUrl[5:], job.Steps[job.CurrentStep].ID)
		req, err = http.NewRequest("DELETE", url, bytes.NewBuffer(json_data))
	} else {
		req, err = http.NewRequest("POST", task.CancelUrl, bytes.NewBuffer(json_data))
	}

	if err != nil {
		return err
	}

	req.Header.Add(EasyTVApiKeyHeader, module.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		log.Infof("job=%v cancel request failed status=%v",
			job.ID,
			resp.StatusCode)
		resp.Body.Close()
		return nil
	}

	json_data, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil
	}

	// Parse resposne json
	var data map[string]interface{}
	err = json.Unmarshal(json_data, &data)
	if err != nil {
		return nil
	}

	code, _ := data["code"].(float64)
	description, _ := data["description"].(string)
	log.Infof("job=%v sent cancel request code=%.0f description='%s'",
		job.ID, code, description)

	return nil
}

// CancelJob as a specific module
func (this *jservice) CancelJobAsModule(module *Module, step_id int64) error {
	log.Infof("step=%v cancel as module=%v,'%v'", step_id, module.ID, module.Name)
	job, err := this.repository.GetJobByStepID(step_id)
	if err != nil {
		return err
	} else if job == nil {
		return ErrNotFound
	}

	if job.IsCompleted {
		return ErrJobIsCompleted
	} else if job.IsCanceled {
		return ErrJobIsCanceled
	}

	// Get the task for the current job step
	err = this.repository.GetJobSteps(job.ID, &job.Steps)

	if err != nil {
		return err
	}

	task_id := job.Steps[job.CurrentStep].TaskID
	task, err := this.task_repository.GetTask(task_id)
	if err != nil {
		return err
	} else if task == nil {
		return fmt.Errorf(
			"A task with id (%d) doesn't exist job:%d, step:%d",
			task_id, job.ID, job.CurrentStep)
	}

	// Check if the current module is responsible for the current job step
	if task.ModuleID != module.ID {
		// As far as this module is concerned, the job doesn't exist
		return ErrNotFound
	} else if step_id > job.Steps[job.CurrentStep].ID {
		return ErrNotFound
	} else if step_id < job.Steps[job.CurrentStep].ID {
		return ErrJobIsCompleted
	}

	// Set and save the finished state of the job
	job.IsCompleted = true
	job.IsCanceled = true
	job.Status = fmt.Sprintf("Canceled by service %s", module.Name)
	job.CompletionDate = new(time.Time)
	*job.CompletionDate = time.Now()

	err = this.repository.SaveFinishedState(job)
	if err != nil {
		return err
	}

	log.Infof("job=%v step=%v saved canceled state", job.ID, step_id)

	return this.SendCancelRequest(job, task, module)
}

// Cancel the job as the content owner that created it
func (this *jservice) CancelJobAsOwner(owner_id, job_id int64) error {
	log.Infof("job=%v cancel as user=%v", job_id, owner_id)
	job, err := this.repository.GetJobByID(job_id)

	if err != nil {
		return err
	} else if job == nil || job.Owner.ID != owner_id {
		return ErrNotFound
	}

	if job.IsCanceled {
		return ErrJobIsCanceled
	} else if job.IsCompleted {
		return ErrJobIsCompleted
	}

	// Get the task for the current job step
	err = this.repository.GetJobSteps(job.ID, &job.Steps)

	if err != nil {
		return err
	}

	task_id := job.Steps[job.CurrentStep].TaskID
	task, err := this.task_repository.GetTask(task_id)
	if err != nil {
		return err
	} else if task == nil {
		return fmt.Errorf(
			"A task with id (%d) doesn't exist job:%d, step:%d",
			task_id, job.ID, job.CurrentStep)
	}

	// Set and save the finished state of the job
	job.IsCompleted = true
	job.IsCanceled = true
	job.Status = "Canceled"
	job.CompletionDate = new(time.Time)
	*job.CompletionDate = time.Now()

	err = this.repository.SaveFinishedState(job)
	if err != nil {
		return err
	}

	log.Infof("job=%d saved canceled state", job.ID)

	module, err := this.module_repository.GetModuleByID(task.ModuleID)

	if err != nil {
		return err
	} else if module == nil {
		return fmt.Errorf("Module id(%d) was not found even thought task %d claims it exists",
			task.ModuleID,
			task.ID)
	}

	return this.SendCancelRequest(job, task, module)
}

func (this jservice) CancelJobsWithExceedingExpirationDate() error {
	now := time.Now()

	BATCH_LIMIT := int64(10000)

	// Get jobs in batches of <BATCH_LIMIT>
	jobs, err := this.repository.GetJobsExceedingExpirationDate(now, BATCH_LIMIT, 0)

	if err != nil {
		return err
	}

	for len(jobs) > 0 {
		log.Printf("Batch of %d jobs", len(jobs))
		for _, job := range jobs {
			log.Printf("Canceling job=%d at current_step=%d", job.ID, job.CurrentStep)

			err = this.repository.GetJobSteps(job.ID, &job.Steps)

			if err != nil {
				log.Printf("Failed to get steps for job=%v err='%v'", job.ID, err)
				continue
			}

			step := job.Steps[job.CurrentStep]

			task, err := this.task_repository.GetTask(step.TaskID)

			if err != nil || task == nil {
				log.Printf("Failed to fetch task=%d for step=%d err='%s'", step.TaskID, step.ID, err)
				continue
			}

			module, err := this.module_repository.GetModuleByID(task.ModuleID)

			if err != nil || module == nil {
				log.Printf("Failed to fetch module=%d for task=%d err='%s'", task.ModuleID, task.ID, err)
				continue
			}

			err = this.SendCancelRequest(job, task, module)

			if err != nil {
				log.Printf("Failed to send cancel request for job=%d err=%s", job.ID, err)
			}
		}

		jobs, err = this.repository.GetJobsExceedingExpirationDate(
			now,
			BATCH_LIMIT,
			jobs[len(jobs)-1].ID)

		if err != nil {
			return err
		}
	}

	return this.repository.CancelJobsWithExceedingExpirationDate(now)
}

// Cancels all the jobs that have exceeded the publication date
func (this jservice) CancelJobsWithExceedingPublicationDate() error {
	now := time.Now()

	BATCH_LIMIT := int64(10000)

	// Get jobs in batches of <BATCH_LIMIT>
	jobs, err := this.repository.GetJobsExceedingPublicationDate(now, BATCH_LIMIT, 0)

	if err != nil {
		return err
	}

	for len(jobs) > 0 {
		log.Printf("Batch of %d jobs", len(jobs))
		for _, job := range jobs {
			log.Printf("Canceling job=%d at current_step=%d", job.ID, job.CurrentStep)

			err = this.repository.GetJobSteps(job.ID, &job.Steps)

			if err != nil {
				log.Printf("Failed to get steps for job=%v err='%v'", job.ID, err)
				continue
			}

			step := job.Steps[job.CurrentStep]

			task, err := this.task_repository.GetTask(step.TaskID)

			if err != nil || task == nil {
				log.Printf("Failed to fetch task=%d for step=%d err='%s'", step.TaskID, step.ID, err)
				continue
			}

			module, err := this.module_repository.GetModuleByID(task.ModuleID)

			if err != nil || module == nil {
				log.Printf("Failed to fetch module=%d for task=%d err='%s'", task.ModuleID, task.ID, err)
				continue
			}

			err = this.SendCancelRequest(job, task, module)

			if err != nil {
				log.Printf("Failed to send cancel request for job=%d err=%s", job.ID, err)
			}
		}

		jobs, err = this.repository.GetJobsExceedingPublicationDate(
			now,
			BATCH_LIMIT,
			jobs[len(jobs)-1].ID)

		if err != nil {
			return err
		}
	}

	return this.repository.CancelJobsWithExceedingPublicatinDate(now)
}

// Creates a new Job for this user
// user_id: the id of the content owner
// publication_date: the latest time that the job should be completed
// expiration_date: the time that the job assets and parameters will be deleted
// tasks: a list of task as a map that contains
//		"id": the id of the task
//		"input": the input of the task in the form of "name":"value"
//		"linked_input": the linked input of the task in the form of "name":"previous_output_name"
func (this jservice) CreateJob(user_id, publication_date, expiration_date int64,
	tasks []map[string]interface{}) (*Job, error) {
	log.Infof("create new job user_id=%v publication_date=%v expiration_date=%v tasks='%v'",
		user_id, publication_date, expiration_date, tasks)

	if publication_date <= time.Now().Unix()+60 {
		return nil, ErrInvalidPublicationDate
	} else if expiration_date < publication_date {
		return nil, ErrInvalidExpirationDate
	} else if len(tasks) == 0 {
		return nil, ErrEmptyTasks
	}

	job := Job{
		CreationDate:    time.Now(),
		IsCanceled:      false,
		IsCompleted:     false,
		Owner:           ContentOwner{ID: user_id},
		Status:          "Started",
		CurrentStep:     0,
		Steps:           make([]*JobStep, len(tasks)),
		PublicationDate: time.Unix(publication_date, 0),
		ExpirationDate:  time.Unix(expiration_date, 0),
	}

	var prev_task *Task

	for order, step_data := range tasks {
		step := JobStep{}

		// Fetch the task
		task_id, ok := step_data["task_id"].(float64)
		if !ok {
			return nil, ErrMissingTaskID
		}
		step.TaskID = int64(task_id)

		task, err := this.task_repository.GetTask(step.TaskID)
		if err != nil {
			return nil, err
		} else if task == nil {
			return nil, &ErrTaskNotFound{TaskID: int64(task_id)}
		} else if !task.Enabled {
			return nil, &ErrTaskIsDisabled{TaskID: int64(task_id)}
		}

		// Check if the service is enabled
		module, err := this.module_repository.GetModuleByID(task.ModuleID)
		if err != nil {
			return nil, err
		} else if module == nil {
			return nil, fmt.Errorf(
				"Service %d, doesn't exist even though task %d claims so",
				task.ModuleID, task.ID)
		} else if !module.Enabled {
			return nil, &ErrModuleIsDisabled{ModuleID: module.ID}
		}

		input, _ := step_data["input"].(map[string]interface{})
		linked_input, _ := step_data["linked_input"].(map[string]interface{})

		// The correct number of input parameters must be given
		// For the starting task there should be no linked_parameters
		if (order == 0 && (len(input) != len(task.Input) || len(linked_input) != 0)) ||
			(order > 0 && len(input)+len(linked_input) != len(task.Input)) {
			return nil, &ErrInvalidTaskInput{Message: "Invalid number of input parameters"}
		}

		step.Input = make(map[string]JobParam)

		// Parse regular input
		for name, value := range input {
			give_type := GetParamType(value)
			if expected_type, ok := task.Input[name]; !ok {
				return nil, &ErrInvalidTaskInput{
					Message: fmt.Sprintf("Task with id=%d doesn't have a parameter named %v",
						task.ID, name)}
			} else if (expected_type == IntParam && give_type == StringParam) ||
				(expected_type != IntParam && expected_type != give_type) {
				// For values that are supposed to be 'int', we also accept `float`
				// and convert them afterwards.
				return nil, &ErrInvalidTaskInput{
					Message: fmt.Sprintf("Parameter %s should be of type %s",
						name, ParamTypeStr(expected_type))}
			} else {
				if expected_type == IntParam {
					value = int64(value.(float64))
				}
				step.Input[name] = JobParam{
					DataType: expected_type,
					Value:    value,
				}
			}
		}

		// Parse linked input
		for name, value := range linked_input {
			if output_name, ok := value.(string); !ok {
				return nil, &ErrInvalidTaskInput{
					Message: "Linked parameter name should be a string",
				}
			} else if in_type, ok := task.Input[name]; !ok {
				return nil, &ErrInvalidTaskInput{
					Message: fmt.Sprintf(
						"Task (%d) doesn't expect a parameter named %s",
						task.ID, name),
				}
			} else if out_type, ok := prev_task.Output[output_name]; !ok {
				return nil, &ErrTaskOutputNotFound{Name: output_name}
			} else if in_type != out_type {
				return nil, &ErrLinkedParameterNotTheSameType{
					InputName:  name,
					OutputName: output_name,
				}
			} else {
				step.Input[name] = JobParam{
					DataType:         in_type,
					LinkedOutputName: &output_name,
				}
			}
		}

		// Extra check, in case I missed a code path where this can happen
		if len(step.Input) != len(task.Input) {
			return nil, &ErrInvalidTaskInput{Message: "Invalid number of input parameters"}
		}

		prev_task = task
		job.Steps[order] = &step
	}

	// Validation passed, store job in db
	if err := this.repository.CreateJob(&job); err != nil {
		return nil, err
	}

	log.Infof("job=%v created by user=%v", job.ID, user_id)

	// Fill content owner information
	// Is need for 'PerformNextStepOfJob'
	if err := this.owner_repository.GetContentOwnerByID(&job.Owner); err != nil {
		return nil, err
	}

	go this.PerformNextStepOfJob(job)

	return &job, nil
}

// Cancels a job because of an error that has occured
func (this jservice) AbortJob(job *Job, reason string) {
	log.Warnf("job=%v abortd reason=%v", job.ID, reason)
	job.IsCompleted = true
	job.IsCanceled = true
	job.CompletionDate = new(time.Time)
	*job.CompletionDate = time.Now()
	job.Status = reason

	err := this.repository.SaveFinishedState(job)

	if err != nil {
		log.Errorf("job=%v abort failed err=%v", job.ID, err)
	}
}

// Perform the next step of the job starting from the Current job
// It is meant to be executed as a goroutine
func (this jservice) PerformNextStepOfJob(job Job) {
	log.Infof("job=%v perform next steps", job.ID)
	if job.IsCompleted || job.IsCanceled {
		log.Infof("job=%v is already completed or canceled", job.ID)
		return
	}

	if len(job.Steps) == 0 {
		log.Infof("job=%v get job steps", job.ID)
		err := this.repository.GetJobSteps(job.ID, &job.Steps)
		if err != nil {
			log.Errorf("job=%v failed to get steps err:%v", job.ID, err)
			return
		}
	}

	// Loop through ever step of the job starting from the current step
	for job.CurrentStep < len(job.Steps) {
		step := job.Steps[job.CurrentStep]
		log.Infof("job=%v current_step=%v step=%v", job.ID, job.CurrentStep, step.ID)

		if step.Input == nil {
			err := this.repository.GetParamsForStep(step)
			if err != nil {
				log.Errorf("job=%v failed to fetch step parameters step=%v  err=%v", job.ID, step.ID, err)
				this.AbortJob(&job, fmt.Sprintf("Failed to fetch parameters for step %d", job.CurrentStep))
				return
			}
		}

		// Get the task for this step
		task, err := this.task_repository.GetTask(step.TaskID)
		if err != nil || task == nil {
			log.Errorf("job=%v task=%v failed to get task err=%v",
				job.ID,
				step.TaskID,
				err)

			this.AbortJob(&job, fmt.Sprintf("Couldn't fetch information for task %d", step.TaskID))
			return
		}

		// Get the service for this step
		service, err := this.module_repository.GetModuleByID(task.ModuleID)
		if err != nil || service == nil {
			log.Errorf("job=%v failed to fetch service=%v error=%v",
				job.ID,
				task.ModuleID,
				err)

			this.AbortJob(&job, fmt.Sprintf("Couldn't fetch information for service %d", task.ModuleID))
			return
		}

		// Prepare the input for the request to be sent to the service
		input_json := make(map[string]interface{})

		for name, param := range step.Input {
			if param.LinkedOutputName == nil {
				input_json[name] = param.Value
			} else if job.CurrentStep == 0 {
				// A linked input for the first step is not allowed
				log.Errorf("job=%v step=%d first step can't have linked input", job.ID, step.ID)
				this.AbortJob(&job, "Internal Server Error")
				return
			} else {
				// Parameter is linked, the data should be retrieved from previous job's output
				previous_step := job.Steps[job.CurrentStep-1]

				if previous_step.Output == nil {
					err := this.repository.GetParamsForStep(step)
					if err != nil {
						log.Errorf("job=%v step=%v failed to fetch parameters err=%v", job.ID, previous_step.ID, err)
						this.AbortJob(&job, fmt.Sprintf("Failed to fetch output for step %d", job.CurrentStep-1))
						return
					}
				}

				output, ok := previous_step.Output[*param.LinkedOutputName]
				if !ok {
					log.Errorf("job=%v step=%v input=%v is linked with output=%v which doesn't exist",
						job.ID, step.ID, name, *param.LinkedOutputName)
					this.AbortJob(&job, "Internal Server Error")
					return
				}
				input_json[name] = output.Value
			}
		}

		// Create the json string for the request
		json_data, _ := json.Marshal(map[string]interface{}{
			"job_id":           step.ID,
			"publication_date": job.PublicationDate.Unix(),
			"expiration_date":  job.ExpirationDate.Unix(),
			"content_owner":    job.Owner.Name,
			"input":            input_json,
		})

		// Send request and checking for errors
		log.Infof("job=%v step=%v send start request url=%s input=%v",
			job.ID, step.ID, task.StartUrl, string(json_data))

		client := http.Client{}
		req, err := http.NewRequest("POST", task.StartUrl, bytes.NewBuffer(json_data))

		if err != nil {
			log.Errorf("job=%v failed to prepare request err=%v", job.ID, err)
			this.AbortJob(&job, "Internal Server Error")
			return
		}

		req.Header.Add(EasyTVApiKeyHeader, service.ApiKey)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)

		if err != nil {
			log.Errorf("job=%v failed to send request err=%v", job.ID, err)
			this.AbortJob(&job, fmt.Sprintf("Task \"%v\" was unreachable", task.Name))
			return
		} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			log.Errorf("job=%v request failed with code=%v", job.ID, resp.StatusCode)
			this.AbortJob(&job, fmt.Sprintf("Task \"%v\" was unreachable", task.Name))
			return
		}

		// Parse http resposne
		json_data, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			log.Errorf("job=%v failed to read response body err=%v", job.ID, err)
			this.AbortJob(&job, "Internal Server Error")
			return
		}

		var data map[string]interface{}

		if err = json.Unmarshal(json_data, &data); err != nil {
			log.Printf("job=%d failed to parse response json err=%s", job.ID, err)
			this.AbortJob(&job, "Internal Server Error")
			return
		}

		codef, ok := data["code"].(float64)
		if !ok {
			log.Errorf("job=%v cant find \"code\" in the response", job.ID)
			this.AbortJob(&job, fmt.Sprintf("Task %d sent a malformed response", task.ID))
			return
		}
		code := int(codef)

		description, ok := data["description"].(string)
		if !ok {
			log.Errorf("job=%v Cant find \"description\" in the response", job.ID)
			this.AbortJob(&job, fmt.Sprintf("Task %d sent a malformed response", task.ID))
			return
		}

		switch code {
		case 200:
			output, ok := data["output"].(map[string]interface{})

			if !ok {
				log.Errorf("job=%d step=%d there is no output", job.ID, step.ID)
				this.AbortJob(&job, fmt.Sprintf("Task \"%v\" sent a malformed response", task.ID))
				return
			}

			log.Printf("job=%d step=%d completed synchronously with output=%v",
				job.ID, step.ID, output)

			if step.Output == nil {
				step.Output = make(map[string]JobParam)
			}

			for name, value := range output {
				given_type := GetParamType(value)
				correct_type, ok := task.Output[name]

				if !ok {
					log.Errorf("job=%d step=%v module sent unregistered output=%s",
						job.ID, step.ID, name)
				} else if (correct_type == IntParam && given_type == StringParam) ||
					(correct_type != IntParam && correct_type != given_type) {
					log.Errorf("job=%d step=%v module sent wrong type=%v for output=%v output_type=%v)",
						job.ID,
						step.ID,
						ParamTypeStr(given_type),
						name,
						ParamTypeStr(correct_type))
				} else if correct_type == IntParam {
					step.Output[name] = JobParam{
						DataType: correct_type,
						Value:    int64(value.(float64)),
					}
				} else {
					step.Output[name] = JobParam{
						DataType: correct_type,
						Value:    value,
					}
				}
			}
			log.Infof("job=%v step=%v current_step=%v completed", job.ID, step.ID, job.CurrentStep)
		case 202:
			// Task will be completed synchronously
			log.Infof("job=%d step=%v pending, it will be completed asynchronously", job.ID, step.ID)
			job.Status = fmt.Sprintf("Pending at task \"%s\" %d/%d",
				task.Name,
				job.CurrentStep,
				len(job.Steps))
			err = this.repository.SaveStatus(&job)
			if err != nil {
				log.Errorf("job=%d failed to save status err=%v", job.ID, err)
			}
			// Stop doing any more steps.
			// It will resume the job when the service sends a `finish` request
			return
		default:
			// Any other code results in an error
			log.Warnf("job=%v task error code=%v description=%v", job.ID, code, description)
			this.AbortJob(&job, fmt.Sprintf("Failed at task \"%v\" with code(%v)", task.Name, code))
			return
		}

		job.CurrentStep++
		err = this.repository.SaveStepProgress(&job)

		if err != nil {
			log.Errorf("job=%v failed to save step progress err=%v", job.ID, err)
			this.AbortJob(&job, "Internal Server Error")
			return
		}
	}
	// If it got here all the steps of the Job have been completed.
	job.IsCompleted = true
	job.CompletionDate = new(time.Time)
	*job.CompletionDate = time.Now()
	job.Status = "Completed"

	err := this.repository.SaveFinishedState(&job)
	if err != nil {
		log.Errorf("job=%d failed to save finished job err=%s", job.ID, err)
	}
	log.Infof("job=%d has been compelted", job.ID)
}

func (this *jservice) FinishJobStep(step_id int64, module *Module, output map[string]interface{}) error {
	log.Infof("step=%v finishing from module=%v,'%v'",
		step_id, module.ID, module.Name)
	job, err := this.repository.GetJobByStepID(step_id)

	if err != nil {
		return err
	} else if job == nil {
		return ErrNotFound
	} else if job.IsCompleted {
		return ErrJobIsCompleted
	} else if job.IsCanceled {
		return ErrJobIsCanceled
	} else if len(output) == 0 {
		return &ErrInvalidTaskOutput{
			Message: "output is empty",
		}
	}

	err = this.repository.GetJobSteps(job.ID, &job.Steps)
	if err != nil {
		return err
	}

	step := job.Steps[job.CurrentStep]

	if step.ID != step_id {
		// as far as the service is concerned the job is completed
		return ErrJobIsCompleted
	}

	task, err := this.task_repository.GetTask(step.TaskID)
	if err != nil {
		return err
	} else if task == nil {
		return fmt.Errorf("Task %v doesn't exist even though step %d claims so",
			step.TaskID, step.ID)
	} else if task.ModuleID != module.ID {
		// as far as this service is concerned the job doesn't exist
		return ErrNotFound
	}

	if len(output) != len(task.Output) {
		return &ErrInvalidTaskOutput{
			Message: "Wrong number of output parameters",
		}
	}

	// Parse the output
	step.Output = make(map[string]JobParam)
	for name, value := range output {
		give_type := GetParamType(value)

		if out_type, ok := task.Output[name]; !ok {
			return &ErrInvalidTaskOutput{
				Message: fmt.Sprintf(
					"Task with id=%d doesn't expect a parameter named %s",
					task.ID, name),
			}
		} else if (out_type == IntParam && give_type == StringParam) ||
			(out_type != IntParam && out_type != give_type) {
			// For values that are supposed to be 'int', we also accept `float`
			// and convert them afterwards.
			return &ErrInvalidTaskOutput{
				Message: fmt.Sprintf("Parameter %s should be a %s",
					name, ParamTypeStr(out_type)),
			}
		} else {
			if out_type == IntParam {
				step.Output[name] = JobParam{
					DataType: out_type,
					Value:    int64(value.(float64)),
				}
			} else {
				step.Output[name] = JobParam{
					DataType: out_type,
					Value:    value,
				}
			}
		}
	}

	// Extra check, in case I missed a code path where this can happen
	if len(step.Output) != len(task.Output) {
		return &ErrInvalidTaskOutput{
			Message: "Wrong number of output parameters",
		}
	}

	log.Infof("job=%v step=%v finished with output=%v",
		job.ID, step_id, output)

	job.CurrentStep++

	if err = this.repository.SaveStepProgress(job); err != nil {
		return err
	}

	if err = this.owner_repository.GetContentOwnerByID(&job.Owner); err != nil {
		return err
	}

	go this.PerformNextStepOfJob(*job)

	return nil
}
