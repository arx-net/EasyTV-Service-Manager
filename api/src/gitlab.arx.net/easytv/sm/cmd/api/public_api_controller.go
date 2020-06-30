package main

import (
	"fmt"
	"net/http"
	"strconv"
//	"time"

	"github.com/go-chi/chi"
	"gitlab.arx.net/arx/gosession"
	"gitlab.arx.net/arx/httpio"
	"gitlab.arx.net/easytv/sm"
)

type PublicApiController struct {
	sessions          *gosession.SessionStore
	module_repository sm.ModuleRepository
	task_repository   sm.TaskRepository
	job_repository    sm.JobRepository
	owner_repository  sm.ContentOwnerRepository
	job_service       sm.JobService
}

func (this *PublicApiController) GetServices(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	var modules []*sm.Module

	err = this.module_repository.GetModulesOBJ(&modules)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	modules_json_array := make([]map[string]interface{}, 0)

	for _, module := range modules {
		tasks, err := this.task_repository.GetTasks(module.ID, false)

		if err != nil {
			InternalServerError(w, err)
			return
		}

		tasks_json_array := make([]map[string]interface{}, 0)

		for _, task := range tasks {
			input := make(map[string]string)

			for name, param_type := range task.Input {
				input[name] = sm.ParamTypeStr(param_type)
			}

			output := make(map[string]string)

			for name, param_type := range task.Output {
				output[name] = sm.ParamTypeStr(param_type)
			}

			tasks_json_array = append(tasks_json_array, map[string]interface{}{
				"name":        task.Name,
				"id":          task.ID,
				"description": task.Description,
				"enabled":     task.Enabled,
				"input":       input,
				"output":      output})
		}

		modules_json_array = append(modules_json_array, map[string]interface{}{
			"name":        module.Name,
			"id":          module.ID,
			"description": module.Description,
			"enabled":     module.Enabled,
			"tasks":       tasks_json_array})
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"services":    modules_json_array})
}

func (this *PublicApiController) GetService(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "service_id"), 10, 64)

	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"service_id\" parameter"})
		return
	}

	module, err := this.module_repository.GetModuleByID(id)

	if module == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A service with id=%d doesn't exist", id)})
		return
	} else if err != nil {
		InternalServerError(w, err)
		return
	}

	tasks, err := this.task_repository.GetTasks(module.ID, false)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	tasks_json_array := make([]map[string]interface{}, 0)

	for _, task := range tasks {
		input := make(map[string]string)

		for name, param_type := range task.Input {
			input[name] = sm.ParamTypeStr(param_type)
		}

		output := make(map[string]string)

		for name, param_type := range task.Output {
			output[name] = sm.ParamTypeStr(param_type)
		}

		tasks_json_array = append(tasks_json_array, map[string]interface{}{
			"name":        task.Name,
			"id":          task.ID,
			"description": task.Description,
			"enabled":     task.Enabled,
			"input":       input,
			"output":      output})
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"service": map[string]interface{}{
			"name":        module.Name,
			"id":          module.ID,
			"description": module.Description,
			"enabled":     module.Enabled,
			"tasks":       tasks_json_array}})
}

func (this *PublicApiController) GetJobs(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	limit, err := strconv.ParseInt(chi.URLParam(r, "limit"), 10, 64)
	if err != nil {
		limit = -1
	}

	before_job_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		before_job_id = -1
	}

	uid, _ := session.Data["user_id"].(int64)
	jobs, err := this.job_repository.GetJobsForContentOwner(uid, limit, before_job_id)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	jobs_json := make([]map[string]interface{}, 0)

	for _, job := range jobs {
		err = this.job_repository.GetJobSteps(job.ID, &job.Steps)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		var completion_date *int64
		var output map[string]interface{}
		var current_step *int

		if job.IsCompleted {
			completion_date = new(int64)
			*completion_date = job.CompletionDate.Unix()

			if !job.IsCanceled {
				last_step := job.Steps[len(job.Steps)-1]

				err = this.job_repository.GetParamsForStep(last_step)

				if err != nil {
					InternalServerError(w, err)
					return
				}

				output = make(map[string]interface{})
				for name, param := range last_step.Output {
					output[name] = param.Value
				}
			}

		} else {
			current_step = &job.CurrentStep
		}

		tasks := make([]map[string]interface{}, len(job.Steps))

		for index, step := range job.Steps {
			task, err := this.task_repository.GetTask(step.TaskID)
			if err != nil {
				InternalServerError(w, err)
				return
			}
			tasks[index] = map[string]interface{}{
				"task_id":   task.ID,
				"task_name": task.Name,
			}
		}

		jobs_json = append(jobs_json, map[string]interface{}{
			"id":               job.ID,
			"is_completed":     job.IsCompleted,
			"is_canceled":      job.IsCanceled,
			"status":           job.Status,
			"creation_date":    job.CreationDate.Unix(),
			"completion_date":  completion_date,
			"publication_date": job.PublicationDate.Unix(),
			"expiration_date":  job.ExpirationDate.Unix(),
			"tasks":            tasks,
			"current_task":     current_step,
			"output":           output})
	}

	var next_url *string
	if limit != -1 && int64(len(jobs_json)) == limit {
		next_url = new(string)
		*next_url = fmt.Sprintf("/api/job/limit/%d/before/%d",
			limit,
			jobs[len(jobs)-1].ID)
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"next":        next_url,
		"jobs":        jobs_json})

}

func (this *PublicApiController) GetJob(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	job_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	job, err := this.job_repository.GetJobByID(job_id)

	uid, _ := session.Data["user_id"].(int64)

	if err != nil {
		InternalServerError(w, err)
		return
	} else if job == nil || job.Owner.ID != uid {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A job with id=%d doesn't exist", job_id)})
		return
	}

	err = this.job_repository.GetJobSteps(job.ID, &job.Steps)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var completion_date *int64
	var output map[string]interface{}
	var current_step *int

	if job.IsCompleted {
		completion_date = new(int64)
		*completion_date = job.CompletionDate.Unix()

		if !job.IsCanceled {
			last_step := job.Steps[len(job.Steps)-1]

			err = this.job_repository.GetParamsForStep(last_step)

			if err != nil {
				InternalServerError(w, err)
				return
			}

			output = make(map[string]interface{})
			for name, param := range last_step.Output {
				output[name] = param.Value
			}
		}
	} else {
		current_step = &job.CurrentStep
	}

	tasks := make([]map[string]interface{}, len(job.Steps))

	for index, step := range job.Steps {
		task, err := this.task_repository.GetTask(step.TaskID)
		if err != nil {
			InternalServerError(w, err)
			return
		}
		tasks[index] = map[string]interface{}{
			"task_id":   task.ID,
			"task_name": task.Name,
		}
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"job": map[string]interface{}{
			"id":               job.ID,
			"is_completed":     job.IsCompleted,
			"is_canceled":      job.IsCanceled,
			"status":           job.Status,
			"creation_date":    job.CreationDate.Unix(),
			"completion_date":  completion_date,
			"publication_date": job.PublicationDate.Unix(),
			"expiration_date":  job.ExpirationDate.Unix(),
			"tasks":            tasks,
			"current_task":     current_step,
			"output":           output}})
}

func (this *PublicApiController) PostJob(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	data, _ := httpio.ReadJSON(r)

	user_id, _ := session.Data["user_id"].(int64)
	publication_date, _ := data["publication_date"].(float64)
	expiration_date, _ := data["expiration_date"].(float64)
	tasks_i, ok := data["tasks"].([]interface{})

	if !ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidInput,
			"description": "Missing valid \"tasks\" array",
		})
		return
	}

	tasks := make([]map[string]interface{}, len(tasks_i))
	for i, task := range tasks_i {
		tasks[i], ok = task.(map[string]interface{})
		if !ok {
			httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"code":        sm.CodeInvalidInput,
				"description": "Missing valid \"tasks\" array",
			})
			return
		}
	}

	job, err := this.job_service.CreateJob(user_id, int64(publication_date), int64(expiration_date), tasks)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "Job created",
			"job_id":      job.ID,
		})
	} else if err == sm.ErrInvalidPublicationDate {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Invalid \"publication_date\"",
		})
	} else if err == sm.ErrInvalidExpirationDate {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidPublicationdate,
			"description": "Invalid \"expiration_date\"",
		})
	} else if err == sm.ErrEmptyTasks {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "No tasks where given",
		})
	} else if err == sm.ErrMissingTaskID {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid task object",
		})
	} else if e, ok := err.(*sm.ErrTaskNotFound); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": fmt.Sprintf("Task (%v) doesn't exist", e.TaskID),
		})
	} else if e, ok := err.(*sm.ErrTaskIsDisabled); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobWithDisabledTasks,
			"description": fmt.Sprintf("Task (%v) is disabled", e.TaskID),
		})
	} else if e, ok := err.(*sm.ErrModuleIsDisabled); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobWithDisabledTasks,
			"description": fmt.Sprintf("Module (%v) is disabled", e.ModuleID),
		})
	} else if e, ok := err.(*sm.ErrInvalidTaskInput); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidInput,
			"description": e.Message,
		})
	} else if e, ok := err.(*sm.ErrTaskOutputNotFound); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeLinkedOutputNotFound,
			"description": e.Error(),
		})
	} else if e, ok := err.(*sm.ErrLinkedParameterNotTheSameType); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeLinkedParameterNotTheSameType,
			"description": e.Error(),
		})
	} else {
		InternalServerError(w, err)
	}
}

func (this *PublicApiController) CancelJob(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleContentOwner) {
		return
	}

	job_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	user_id, _ := session.Data["user_id"].(int64)

	err = this.job_service.CancelJobAsOwner(user_id, job_id)
	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "The job was canceled",
		})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A job with id=%d doesn't exist", job_id),
		})
	} else if err == sm.ErrJobIsCompleted {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobAlreadyCompleted,
			"description": "The job is already completed",
		})
	} else if err == sm.ErrJobIsCanceled {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobAlreadyCanceled,
			"description": "The job is already canceled",
		})
	} else {
		InternalServerError(w, err)
	}
}
