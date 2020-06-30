package db

import (
	"database/sql"
	"fmt"

	"gitlab.arx.net/easytv/sm"
)

type ContentOwnerRepository struct {
	Pool *DatabasePool
}

func (this *ContentOwnerRepository) GetContentOwnerByID(owner *sm.ContentOwner) error {
	stmt, err := this.Pool.Prepare(`
		select username, email, name, password
		from content_owner
		where id=$1
	`)

	if err != nil {
		return err
	}

	row := stmt.QueryRow(owner.ID)

	err = row.Scan(
		&owner.Username,
		&owner.Email,
		&owner.Name,
		&owner.Password)

	return err
}

func (this *ContentOwnerRepository) GetAll() ([]*sm.ContentOwner, error) {
	rows, err := this.Pool.DB.Query(`
		select id, username, email, name, password
		from content_owner
	`)

	if err != nil {
		return nil, err
	}

	users := make([]*sm.ContentOwner, 0)

	for rows.Next() {
		user := sm.ContentOwner{}

		err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Name,
			&user.Password)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

func (this *ContentOwnerRepository) GetContentOwnerByUsername(
	username string) (*sm.ContentOwner, error) {
	stmt, err := this.Pool.Prepare(`
		select id, email, name, password
		from content_owner
		where username=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(username)

	owner := sm.ContentOwner{Username: username}

	err = row.Scan(
		&owner.ID,
		&owner.Email,
		&owner.Name,
		&owner.Password)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &owner, nil
}

// NEVER give 'col' a value from user input
func (this *ContentOwnerRepository) exists(col, value string) (bool, error) {
	stmt, err := this.Pool.Prepare(fmt.Sprintf(`
		select id
		from content_owner
		where %s=$1
	`, col))

	if err != nil {
		return false, err
	}

	row := stmt.QueryRow(value)

	var id int64
	err = row.Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (this *ContentOwnerRepository) NameExists(name string) (bool, error) {
	return this.exists("name", name)
}

func (this *ContentOwnerRepository) UsernameExists(username string) (bool, error) {
	return this.exists("username", username)
}

func (this *ContentOwnerRepository) EmailExists(email string) (bool, error) {
	return this.exists("email", email)
}

func (this *ContentOwnerRepository) Insert(owner *sm.ContentOwner) error {
	stmt, err := this.Pool.Prepare(`
		insert into content_owner (username, name, password, email)
		values ($1, $2, $3, $4)
		returning id
	`)

	if err != nil {
		return err
	}

	row := stmt.QueryRow(owner.Username,
		owner.Name,
		owner.Password,
		owner.Email)

	err = row.Scan(&owner.ID)

	return err
}

func (this *ContentOwnerRepository) SavePassword(owner *sm.ContentOwner) error {
	stmt, err := this.Pool.Prepare(`
		update content_owner set
		password=$2
		where id=$1
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(owner.ID, owner.Password)

	return err
}

func (this *ContentOwnerRepository) Save(owner *sm.ContentOwner) error {
	stmt, err := this.Pool.Prepare(`
		update content_owner set
		name=$2, username=$3, email=$4
		where id=$1
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(owner.ID, owner.Name, owner.Username, owner.Email)

	return err
}
