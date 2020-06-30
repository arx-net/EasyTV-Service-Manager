package sm

import (
	"errors"
	"fmt"
	"time"
)

type JobParam struct {
	DataType ParamType
	Value    interface{}
	// can be linked with an output param of the previous step
	LinkedOutputName *string
}

type JobStep struct {
	ID     int64
	TaskID int64
	// The `key` of the map is the name of the parameter
	Input  map[string]JobParam
	Output map[string]JobParam
}

type Job struct {
	ID              int64
	IsCompleted     bool
	IsCanceled      bool
	CreationDate    time.Time
	CompletionDate  *time.Time
	PublicationDate time.Time
	ExpirationDate  time.Time
	Owner           ContentOwner
	Status          string
	CurrentStep     int
	Steps           []*JobStep
}

type JobRepository interface {
	TaskHasActiveJobs(task_id int64) (bool, error)

	GetJobsStepsForModule(
		module_id,
		limit,
		before_job_id int64) ([]map[string]interface{}, error)

	GetJobStepForModule(
		step_id, module_id int64) (map[string]interface{}, error)

	GetJobsForContentOwner(
		owner_id, limit, before_job_id int64) ([]*Job, error)

	GetJobByStepID(step_id int64) (*Job, error)

	GetJobByID(job_id int64) (*Job, error)

	GetJobSteps(job_id int64, steps *[]*JobStep) error

	GetParamsForStep(step *JobStep) error

	SaveStatus(job *Job) error

	CreateJob(job *Job) error

	SaveStepProgress(job *Job) error

	SaveFinishedState(job *Job) error

	GetJobsExceedingPublicationDate(
		timestamp time.Time,
		limit, offset_id int64) ([]*Job, error)

	CancelJobsWithExceedingPublicatinDate(timestamp time.Time) error

	GetJobsExceedingExpirationDate(
		timestamp time.Time,
		limit, offset_id int64) ([]*Job, error)

	CancelJobsWithExceedingExpirationDate(timestamp time.Time) error

	GetNewExpiredJobsAt(
		timestamp time.Time,
		limit int64,
		offset_id int64) ([]*Job, error)

	ExpireJobsBefore(timestamp time.Time) error

	DeleteParamsForJob(job_id int64) error
}

type JobService interface {
	SetJobStatusForStep(step_id int64, status string) error

	CreateJob(user_id, publication_date, expiration_date int64,
		tasks []map[string]interface{}) (*Job, error)

	CancelJobsWithExceedingPublicationDate() error

	CancelJobsWithExceedingExpirationDate() error

	CancelJobAsModule(module *Module, step_id int64) error

	CancelJobAsOwner(owner_id, job_id int64) error

	SendCancelRequest(job *Job, task *Task, module *Module) error

	// The job is passed by value in order to make the function
	// safe to call as a goroutine
	PerformNextStepOfJob(job Job)

	FinishJobStep(step_id int64, module *Module, output map[string]interface{}) error
}

// errors

var ErrJobIsCanceled = errors.New("Job is canceled")
var ErrJobIsCompleted = errors.New("Job is completed")
var ErrJobStatusNotUpdatable = errors.New(
	"Job status can't be updated because it is copmleted")

var ErrInvalidPublicationDate = errors.New("Publication date should be in the future")
var ErrInvalidExpirationDate = errors.New("Expiration date should be after publication date")
var ErrEmptyTasks = errors.New("There should be at least on task for a job")
var ErrMissingTaskID = errors.New("The task_id is missing in the input")

type ErrTaskNotFound struct{ TaskID int64 }

func (e *ErrTaskNotFound) Error() string {
	return fmt.Sprintf("A task with id(%d) does not exist", e.TaskID)
}

type ErrTaskIsDisabled struct{ TaskID int64 }

func (e *ErrTaskIsDisabled) Error() string {
	return fmt.Sprintf("Task %v is disabled", e.TaskID)
}

type ErrModuleIsDisabled struct{ ModuleID int64 }

func (e *ErrModuleIsDisabled) Error() string {
	return fmt.Sprintf("Module %v is disabled", e.ModuleID)
}

type ErrInvalidTaskInput struct{ Message string }

func (e *ErrInvalidTaskInput) Error() string {
	return e.Message
}

type ErrInvalidTaskOutput struct{ Message string }

func (e *ErrInvalidTaskOutput) Error() string {
	return e.Message
}

type ErrTaskOutputNotFound struct{ Name string }

func (e *ErrTaskOutputNotFound) Error() string {
	return fmt.Sprintf(
		"Previous task doesn't have an output named %s", e.Name)
}

type ErrLinkedParameterNotTheSameType struct {
	InputName  string
	OutputName string
}

func (e *ErrLinkedParameterNotTheSameType) Error() string {
	return fmt.Sprintf(
		"\"%s\" and \"%s\" are not of the same type",
		e.InputName, e.OutputName)
}
