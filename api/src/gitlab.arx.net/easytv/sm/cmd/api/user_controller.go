package main

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.arx.net/arx/gosession"
	"gitlab.arx.net/arx/httpio"
	"gitlab.arx.net/easytv/sm"
)

type UserController struct {
	sessions         *gosession.SessionStore
	owner_repository sm.ContentOwnerRepository
	owner_service    sm.ContentOwnerService
	admin_repository sm.AdminRepository
	admin_service    sm.AdminService
}

func VerifySession(session *gosession.Session, w http.ResponseWriter) bool {
	if session != nil {
		if _, ok := session.Data["user_id"]; ok {
			if last_accessed, ok := session.Data["last_accessed"].(int64); ok {
				if time.Now().Unix()-last_accessed <= sm.SessionExpiration {
					session.Data["last_accessed"] = time.Now().Unix()
					session.Save()
					return true
				} else {
					session.Destroy(w)
				}
			}
		}
	}
	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.CodeNoSession,
		"description": "No valid session"})
	return false
}

func VerifySessionWithRole(session *gosession.Session, w http.ResponseWriter, role int) bool {
	if session != nil {
		if _, ok := session.Data["user_id"]; ok {
			if stored_role, ok := session.Data["role"].(int); ok && stored_role == role {
				if last_accessed, ok := session.Data["last_accessed"].(int64); ok {
					if time.Now().Unix()-last_accessed <= sm.SessionExpiration {
						session.Data["last_accessed"] = time.Now().Unix()
						session.Save()
						return true
					} else {
						session.Destroy(w)
					}
				}
			}
		}
	}
	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.CodeNoSession,
		"description": "No valid session"})
	return false
}

func (this *UserController) Login(w http.ResponseWriter, r *http.Request) {
	data, _ := httpio.ReadJSON(r)

	var username, password string
	var ok bool

	if username, ok = data["username"].(string); !ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "\"username\" string parameter is missing"})
		return
	}

	if password, ok = data["password"].(string); !ok {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeMissingInput,
			"description": "\"password\" string parameter is missing"})
		return
	}

	owner, err := this.owner_service.Login(username, password)

	if err != nil {
		InternalServerError(w, err)
		return
	} else if owner == nil {
		// user is not a content onwer
		// check if he is an admin
		admin, err := this.admin_service.Login(username, password)

		if err != nil {
			InternalServerError(w, err)
			return
		} else if admin != nil {
			session, err := this.sessions.New(w)
			if err != nil {
				InternalServerError(w, err)
				return
			}
			session.Data["user_id"] = admin.ID
			session.Data["username"] = username
			session.Data["role"] = sm.RoleAdmin
			session.Data["last_accessed"] = time.Now().Unix()
			session.Save()

			httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"code":          sm.OK,
				"description":   fmt.Sprintf("Success, hello %s", username),
				"session_token": session.ID,
				"is_admin":      true})
			return
		}

		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNoSession,
			"description": "Username or password are not correct"})
		return
	}

	session, session_err := this.sessions.New(w)

	if session_err != nil {
		InternalServerError(w, session_err)
		return
	}

	session.Data["user_id"] = owner.ID
	session.Data["role"] = sm.RoleContentOwner
	session.Data["last_accessed"] = time.Now().Unix()

	session.Save()

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":          sm.OK,
		"description":   fmt.Sprintf("Success, hello %s", owner.Name),
		"session_token": session.ID})
}

func (this *UserController) Ping(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySession(session, w) {
		return
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Pong"})
}

func (this *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySession(session, w) {
		return
	}

	err := session.Destroy(w)

	if err != nil {
		InternalServerError(w, err)
		return
	}

	httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"code":        sm.OK,
		"description": "Logged out successfuly"})
}

func (this *UserController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	session, _ := this.sessions.Get(r, w)

	if !VerifySession(session, w) {
		return
	}

	data, _ := httpio.ReadJSON(r)

	old_password, _ := data["old_password"].(string)
	new_password, _ := data["new_password"].(string)
	new_password_verification, _ := data["new_password_verification"].(string)

	if new_password != new_password_verification {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeNewPasswordDoesntMatchVerification,
			"description": "'new_password' and 'new_password_verifycation' doesn't match",
		})
		return
	}

	user_id, _ := session.Data["user_id"].(int64)
	role, ok := session.Data["role"].(int)

	if !ok {
		InternalServerError(w, fmt.Errorf("role not defined in session with id=%v", session.ID))
		return
	}

	var err error

	if role == sm.RoleAdmin {
		err = this.admin_service.ChangePassword(user_id, old_password, new_password)
	} else if role == sm.RoleContentOwner {
		err = this.owner_service.ChangePassword(user_id, old_password, new_password)
	}

	if err == nil {
		session.Destroy(w)
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.OK,
			"description": "password is changed",
		})
	} else if err == sm.ErrPasswordTooShort {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodePasswordIsTooShort,
			"description": "Password is too short",
		})
	} else if err == sm.ErrInvalidCredentials {
		httpio.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"code":        sm.CodeInvalidCredentials,
			"description": "Credentials are not correct",
		})
	} else {
		InternalServerError(w, err)
	}
}
