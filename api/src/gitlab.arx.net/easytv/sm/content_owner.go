package sm

import (
	"errors"
)

type ContentOwner struct {
	ID       int64
	Username string
	Password string
	Email    string
	Name     string
}

type ContentOwnerRepository interface {
	NameExists(name string) (bool, error)

	UsernameExists(username string) (bool, error)

	EmailExists(email string) (bool, error)

	GetContentOwnerByUsername(username string) (*ContentOwner, error)

	GetContentOwnerByID(owner *ContentOwner) error

	Insert(owner *ContentOwner) error

	SavePassword(owner *ContentOwner) error

	Save(owner *ContentOwner) error

	GetAll() ([]*ContentOwner, error)
}

// errors

var ErrOwnerNameTooShort = errors.New("Owner name too short")
var ErrOwnerEmailTooShort = errors.New("Owner email too short")
var ErrOwnerUsernameTooShort = errors.New("Owner username too short")
var ErrOwnerNameExists = errors.New("Owner name exists")
var ErrOwnerEmailExists = errors.New("Owner email exists")
var ErrOwnerUsernameExists = errors.New("Owner username exists")

// service
type ContentOwnerService interface {
	CreateContentOwner(name, username, email, password string) (*ContentOwner, error)

	GenerateRandomPassword(n int) string

	Login(username, password string) (*ContentOwner, error)

	ChangePassword(user_id int64, old_password, new_password string) error

	ResetPassword(user_id int64, new_password string) error

	Update(user_id int64, fields map[string]string) error
}
