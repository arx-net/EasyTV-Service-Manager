package db

import (
	"database/sql"

	"gitlab.arx.net/easytv/sm"
)

type AdminRepository struct {
	Pool *DatabasePool
}

func (this *AdminRepository) GetByID(id int64) (*sm.AdminUser, error) {
	stmt, err := this.Pool.Prepare(`
		select username, password
		from admin_user
		where id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(id)

	user := sm.AdminUser{ID: id}

	err = row.Scan(&user.Username, &user.Password)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (this *AdminRepository) GetByUsername(username string) (*sm.AdminUser, error) {
	stmt, err := this.Pool.Prepare(`
		select id, password
		from admin_user
		where username=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(username)

	user := sm.AdminUser{Username: username}

	err = row.Scan(&user.ID, &user.Password)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (this *AdminRepository) Insert(user *sm.AdminUser) error {
	stmt, err := this.Pool.Prepare(`
		insert into admin_user (username, password)
		values ($1, $2)
		returning id
	`)

	if err != nil {
		return err
	}

	row := stmt.QueryRow(user.Username, user.Password)

	return row.Scan(&user.ID)
}

func (this *AdminRepository) SavePassword(admin *sm.AdminUser) error {
	stmt, err := this.Pool.Prepare(`
		update admin_user set
		password=$2
		where id=$1
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(admin.ID, admin.Password)

	return err
}
