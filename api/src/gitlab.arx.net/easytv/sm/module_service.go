package sm

import (
	"crypto/sha256"
	"encoding/hex"

	log "github.com/sirupsen/logrus"
)

type mservice struct {
	repository ModuleRepository
}

func NewModuleService(repo ModuleRepository) ModuleService {
	return &mservice{
		repository: repo,
	}
}

func (this *mservice) CreateService(name, description string) (*Module, error) {
	log.Infof("create service=%v", name)

	if len(name) <= 1 {
		return nil, ErrServiceNameTooShort
	} else if len(description) <= 1 {
		return nil, ErrServiceDescTooShort
	}

	exists, err := this.repository.IsNameInUse(name)

	if err != nil {
		return nil, err
	} else if exists {
		return nil, ErrServiceNameInUse
	}

	api_key := this.repository.GenerateApiKey(name)

	hashed_api_key := sha256.Sum256([]byte(api_key))

	module := Module{
		Name:        name,
		Description: description,
		Enabled:     false,
		ApiKey:      hex.EncodeToString(hashed_api_key[:]),
	}

	err = this.repository.CreateModule(&module)

	if err != nil {
		return nil, err
	}

	// Return the key in plain text, after this it is lost.
	module.ApiKey = api_key

	log.Printf("created service name=%v id=%v", name, module.ID)

	return &module, nil
}

func (this *mservice) SetAvailability(id int64, enable bool) error {
	service, err := this.repository.GetModuleByID(id)
	if err != nil {
		return err
	} else if service == nil {
		return ErrNotFound
	}

	log.Printf("service=%v,'%v' set availability to %v",
		service.ID,
		service.Name,
		enable)

	return this.repository.SetAvailability(id, enable)
}

func (this *mservice) UpdateName(id int64, name string) error {
	service, err := this.repository.GetModuleByID(id)
	if err != nil {
		return err
	} else if service == nil {
		return ErrNotFound
	}

	log.Printf("service=%v,'%v' update name to '%v'",
		service.ID, service.Name, name)

	service.Name = name

	return this.repository.Save(service)
}

func (this *mservice) UpdateDescription(id int64, description string) error {
	service, err := this.repository.GetModuleByID(id)
	if err != nil {
		return err
	} else if service == nil {
		return ErrNotFound
	}

	log.Printf("service=%v,'%v' update description to '%v'",
		service.ID, service.Name, description)

	service.Description = description

	return this.repository.Save(service)
}

func (this *mservice) Update(id int64, name string, description string) error {
	service, err := this.repository.GetModuleByID(id)
	if err != nil {
		return err
	} else if service == nil {
		return ErrNotFound
	}

	log.Printf("service=%v,'%v' update name='%v' description='%v'",
		service.ID, service.Name, name, description)

	service.Description = description
	service.Name = name

	return this.repository.Save(service)
}

func (this *mservice) RenewApiKey(id int64) (*Module, error) {
	service, err := this.repository.GetModuleByID(id)

	if err != nil {
		return nil, err
	} else if service == nil {
		return nil, ErrNotFound
	}

	api_key := this.repository.GenerateApiKey(service.Name)

	hashed_key := sha256.Sum256([]byte(api_key))

	service.ApiKey = hex.EncodeToString(hashed_key[:])

	err = this.repository.Save(service)
	if err != nil {
		return nil, err
	}

	// return in plain text in order to notify the user of the new key
	// after that its lost
	service.ApiKey = api_key
	return service, nil
}
