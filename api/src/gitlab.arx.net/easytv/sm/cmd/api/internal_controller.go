package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"encoding/hex"
	
	"github.com/go-chi/chi"

	"gitlab.arx.net/arx/httpio"
	"gitlab.arx.net/easytv/sm"

)

type InternalController struct {
	module_repository sm.ModuleRepository
	job_repository    sm.JobRepository
	job_service       sm.JobService
	task_repository   sm.TaskRepository
	task_service      sm.TaskService
	owner_repository  sm.ContentOwnerRepository
	asset_repository  sm.AssetRepository
	asset_service     sm.AssetService
}

//
//	From the service's point of view, the job_id is the
//	id of the current step
//

// Checks if the request contains a valid API key
//
// If the key is correct returns a valid connection as well as the module
//
// If the key is not correct both returned values will be nil, this function
// will always close the connection in this case
func (this *InternalController) check_api_key(
	w http.ResponseWriter,
	r *http.Request) *sm.Module {

	var api_key string

	if api_keys, ok := r.Header[sm.EasyTVApiKeyHeader]; ok {
		api_key = api_keys[0]
	} else {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNoSession,
			"description": "No valid api key"})
		return nil
	}

	hashed_key := sha256.Sum256([]byte(api_key))
	hashed_key_str := hex.EncodeToString(hashed_key[:])

	module, err := this.module_repository.GetModuleByKey(hashed_key_str[:])

	if err == sql.ErrNoRows {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNoSession,
			"description": "No valid api key"})

		return nil
	} else if err != nil {
		InternalServerError(w, err)
		return nil
	}

	return module
}

func (this *InternalController) RegisterTask(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	data, _ := httpio.ReadJSON(r)

	input := make(map[string]sm.ParamType)
	if input_data, ok := data["input"].(map[string]interface{}); ok {
		for name, value := range input_data {

			str_val, ok := value.(string)
			if !ok {
				httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
					"code":        sm.CodeInvalidInput,
					"description": fmt.Sprintf("\"%s\" should have a string value", name)})
				return
			}
			input[name] = sm.GetParamTypeFromSting(str_val)
		}
	}

	output := make(map[string]sm.ParamType)
	if output_data, ok := data["output"].(map[string]interface{}); ok {
		for name, value := range output_data {

			str_val, ok := value.(string)
			if !ok {
				httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
					"code":        sm.CodeInvalidOutput,
					"description": fmt.Sprintf("\"%s\" should have a string value", name)})
				return
			}
			output[name] = sm.GetParamTypeFromSting(str_val)
		}
	}

	name, _ := data["name"].(string)
	desc, _ := data["description"].(string)
	start_url, _ := data["start_url"].(string)
	cancel_url, _ := data["cancel_url"].(string)

	if cancel_url == "" {
		cancel_url, _ = data["cancel_rest_url"].(string)
		cancel_url = fmt.Sprintf("REST %v", cancel_url)
	}

	task, err := this.task_service.RegisterTask(
		module.ID, name, desc, start_url, cancel_url, input, output)

	switch err {
	case nil:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "The task was registered successfully",
			"task_id":     task.ID})
	case sm.ErrTaskNameTooShort:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "A valid \"name\" string parameter is missing"})
	case sm.ErrTaskDescTooShort:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "A valid \"description\" string parameter is missing"})
	case sm.ErrTaskAlreadyExists:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeTaskAlreadyExists,
			"description": "A task with that name already exists"})
	case sm.ErrInvalidStartUrl:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidStartUrl,
			"description": "The \"start_url\" is not a valid URL"})
	case sm.ErrInvalidCancelUrl:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidCancelUrl,
			"description": "The \"cancel_url\" is not a valid URL"})
	case sm.ErrInvalidInputType:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidInput,
			"description": "Input should be (\"string\", \"int\" or \"double\")"})
	case sm.ErrInvalidOutputType:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidOutput,
			"description": "Output should be (\"string\", \"int\" or \"double\")"})
	case sm.ErrEmptyInput:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeTaskNoInputParameter,
			"description": "There should be at least one input parameter"})
	case sm.ErrEmptyOutput:
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeTaskNoOutputParameter,
			"description": "There should be at least one output parameter"})
	default:
		InternalServerError(w, err)
	}
}

func (this *InternalController) SetTaskAvailability(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "task_id"), 10, 64)

	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"id\" parameter"})
		return
	}

	data, _ := httpio.ReadJSON(r)

	var is_enabled bool

	if disabled, ok := data["disabled"].(bool); !ok {
		if enabled, ok := data["enabled"].(bool); !ok {
			httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"code":        sm.CodeMissingInput,
				"description": "Missing \"disabled\" parameter"})
			return
		} else {
			is_enabled = enabled
		}
	} else {
		is_enabled = !disabled
	}

	err = this.task_service.SetAvailability(id, is_enabled)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "success"})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": "Task doesn't exist"})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "task_id"), 10, 64)

	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"id\" parameter"})
		return
	}

	err = this.task_service.DeleteTask(id)

	if err == nil {
		httpio.WriteJSON(w, sm.OK, map[string]interface{}{
			"code":        sm.OK,
			"description": "success"})

	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": "Task doesn't exist"})
	} else if err == sm.ErrTaskIsEnabled {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeTaskNotDisabled,
			"description": "The task must be disabled before it is deleted"})
	} else if err == sm.ErrTaskHasJobStepsInProgress {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeTaskHasActiveJobs,
			"description": "The task has active jobs and can't be deleted"})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) GetTasks(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	tasks, err := this.task_repository.GetTasks(module.ID, false)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	task_data := make([]map[string]interface{}, len(tasks))

	for i, task := range tasks {

		input := make(map[string]string)

		for name, param_type := range task.Input {
			input[name] = sm.ParamTypeStr(param_type)
		}

		output := make(map[string]string)

		for name, param_type := range task.Output {
			output[name] = sm.ParamTypeStr(param_type)
		}

		task_data[i] = map[string]interface{}{
			"id":          task.ID,
			"name":        task.Name,
			"description": task.Description,
			"start_url":   task.StartUrl,
			"cancel_url":  task.CancelUrl,
			"enabled":     task.Enabled,
			"input":       input,
			"output":      output}
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"tasks":       task_data})

}

func (this *InternalController) GetJobs(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	limit, err := strconv.ParseInt(chi.URLParam(r, "limit"), 10, 64)
	if err != nil {
		limit = -1
	}

	before_job_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		before_job_id = -1
	}

	jobs_data, err := this.job_repository.GetJobsStepsForModule(
		module.ID, limit, before_job_id)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	var next_url *string
	if limit != -1 && int64(len(jobs_data)) == limit {
		next_url = new(string)
		*next_url = fmt.Sprintf("/internal/job/limit/%d/before/%d",
			limit,
			jobs_data[len(jobs_data)-1]["id"].(int64))
	}
	httpio.WriteJSON(w, sm.OK, map[string]interface{}{
		"code":        sm.OK,
		"description": "success",
		"next":        next_url,
		"jobs":        jobs_data})
}

func (this *InternalController) GetJob(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	step_id, id_err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)

	if id_err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"job_id\" parameter"})
		return
	}

	job_data, err := this.job_repository.GetJobStepForModule(
		step_id, module.ID)

	if job_data == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": "Job doesn't exist"})
		return
	} else if err != nil {
		InternalServerError(w, err)
		return
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "success",
		"job":         job_data,
	})
}

func (this *InternalController) SetJobStatus(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	step_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)

	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"job_id\" parameter"})
		return
	}

	data, _ := httpio.ReadJSON(r)

	status, ok := data["status"].(string)

	if !ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"status\""})
		return
	}

	err = this.job_service.SetJobStatusForStep(step_id, status)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "Success"})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": "Job doesn't exist"})
	} else if err == sm.ErrJobStatusNotUpdatable {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobStatusNotUpdatable,
			"description": "The status can't be updated because the job has been completed"})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) CancelJob(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	step_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	err = this.job_service.CancelJobAsModule(module, step_id)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "The job was canceled",
		})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A job with id=%d doesn't exist", step_id),
		})
	} else if err == sm.ErrJobIsCanceled {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobAlreadyCanceled,
			"description": "The job is already canceled",
		})
	} else if err == sm.ErrJobIsCompleted {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeJobAlreadyCompleted,
			"description": "The job is already completed",
		})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) FinishJob(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	step_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	data, err := httpio.ReadJSON(r)

	output, ok := data["output"].(map[string]interface{})

	if !ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing output values",
		})
		return
	}

	err = this.job_service.FinishJobStep(step_id, module, output)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": fmt.Sprintf("The job %d was finished", step_id),
		})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A job with id=%d doesn't exist", step_id),
		})
	} else if err == sm.ErrJobIsCompleted || err == sm.ErrJobIsCanceled {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotCompletable,
			"description": "The job can't be completed by this service",
		})
	} else if e, ok := err.(*sm.ErrInvalidTaskOutput); ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidOutput,
			"description": e.Message,
		})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) UploadAsset(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	step_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	r.ParseMultipartForm(1024 * 512) // 1MB

	file, handler, err := r.FormFile("asset")

	if err != nil {
		InternalServerError(w, err)
		return
	}

	defer file.Close()

	asset, err := this.asset_service.CreateAsset(
		step_id, module, file, handler.Filename, handler.Size)

	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "Asset was uploaded",
			"asset_id":    asset.ID,
			"asset_url":   fmt.Sprintf("/asset/%s", asset.UrlParam),
		})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("A job with id=%d doesn't exist", step_id),
		})
	} else if err == sm.ErrJobIsCanceled || err == sm.ErrJobIsCompleted {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeForbiddenAsset,
			"description": "You can't upload an asset for a job that is completed or canceled",
		})
	} else {
		InternalServerError(w, err)
	}
}

func (this *InternalController) GetAssets(w http.ResponseWriter, r *http.Request) {
	module := this.check_api_key(w, r)
	if module == nil {
		return // invalid API key, check_api_key handled the response
	}

	job_id, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid job id"})
		return
	}

	assets := make([]*sm.Asset, 0)
	err = this.asset_repository.GetAssetsForJob(job_id, &assets)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	asset_json := make([]map[string]interface{}, len(assets))

	for index, asset := range assets {
		asset_json[index] = map[string]interface{}{
			"asset_id":   asset.ID,
			"asset_url":  fmt.Sprintf("/asset/%s", asset.UrlParam),
			"asset_size": asset.Size,
		}
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"assets":      asset_json,
	})
}

func (this *InternalController) DownloadAsset(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "asset_param")

	asset, err := this.asset_repository.GetAssetByUrlParam(param)

	if err != nil {
		InternalServerError(w, err)
	} else if asset == nil {
		httpio.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": "Asset doesn't exist",
		})
	} else {
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachement; filename=%s", url.QueryEscape(filepath.Base(asset.Path))))
		http.ServeFile(w, r, asset.Path)
	}
}
