package sm

import (
	"errors"
)

type Module struct {
	ID          int64
	Name        string
	Description string
	ApiKey      string
	Enabled     bool
}

type ModuleRepository interface {
	CreateModule(module *Module) error

	IsNameInUse(name string) (bool, error)

	SetAvailability(id int64, enabled bool) error

	GetModules(modules *[]map[string]interface{}) error

	GetModulesOBJ(modules *[]*Module) error

	GetModuleByKey(api_key string) (*Module, error)

	GetModuleByID(id int64) (*Module, error)

	GenerateApiKey(module_name string) string

	Save(module *Module) error
}

// errors

var ErrServiceNameTooShort = errors.New("Name is too short")
var ErrServiceDescTooShort = errors.New("Desc is too short")
var ErrServiceNameInUse = errors.New("Name is in use")

// services

type ModuleService interface {
	CreateService(name, description string) (*Module, error)

	UpdateName(id int64, name string) error

	UpdateDescription(id int64, description string) error

	Update(id int64, name string, description string) error

	SetAvailability(id int64, enable bool) error

	RenewApiKey(id int64) (*Module, error)
}
