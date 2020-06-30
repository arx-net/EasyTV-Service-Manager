package db

import (
	"database/sql"

	log "github.com/sirupsen/logrus"

	"strings"
	"time"

	"gitlab.arx.net/easytv/sm"
)

type JobRepository struct {
	Pool *DatabasePool
}

func (this *JobRepository) TaskHasActiveJobs(task_id int64) (bool, error) {
	stmt, err := this.Pool.Prepare(`
		select job_step.id from job
		inner join job_step
			on job_step.job_id = job.id
		where 
			job_step.task_id=$1 and 
			job_step.step_order>=job.current_step and 
			not job.is_canceled and 
			not job.is_completed
		limit 1
	`)

	if err != nil {
		return false, nil
	}

	row := stmt.QueryRow(task_id)

	var job_step_id int64
	err = row.Scan(&job_step_id)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (this *JobRepository) GetJobStepForModule(
	step_id, module_id int64) (map[string]interface{}, error) {

	stmt, err := this.Pool.Prepare(`
		select
			(j.is_completed or s.step_order < j.current_step) as is_completed,
			j.is_canceled,
			j.status,
			j.creation_date,
			j.completion_date,
			j.publication_date,
			j.expiration_date,
			o.name
		from job_step s
		inner join job j
			on j.id=s.job_id
		inner join content_owner o
			on o.id=j.owner_id
		inner join task t
			on t.id=s.task_id
		where 
			s.id=$1 and t.module_id=$2 and
			(j.current_step is null or s.step_order <= j.current_step)
		`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(step_id, module_id)

	var status, owner_name string
	var is_canceled, is_completed bool
	var creation_date, expiration_date, publication_date time.Time
	var completion_date *time.Time

	err = row.Scan(&is_completed,
		&is_canceled,
		&status,
		&creation_date,
		&completion_date,
		&publication_date,
		&expiration_date,
		&owner_name,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var completion_date_unix *int64
	if completion_date != nil {
		completion_date_unix = new(int64)
		*completion_date_unix = completion_date.Unix()
	}

	return map[string]interface{}{
		"id":               step_id,
		"is_completed":     is_completed,
		"is_canceled":      is_canceled,
		"status":           status,
		"creation_date":    creation_date.Unix(),
		"completion_date":  completion_date_unix,
		"publication_date": publication_date.Unix(),
		"expiration_date":  expiration_date.Unix(),
		"content_owner":    owner_name,
	}, nil
}

func (this *JobRepository) GetJobsStepsForModule(
	module_id,
	limit,
	before_step_id int64) ([]map[string]interface{}, error) {

	var query_builder strings.Builder

	query_builder.WriteString(`
		select
			s.id,
			(j.is_completed or s.step_order < j.current_step) as is_completed,
			j.is_canceled,
			j.status,
			j.creation_date,
			j.completion_date,
			j.publication_date,
			j.expiration_date,
			o.name
		from job_step s
		inner join job j
			on j.id=s.job_id
		inner join content_owner o
			on o.id=j.owner_id
		inner join task t
			on t.id=s.task_id
		where 
			t.module_id=$1 and
			(j.current_step is null or s.step_order <= j.current_step)
	`)

	if limit != -1 && before_step_id != -1 {
		query_builder.WriteString(`
			and s.id<$3
			order by s.id desc
			limit $2
		`)
	} else if limit != -1 {
		query_builder.WriteString(`
			order by s.id desc
			limit $2
		`)
	} else {
		query_builder.WriteString(`
			order by s.id desc
		`)
	}

	stmt, err := this.Pool.Prepare(query_builder.String())

	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if limit != -1 && before_step_id != -1 {
		rows, err = stmt.Query(module_id, limit, before_step_id)
	} else if limit != -1 {
		rows, err = stmt.Query(module_id, limit)
	} else {
		rows, err = stmt.Query(module_id)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]map[string]interface{}, 0)

	for rows.Next() {
		var step_id int64
		var status, owner_name string
		var is_canceled, is_completed bool
		var creation_date, expiration_date, publication_date time.Time
		var completion_date *time.Time

		err = rows.Scan(&step_id,
			&is_completed,
			&is_canceled,
			&status,
			&creation_date,
			&completion_date,
			&publication_date,
			&expiration_date,
			&owner_name,
		)

		if err != nil {
			return nil, err
		}

		var completion_date_unix *int64
		if completion_date != nil {
			completion_date_unix = new(int64)
			*completion_date_unix = completion_date.Unix()
		}

		jobs = append(jobs, map[string]interface{}{
			"id":               step_id,
			"is_completed":     is_completed,
			"is_canceled":      is_canceled,
			"status":           status,
			"creation_date":    creation_date.Unix(),
			"completion_date":  completion_date_unix,
			"publication_date": publication_date.Unix(),
			"expiration_date":  expiration_date.Unix(),
			"content_owner":    owner_name,
		})
	}

	return jobs, nil
}

func (this *JobRepository) GetJobByID(job_id int64) (*sm.Job, error) {
	stmt, err := this.Pool.Prepare(`
		select
			is_completed,
			is_canceled,
			creation_date,
			completion_date,
			publication_date,
			expiration_date,
			current_step,
			owner_id,
			status
		from job
		where id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(job_id)

	job := sm.Job{ID: job_id}

	err = row.Scan(
		&job.IsCompleted,
		&job.IsCanceled,
		&job.CreationDate,
		&job.CompletionDate,
		&job.PublicationDate,
		&job.ExpirationDate,
		&job.CurrentStep,
		&job.Owner.ID,
		&job.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (this *JobRepository) GetJobByStepID(step_id int64) (*sm.Job, error) {
	stmt, err := this.Pool.Prepare(`
		select
			j.id,
			j.is_completed,
			j.is_canceled,
			j.creation_date,
			j.completion_date,
			j.publication_date,
			j.expiration_date,
			j.current_step,
			j.owner_id,
			j.status
		from job j
		inner join job_step s
			on s.job_id=j.id
		where s.id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(step_id)

	job := sm.Job{}

	err = row.Scan(
		&job.ID,
		&job.IsCompleted,
		&job.IsCanceled,
		&job.CreationDate,
		&job.CompletionDate,
		&job.PublicationDate,
		&job.ExpirationDate,
		&job.CurrentStep,
		&job.Owner.ID,
		&job.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (this *JobRepository) GetJobsForContentOwner(
	owner_id, limit, before_job_id int64) ([]*sm.Job, error) {

	var stmt *sql.Stmt
	var err error

	if limit != -1 && before_job_id != -1 {
		stmt, err = this.Pool.Prepare(`
			select
				id,
				is_completed,
				is_canceled,
				creation_date,
				completion_date,
				publication_date,
				expiration_date,
				current_step,
				status
			from job
			where owner_id=$1 and job.id<$3
			order by id desc
			LIMIT $2
		`)
	} else if limit != -1 {
		stmt, err = this.Pool.Prepare(`
			select
				id,
				is_completed,
				is_canceled,
				creation_date,
				completion_date,
				publication_date,
				expiration_date,
				current_step,
				status
			from job
			where owner_id=$1
			order by id desc
			LIMIT $2
		`)
	} else {
		stmt, err = this.Pool.Prepare(`
			select
				id,
				is_completed,
				is_canceled,
				creation_date,
				completion_date,
				publication_date,
				expiration_date,
				current_step,
				status
			from job
			where owner_id=$1
			order by id desc
		`)
	}

	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if limit != -1 && before_job_id != -1 {
		rows, err = stmt.Query(owner_id, limit, before_job_id)
	} else if limit != -1 {
		rows, err = stmt.Query(owner_id, limit)
	} else {
		rows, err = stmt.Query(owner_id)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]*sm.Job, 0)

	owner := sm.ContentOwner{ID: owner_id}
	for rows.Next() {
		job := sm.Job{Owner: owner}
		err = rows.Scan(
			&job.ID,
			&job.IsCompleted,
			&job.IsCanceled,
			&job.CreationDate,
			&job.CompletionDate,
			&job.PublicationDate,
			&job.ExpirationDate,
			&job.CurrentStep,
			&job.Status)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (this *JobRepository) GetJobSteps(job_id int64, steps *[]*sm.JobStep) error {
	stmt, err := this.Pool.Prepare(`
		select id, task_id
		from job_step
		where job_id=$1
		order by step_order asc
	`)

	if err != nil {
		return err
	}

	rows, err := stmt.Query(job_id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		step := sm.JobStep{}

		err = rows.Scan(&step.ID, &step.TaskID)

		if err != nil {
			steps = nil
			return err
		}

		*steps = append(*steps, &step)
	}

	return nil
}

func (this *JobRepository) GetParamsForStep(step *sm.JobStep) error {
	stmt, err := this.Pool.Prepare(`
		select is_input, name, data_type, value, linked_output_name
		from job_param
		where job_step_id=$1
	`)

	if err != nil {
		return err
	}

	rows, err := stmt.Query(step.ID)

	if err != nil {
		return err
	}
	defer rows.Close()

	var is_input bool
	var name string
	var param sm.JobParam

	step.Input = make(map[string]sm.JobParam)
	step.Output = make(map[string]sm.JobParam)

	for rows.Next() {

		err = rows.Scan(
			&is_input,
			&name,
			&param.DataType,
			&param.Value,
			&param.LinkedOutputName)

		if err != nil {
			step.Output = nil
			step.Input = nil
			return err
		}

		if is_input {
			step.Input[name] = param
		} else {
			step.Output[name] = param
		}
	}

	return nil
}

func (this *JobRepository) SaveStatus(job *sm.Job) error {
	stmt, err := this.Pool.Prepare(`
		update job
		set status=$1
		where id=$2
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(job.Status, job.ID)

	return err
}

//	SaveStepProgress updates the current_step and also saves the output of the previous task to the database
//
func (this *JobRepository) SaveStepProgress(job *sm.Job) error {
	tx, err := this.Pool.DB.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		update job
		set current_step=$1
		where id=$2
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	param_stmt, err := tx.Prepare(`
		insert into job_param
			(job_step_id, name, is_input, data_type, value, linked_output_name)
		values
			($1, $2, false, $3, $4, NULL)
	`)

	if err != nil {
		return err
	}

	defer param_stmt.Close()

	_, err = stmt.Exec(job.CurrentStep, job.ID)

	if err != nil {
		tx.Rollback()
		return err
	}

	if job.CurrentStep > 0 {
		step := job.Steps[job.CurrentStep-1]

		for name, param := range step.Output {
			_, err = param_stmt.Exec(
				step.ID,
				name,
				param.DataType,
				param.Value,
			)

			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (this *JobRepository) CreateJob(job *sm.Job) error {
	tx, err := this.Pool.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		insert into job (
			is_completed, 
			is_canceled, 
			creation_date,
			completion_date,
			publication_date,
			expiration_date,
			current_step,
			owner_id,
			status,
			is_expiration_processed)
		values (false, false, $1, null, $2, $3, 0, $4, $5, false)
		returning id
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	row := stmt.QueryRow(
		job.CreationDate,
		job.PublicationDate,
		job.ExpirationDate,
		job.Owner.ID,
		job.Status,
	)

	err = row.Scan(&job.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	step_stmt, err := tx.Prepare(`
		insert into job_step (job_id, task_id, step_order)
		values ($1, $2, $3)
		returning id
	`)

	if err != nil {
		return err
	}
	defer step_stmt.Close()

	param_stmt, err := tx.Prepare(`
		insert into job_param (
			job_step_id, 
			name, 
			is_input, 
			data_type, 
			value, 
			linked_output_name)
		values ($1, $2, $3, $4, $5, $6)
	`)

	if err != nil {
		return err
	}
	defer param_stmt.Close()

	for i, step := range job.Steps {
		row = step_stmt.QueryRow(job.ID, step.TaskID, i)

		err = row.Scan(&step.ID)
		if err != nil {
			return err
		}

		for name, param := range step.Input {
			_, err = param_stmt.Exec(
				step.ID,
				name,
				true,
				param.DataType,
				param.Value,
				param.LinkedOutputName,
			)
			if err != nil {
				return err
			}
		}

		for name, param := range step.Output {
			_, err = param_stmt.Exec(
				step.ID,
				name,
				false,
				param.DataType,
				param.Value,
				nil,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// SaveFinishedState saves the completion_date, status and sets the is_completed to true
//
func (this *JobRepository) SaveFinishedState(job *sm.Job) error {
	stmt, err := this.Pool.Prepare(`
		update job
		set completion_date=$1,
			status=$2,
			is_completed=true,
			is_canceled=$3
		where id=$4
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		job.CompletionDate,
		job.Status,
		job.IsCanceled,
		job.ID,
	)

	return err
}

func (this *JobRepository) GetJobsExceedingPublicationDate(
	timestamp time.Time,
	limit, offset_id int64) ([]*sm.Job, error) {

	stmt, err := this.Pool.Prepare(`
		select
			id,
			is_completed,
			is_canceled,
			creation_date,
			completion_date,
			publication_date,
			expiration_date,
			current_step,
			owner_id,
			status
		from job
		where 
			(not is_completed) and
			id>$1 and 
			publication_date<=$2
		ORDER BY id ASC
		LIMIT $3
	`)

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(offset_id, timestamp, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]*sm.Job, 0)

	owner := sm.ContentOwner{}
	for rows.Next() {
		job := sm.Job{Owner: owner}
		err = rows.Scan(
			&job.ID,
			&job.IsCompleted,
			&job.IsCanceled,
			&job.CreationDate,
			&job.CompletionDate,
			&job.PublicationDate,
			&job.ExpirationDate,
			&job.CurrentStep,
			&job.Owner.ID,
			&job.Status)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (this *JobRepository) CancelJobsWithExceedingExpirationDate(
	timestamp time.Time) error {
	stmt, err := this.Pool.Prepare(`
		update job set 
			is_completed=true,
			is_canceled=true,
			completion_date=$1,
			status='Expired'
		where
			(not is_completed) and
			(not is_canceled) and
			expiration_date<=$1
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(timestamp)

	if rows_affected, e := res.RowsAffected(); e != nil {
		log.Infof("Checked expired jobs, %d jobs affected", rows_affected)
	}

	return err
}

func (this *JobRepository) GetJobsExceedingExpirationDate(
	timestamp time.Time,
	limit, offset_id int64) ([]*sm.Job, error) {

	stmt, err := this.Pool.Prepare(`
		select
			id,
			is_completed,
			is_canceled,
			creation_date,
			completion_date,
			publication_date,
			expiration_date,
			current_step,
			owner_id,
			status
		from job
		where 
			(not is_completed) and
			id>$1 and 
			expiration_date<=$2
		ORDER BY id ASC
		LIMIT $3
	`)

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(offset_id, timestamp, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]*sm.Job, 0)

	owner := sm.ContentOwner{}
	for rows.Next() {
		job := sm.Job{Owner: owner}
		err = rows.Scan(
			&job.ID,
			&job.IsCompleted,
			&job.IsCanceled,
			&job.CreationDate,
			&job.CompletionDate,
			&job.PublicationDate,
			&job.ExpirationDate,
			&job.CurrentStep,
			&job.Owner.ID,
			&job.Status)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (this *JobRepository) CancelJobsWithExceedingPublicatinDate(
	timestamp time.Time) error {
	stmt, err := this.Pool.Prepare(`
		update job set 
			is_completed=true,
			is_canceled=true,
			completion_date=$1,
			status='Expired'
		where
			(not is_completed) and
			(not is_canceled) and
			publication_date<=$1
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(timestamp)

	if rows_affected, e := res.RowsAffected(); e != nil {
		log.Infof("Checked expired jobs, %d jobs affected", rows_affected)
	}

	return err
}

func (this *JobRepository) GetNewExpiredJobsAt(
	timestamp time.Time,
	limit int64,
	offset_id int64) ([]*sm.Job, error) {

	stmt, err := this.Pool.Prepare(`
		select
			id,
			is_completed,
			is_canceled,
			creation_date,
			completion_date,
			publication_date,
			expiration_date,
			current_step,
			owner_id,
			status
		from job
		where 
			id>$1 and
			(not is_expiration_processed) and 
			expiration_date<=$2
		ORDER BY id ASC
		LIMIT $3
	`)

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(offset_id, timestamp, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]*sm.Job, 0)

	owner := sm.ContentOwner{}
	for rows.Next() {
		job := sm.Job{Owner: owner}
		err = rows.Scan(
			&job.ID,
			&job.IsCompleted,
			&job.IsCanceled,
			&job.CreationDate,
			&job.CompletionDate,
			&job.PublicationDate,
			&job.ExpirationDate,
			&job.CurrentStep,
			&job.Owner.ID,
			&job.Status)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (this *JobRepository) ExpireJobsBefore(timestamp time.Time) error {
	stmt, err := this.Pool.Prepare(`
		update job set 
			is_expiration_processed=true
		where
			(not is_expiration_processed) and
			expiration_date<=$1
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(time.Now())

	if rows_affected, e := res.RowsAffected(); e != nil {
		log.Infof("Checked expired jobs, %d jobs affected", rows_affected)
	}

	return err
}

func (this *JobRepository) DeleteParamsForJob(job_id int64) error {
	stmt, err := this.Pool.Prepare(`
		delete from job_param
		using job_step
		where job_step.id=job_param.job_step_id and job_step.job_id=$1
	`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(job_id)

	if rows_affected, e := res.RowsAffected(); e != nil {
		log.Infof("Deleted params for job=%d, %d jobs affected",
			job_id,
			rows_affected)
	}

	return err
}
