package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	"gitlab.arx.net/easytv/sm"
)

type ModuleRepository struct {
	Pool *DatabasePool
}

func (this *ModuleRepository) CreateModule(module *sm.Module) error {
	stmt, err := this.Pool.Prepare(`insert into module
	(name, description, api_key, enabled)
	values ($1, $2, $3, $4) returning id`)

	if err != nil {
		return err
	}

	row := stmt.QueryRow(module.Name,
		module.Description,
		module.ApiKey,
		module.Enabled)

	return row.Scan(&module.ID)
}

func (this *ModuleRepository) IsNameInUse(name string) (bool, error) {
	stmt, err := this.Pool.Prepare(`
	select id
	from module
	where name=$1
	limit 1`)

	if err != nil {
		return false, nil
	}

	row := stmt.QueryRow(name)

	var id int
	err = row.Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func (this *ModuleRepository) GenerateApiKey(module_name string) string {
	var random_bytes [20]byte
	_, _ = rand.Read(random_bytes[:])

	module_name += string(random_bytes[:])

	hash := sha256.Sum256([]byte(module_name))

	key := make([]byte, len(hash)*2)
	hex.Encode([]byte(key), hash[:])
	return string(key)
}

func (this *ModuleRepository) SetAvailability(id int64, enabled bool) error {

	stmt, err := this.Pool.Prepare(`
	update module
	set enabled=$1
	where id=$2`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(enabled, id)

	if err != nil {
		return err
	} else if row, _ := res.RowsAffected(); row == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (this *ModuleRepository) Save(module *sm.Module) error {
	stmt, err := this.Pool.Prepare(`
		update module
		set name=$1, description=$2, api_key=$4
		where id=$3
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(module.Name,
		module.Description,
		module.ID,
		module.ApiKey)

	if err != nil {
		return err
	} else if row, _ := res.RowsAffected(); row == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (this *ModuleRepository) GetModulesOBJ(modules *[]*sm.Module) error {

	rows, err := this.Pool.DB.Query(`
		select id, api_key, name, description, enabled
		from module`)

	if err != nil {
		return err
	}

	for rows.Next() {
		module := sm.Module{}

		err := rows.Scan(
			&module.ID,
			&module.ApiKey,
			&module.Name,
			&module.Description,
			&module.Enabled)

		if err != nil {
			return err
		}

		*modules = append(*modules, &module)
	}

	return nil
}

func (this *ModuleRepository) GetModules(modules *[]map[string]interface{}) error {
	rows, err := this.Pool.DB.Query(`
	select id, api_key, name, description, enabled
	from module`)

	if err != nil {
		return err
	}

	defer rows.Close()

	module := sm.Module{}
	for rows.Next() {

		err := rows.Scan(
			&module.ID,
			&module.ApiKey,
			&module.Name,
			&module.Description,
			&module.Enabled)

		if err != nil {
			return err
		}

		*modules = append(*modules, map[string]interface{}{
			"id":          module.ID,
			"api_key":     module.ApiKey,
			"name":        module.Name,
			"description": module.Description,
			"enabled":     module.Enabled})
	}

	return nil
}

func (this *ModuleRepository) GetModuleByID(id int64) (*sm.Module, error) {
	stmt, err := this.Pool.Prepare(`
		select api_key, name, description, enabled
		from module
		where id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(id)

	module := sm.Module{ID: id}

	err = row.Scan(
		&module.ApiKey,
		&module.Name,
		&module.Description,
		&module.Enabled)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &module, err
}

func (this *ModuleRepository) GetModuleByKey(api_key string) (*sm.Module, error) {
	stmt, err := this.Pool.Prepare(`
		select id, name, description, enabled
		from module
		where api_key=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(api_key)

	module := sm.Module{ApiKey: api_key}

	err = row.Scan(
		&module.ID,
		&module.Name,
		&module.Description,
		&module.Enabled)

	return &module, err
}
