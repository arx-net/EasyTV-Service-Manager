package sm

import "errors"

type AdminUser struct {
	ID       int64
	Username string
	Password string
}

type AdminRepository interface {
	GetByID(id int64) (*AdminUser, error)

	GetByUsername(username string) (*AdminUser, error)

	Insert(user *AdminUser) error

	SavePassword(admin *AdminUser) error
}

var ErrInvalidCredentials = errors.New("credentials were not valid")
var ErrAdminNameTooShort = errors.New("Admin name is too short")
var ErrPasswordTooShort = errors.New("Admin password too short")

type AdminService interface {
	Login(username, password string) (*AdminUser, error)

	CreateAdminUser(username, password string) (*AdminUser, error)

	ChangePassword(admin_id int64, old_password, new_password string) error
}
