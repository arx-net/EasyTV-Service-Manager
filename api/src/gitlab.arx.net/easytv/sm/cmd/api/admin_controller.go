package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-chi/chi"

	"gitlab.arx.net/arx/gosession"
	"gitlab.arx.net/arx/httpio"
	"gitlab.arx.net/easytv/sm"
)

type AdminController struct {
	sessions          *gosession.SessionStore
	module_repository sm.ModuleRepository
	module_service    sm.ModuleService
	owner_repository  sm.ContentOwnerRepository
	owner_service     sm.ContentOwnerService
}

func (this *AdminController) GetLog(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	http.ServeFile(w, r, "/var/log/sm/sm.log")
}

func (this *AdminController) SrtCommand(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	// Read all
	json_str, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		InternalServerError(w, err)
		return
	}

	// Unmarshal
	var args []string
	err = json.Unmarshal(json_str, &args)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	logrus.Infof("Exec: %v %v", os.Getenv("SRT_CMD"), strings.Join(args, " "))

	output, err := exec.Command(os.Getenv("SRT_CMD"), args...).Output()

	if err != nil {
		InternalServerError(w, err)
		return
	}

	httpio.WriteText(w, http.StatusOK, string(output))
}

func (this *AdminController) CreateService(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	data, _ := httpio.ReadJSON(r)

	name, _ := data["name"].(string)
	desc, _ := data["description"].(string)

	module, err := this.module_service.CreateService(name, desc)

	if err == nil && module != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "Service created",
			"api_key":     module.ApiKey,
			"service_id":  module.ID})
	} else if err == sm.ErrServiceNameTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"name\" parameter"})
	} else if err == sm.ErrServiceDescTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"description\" parameter"})
	} else if err == sm.ErrServiceNameInUse {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeServiceNameInUse,
			"description": "The name is alredy in use"})
	} else {
		InternalServerError(w, err)
	}
}

func (this *AdminController) GetServices(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}
	modules := make([]map[string]interface{}, 0)

	err = this.module_repository.GetModules(&modules)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Returning the list of services",
		"services":    modules})
}

func (this *AdminController) GetService(w http.ResponseWriter, r *http.Request) {
	session, err := this.sessions.Get(r, w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	id, atoi_err := strconv.ParseInt(chi.URLParam(r, "service_id"), 10, 64)

	if atoi_err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"id\" parameter"})
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

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Success",
		"service": map[string]interface{}{
			"id":          module.ID,
			"name":        module.Name,
			"description": module.Description,
			"api_key":     module.ApiKey,
			"enabled":     module.Enabled}})
}

func (this *AdminController) SetAvailability(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	data, _ := httpio.ReadJSON(r)

	id, atoi_err := strconv.ParseInt(chi.URLParam(r, "service_id"), 10, 64)

	if atoi_err != nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"id\" parameter"})
		return
	}

	var enable bool

	if value, ok := data["enable"].(bool); !ok {
		if value, ok := data["disable"].(bool); !ok {
			httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"code":        sm.CodeMissingInput,
				"description": "Missing \"enable\" parameter"})
		} else {
			enable = !value
		}
	} else {
		enable = value
	}

	err := this.module_service.SetAvailability(id, enable)
	if err == nil {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "Success"})
	} else if err == sm.ErrNotFound {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("Service with id=%d was not found", id)})
	} else {
		InternalServerError(w, err)
	}
}

func (this *AdminController) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySessionWithRole(session, w, sm.RoleAdmin) {
		return
	}

	data, _ := httpio.ReadJSON(r)

	name, _ := data["name"].(string)
	email, _ := data["email"].(string)
	username, _ := data["username"].(string)

	content_owner, err := this.owner_service.CreateContentOwner(
		name,
		username,
		email,
		this.owner_service.GenerateRandomPassword(15))

	if err == nil {
		// This is the last time the server will have access to the password in plaintext
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":                   sm.OK,
			"description":            "Registration was successful",
			"contenet_owner_id":      content_owner.ID,
			"content_owner_password": content_owner.Password})
	} else if err == sm.ErrOwnerNameTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"name\" string parameter"})
	} else if err == sm.ErrOwnerUsernameTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"username\" string parameter"})
	} else if err == sm.ErrOwnerEmailTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "Missing valid \"email\" string parameter"})
	} else if err == sm.ErrOwnerNameExists {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeContentOwnerNameExists,
			"description": "A content owner with this name already exists"})
	} else if err == sm.ErrOwnerUsernameExists {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeContentOwnerUsernameExists,
			"description": "A content owner with this username already exists"})
	} else if err == sm.ErrOwnerEmailExists {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeContentOwnerEmailExists,
			"description": "A content owner with this email already exists"})
	} else {
		InternalServerError(w, err)
	}
}
