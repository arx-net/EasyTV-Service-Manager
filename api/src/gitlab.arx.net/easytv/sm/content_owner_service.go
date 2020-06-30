package sm

import (
	log "github.com/sirupsen/logrus"

	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type coservice struct {
	repository ContentOwnerRepository
	rand_src   rand.Source
}

func NewContentOwnerService(repo ContentOwnerRepository) ContentOwnerService {
	return &coservice{
		repository: repo,
		rand_src:   rand.NewSource(time.Now().UnixNano()),
	}
}

func (this *coservice) GenerateRandomPassword(n int) string {
	b := make([]byte, n)
	// A this.rand_src.Int63() generates 63 random bits,
	// enough for letterIdxMax characters!
	for i, cache, remain := n-1, this.rand_src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = this.rand_src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func (this *coservice) CreateContentOwner(
	name, username, email, password string) (*ContentOwner, error) {

	if len(name) < 2 {
		return nil, ErrOwnerNameTooShort
	} else if len(username) < 2 {
		return nil, ErrOwnerUsernameTooShort
	} else if len(email) < 3 {
		return nil, ErrOwnerEmailTooShort
	}

	if exists, err := this.repository.NameExists(name); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrOwnerNameExists
	}

	if exists, err := this.repository.EmailExists(email); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrOwnerEmailExists
	}

	if exists, err := this.repository.UsernameExists(username); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrOwnerUsernameExists
	}

	hashed_password, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	owner := ContentOwner{
		Name:     name,
		Username: username,
		Email:    email,
		Password: string(hashed_password),
	}

	log.Infof("creating content owner username=%v name=%v email=%v", username, name, email)

	err = this.repository.Insert(&owner)
	if err != nil {
		return nil, err
	}
	log.Infof("created content owner username=%v with id=%v", username, owner.ID)
	// The password is returned in plain-text only in the creation
	owner.Password = password
	return &owner, nil
}

func (this *coservice) Login(username, password string) (*ContentOwner, error) {
	if len(username) == 0 || len(password) == 0 {
		return nil, nil
	}
	user, err := this.repository.GetContentOwnerByUsername(username)
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

func (this *coservice) ChangePassword(
	user_id int64, old_password, new_password string) error {
	if len(old_password) == 0 {
		return ErrInvalidCredentials
	} else if len(new_password) < 8 {
		return ErrPasswordTooShort
	}

	user := ContentOwner{ID: user_id}
	err := this.repository.GetContentOwnerByID(&user)

	if err != nil {
		return err
	} else if len(user.Password) == 0 {
		return ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(old_password)) != nil {
		return ErrInvalidCredentials
	}

	new_hash, err := bcrypt.GenerateFromPassword(
		[]byte(new_password), bcrypt.DefaultCost)

	user.Password = string(new_hash)

	log.Infof("Changing password for user=%v username=%v", user.ID, user.Name)

	return this.repository.SavePassword(&user)
}

func (this *coservice) ResetPassword(user_id int64, new_password string) error {
	if len(new_password) < 8 {
		return ErrPasswordTooShort
	}

	user := ContentOwner{ID: user_id}
	err := this.repository.GetContentOwnerByID(&user)

	if err != nil {
		return err
	} else if len(user.Password) == 0 {
		return ErrInvalidCredentials
	}

	new_hash, err := bcrypt.GenerateFromPassword(
		[]byte(new_password), bcrypt.DefaultCost)

	user.Password = string(new_hash)

	log.Infof("Changing password for user=%v username=%v", user.ID, user.Name)

	return this.repository.SavePassword(&user)
}

func (this *coservice) Update(user_id int64, fields map[string]string) error {
	user := ContentOwner{ID: user_id}

	err := this.repository.GetContentOwnerByID(&user)

	if err != nil {
		return err
	} else if len(user.Name) == 0 {
		return ErrNotFound
	}

	for name, value := range fields {
		if name == "Name" {
			if len(value) <= 2 {
				return ErrOwnerNameExists
			}

			exists, err := this.repository.NameExists(value)

			if err != nil {
				return err
			} else if exists {
				return ErrOwnerNameExists
			}

			user.Name = value
		} else if name == "Username" {
			if len(value) <= 2 {
				return ErrOwnerUsernameTooShort
			}

			exists, err := this.repository.UsernameExists(value)

			if err != nil {
				return err
			} else if exists {
				return ErrOwnerUsernameExists
			}

			user.Username = value
		} else if name == "Email" {
			if len(value) <= 3 {
				return ErrOwnerEmailTooShort
			}

			exists, err := this.repository.EmailExists(value)

			if err != nil {
				return err
			} else if exists {
				return ErrOwnerEmailExists
			}

			user.Email = value
		}
	}

	return this.repository.Save(&user)
}
