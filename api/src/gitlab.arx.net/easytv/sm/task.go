package sm

import "errors"

// ParamType is the type of an input or output parameter of a task
type ParamType int

const (
	IntParam             ParamType = 0
	StringParam          ParamType = 1
	DoubleParam          ParamType = 2
	UnsupportedTypeParam ParamType = -1
)

// ParamTypeStr returns a string name of the given ParamType
func ParamTypeStr(param_type ParamType) string {
	switch param_type {
	case IntParam:
		return "int"
	case StringParam:
		return "string"
	case DoubleParam:
		return "double"
	}
	return ""
}

func GetParamType(value interface{}) ParamType {
	switch value.(type) {
	case string:
		return StringParam
	case int64:
		return IntParam
	case float64:
		return DoubleParam
	}
	return UnsupportedTypeParam
}

func GetParamTypeFromSting(value string) ParamType {
	switch value {
	case "string":
		return StringParam
	case "int":
		return IntParam
	case "double":
		return DoubleParam
	}
	return UnsupportedTypeParam
}

type Task struct {
	ID          int64
	ModuleID    int64
	Name        string
	Description string
	StartUrl    string
	CancelUrl   string
	Enabled     bool
	Deleted     bool
	Input       map[string]ParamType
	Output      map[string]ParamType
}

type TaskRepository interface {
	CreateTask(task *Task) error

	NameExists(name string) (bool, error)

	GetTask(id int64) (*Task, error)

	GetTasks(module_id int64, fetch_deleted bool) ([]*Task, error)

	SetAvailability(id int64, enabled bool) error

	DeleteTask(id int64) error

	Save(task *Task) error

	SaveVars(task *Task, is_input bool) error
}

// Errors

var ErrTaskNameTooShort = errors.New("Task name is short")
var ErrTaskAlreadyExists = errors.New("Task already exists")
var ErrTaskDescTooShort = errors.New("Task desc is short")
var ErrInvalidStartUrl = errors.New("invalid start url")
var ErrInvalidCancelUrl = errors.New("invalid cancel url")
var ErrInvalidInputType = errors.New("invalid input type")
var ErrInvalidOutputType = errors.New("invalid output type")
var ErrEmptyInput = errors.New("Input can't be empty")
var ErrEmptyOutput = errors.New("Output can't be empty")
var ErrTaskIsEnabled = errors.New("Task is enabled")
var ErrTaskHasJobStepsInProgress = errors.New("Task has job steps in progress")

// Service
type TaskService interface {
	RegisterTask(
		module_id int64,
		name, description, start_url, cancel_url string,
		input, output map[string]ParamType) (*Task, error)

	SetAvailability(id int64, enabled bool) error

	DeleteTask(id int64) error

	Update(id int64, fields map[string]string) error

	UpdateVars(id int64, data map[string]ParamType, is_input bool) error
}
