package sm

import (
	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"
)

type admservice struct {
	repository AdminRepository
}

func NewAdminService(repository AdminRepository) AdminService {
	return &admservice{
		repository: repository,
	}
}

func (this *admservice) Login(username, password string) (*AdminUser, error) {
	user, err := this.repository.GetByUsername(username)

	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password)) != nil {
		return nil, nil
	}

	return user, nil
}

func (this *admservice) CreateAdminUser(
	username, password string) (*AdminUser, error) {
	if len(username) < 1 {
		return nil, ErrAdminNameTooShort
	} else if len(password) < 5 {
		return nil, ErrPasswordTooShort
	}

	hashed_password, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	log.Infof("Creating admin username=%s", username)

	user := AdminUser{
		Username: username,
		Password: string(hashed_password),
	}

	err = this.repository.Insert(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (this *admservice) ChangePassword(
	admin_id int64, old_password, new_password string) error {
	log.Infof("change from %v", old_password)
	if len(old_password) == 0 {
		return ErrInvalidCredentials
	} else if len(new_password) < 8 {
		return ErrPasswordTooShort
	}

	user, err := this.repository.GetByID(admin_id)

	if err != nil {
		return err
	} else if user == nil {
		return ErrInvalidCredentials
	}

	log.Infof("Compare password %v %v", old_password, new_password)

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(old_password)) != nil {
		return ErrInvalidCredentials
	}

	new_hash, err := bcrypt.GenerateFromPassword(
		[]byte(new_password), bcrypt.DefaultCost)

	user.Password = string(new_hash)

	log.Infof("change admin password username=%v", user.Username)

	return this.repository.SavePassword(user)
}
