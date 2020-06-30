package db

import (
	"database/sql"

	"gitlab.arx.net/easytv/sm"
)

type TaskRepository struct {
	Pool *DatabasePool
}

func (this *TaskRepository) CreateTask(task *sm.Task) error {
	tx, err := this.Pool.DB.Begin()

	if err != nil {
		return nil
	}

	stmt, query_err := tx.Prepare(`
		insert into task (module_id,
			name, 
			description, 
			start_url,
			cancel_url,
			enabled,
			deleted)
		values ($1, $2, $3, $4, $5, false, false)
		returning id
	`)

	defer stmt.Close()

	if query_err != nil {
		return query_err
	}

	row := stmt.QueryRow(
		task.ModuleID,
		task.Name,
		task.Description,
		task.StartUrl,
		task.CancelUrl)

	err = row.Scan(&task.ID)

	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, query_err = tx.Prepare(`
		insert into task_parameter (task_id, name, data_type, is_input)
		values ($1, $2, $3, $4)
	`)

	if query_err != nil {
		tx.Rollback()
		return query_err
	}

	defer stmt.Close()

	for name, parameter_type := range task.Input {
		_, err = stmt.Exec(task.ID, name, parameter_type, true)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for name, parameter_type := range task.Output {
		_, err = stmt.Exec(task.ID, name, parameter_type, false)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (this *TaskRepository) SaveVars(task *sm.Task, is_input bool) error {
	tx, err := this.Pool.DB.Begin()

	if err != nil {
		return err
	}

	delete_stmt, err := tx.Prepare(`
		delete from task_parameter where task_id=$1 and is_input=$2
	`)

	if err != nil {
		tx.Rollback()
		return err
	}

	insert_stmt, err := tx.Prepare(`
		insert into task_parameter (task_id, name, data_type, is_input)
		values ($1, $2, $3, $4)
	`)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = delete_stmt.Exec(task.ID, is_input)

	if err != nil {
		tx.Rollback()
		return err
	}

	if is_input {
		for name, parameter_type := range task.Input {
			_, err = insert_stmt.Exec(task.ID, name, parameter_type, true)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	} else {
		for name, parameter_type := range task.Output {
			_, err = insert_stmt.Exec(task.ID, name, parameter_type, false)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	tx.Commit()
	return nil
}

func (this *TaskRepository) NameExists(name string) (bool, error) {
	stmt, err := this.Pool.Prepare(`
		select id from task
		where name=$1
		limit 1
	`)

	if err != nil {
		return false, err
	}

	row := stmt.QueryRow(name)

	var id int64
	err = row.Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (this *TaskRepository) GetTask(id int64) (*sm.Task, error) {
	stmt, err := this.Pool.Prepare(`
		select module_id, name, description, start_url, cancel_url, enabled, deleted
		from task
		where id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(id)

	task := sm.Task{ID: id}

	err = row.Scan(&task.ModuleID,
		&task.Name,
		&task.Description,
		&task.StartUrl,
		&task.CancelUrl,
		&task.Enabled,
		&task.Deleted)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var param_stmt *sql.Stmt
	param_stmt, err = this.Pool.Prepare(`
		select name, data_type, is_input
		from task_parameter
		where task_id=$1
	`)

	if err != nil {
		return nil, err
	}

	param_rows, param_err := param_stmt.Query(task.ID)

	if param_err != nil {
		return nil, param_err
	}

	defer param_rows.Close()

	task.Input = make(map[string]sm.ParamType)
	task.Output = make(map[string]sm.ParamType)

	for param_rows.Next() {
		var name string
		var parameter_type sm.ParamType
		var is_input bool
		err = param_rows.Scan(&name, &parameter_type, &is_input)
		if err != nil {
			return nil, err
		}

		if is_input {
			task.Input[name] = parameter_type
		} else {
			task.Output[name] = parameter_type
		}
	}

	return &task, nil
}

func (this *TaskRepository) GetTasks(module_id int64, fetch_deleted bool) ([]*sm.Task, error) {
	var select_task_query string
	if fetch_deleted {
		select_task_query = `
							select id, name, description, start_url, cancel_url, enabled, deleted
							from task
							where module_id=$1
							`
	} else {
		select_task_query = `
							select id, name, description, start_url, cancel_url, enabled, deleted
							from task
							where deleted=false and module_id=$1
							`
	}

	stmt, err := this.Pool.Prepare(select_task_query)

	if err != nil {
		return nil, err
	}

	param_stmt, err := this.Pool.Prepare(`
							select name, data_type, is_input
							from task_parameter
							where task_id=$1
						`)

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(module_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := make([]*sm.Task, 0)

	for rows.Next() {
		task := sm.Task{ModuleID: module_id}

		err = rows.Scan(&task.ID,
			&task.Name,
			&task.Description,
			&task.StartUrl,
			&task.CancelUrl,
			&task.Enabled,
			&task.Deleted)

		if err != nil {
			return nil, err
		}

		param_rows, err := param_stmt.Query(task.ID)

		if err != nil {
			return nil, err
		}

		defer param_rows.Close()

		task.Input = make(map[string]sm.ParamType)
		task.Output = make(map[string]sm.ParamType)

		for param_rows.Next() {
			var name string
			var parameter_type sm.ParamType
			var is_input bool
			err = param_rows.Scan(&name, &parameter_type, &is_input)
			if err != nil {
				return nil, err
			}

			if is_input {
				task.Input[name] = parameter_type
			} else {
				task.Output[name] = parameter_type
			}
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (this *TaskRepository) SetAvailability(id int64, enabled bool) error {
	stmt, err := this.Pool.Prepare(`
		update task
		set enabled=$1
		where id=$2
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(enabled, id)

	return err
}

func (this *TaskRepository) Save(task *sm.Task) error {
	stmt, err := this.Pool.Prepare(`
		update task set
			name=$1, description=$2, start_url=$3, cancel_url=$4
		where id=$5
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(task.Name,
		task.Description,
		task.StartUrl,
		task.CancelUrl,
		task.ID)

	if err != nil {
		return err
	} else if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (this *TaskRepository) DeleteTask(id int64) error {
	stmt, err := this.Pool.Prepare(`
		update task
		set deleted=true, name= id || '_DELETED_' || name
		where id=$1
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)

	return err
}
