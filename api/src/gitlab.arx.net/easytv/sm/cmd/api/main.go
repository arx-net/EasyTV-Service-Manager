package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.arx.net/easytv/sm/db"

	"gitlab.arx.net/arx/gosession"
	"gitlab.arx.net/arx/httpio"
	"gitlab.arx.net/easytv/sm"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-chi/chi"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, r)
	})
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"addr":   r.RemoteAddr,
			"method": r.Method,
			"url":    r.URL.String(),
			"agent":  r.UserAgent(),
		}).Info()
		next.ServeHTTP(w, r)
	})
}

func InternalServerError(w http.ResponseWriter, err error) {
	log.Error(err)
	httpio.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"code":        sm.CodeInternalServerError,
		"description": "Internal server error"})
}

func main() {
	// Setup logging
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)

	os.MkdirAll("/var/log/sm", os.ModePerm)
	f, err := os.OpenFile("/var/log/sm/sm.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	log.SetOutput(os.Stdout)

	// setup sessions store

	sessions := gosession.NewHeaderBasedSessionStore(
		&gosession.MemcachedProvider{Connection: memcache.New("service_manager_cache:11211"), KeyPrefix: "sm"},
		sm.EasyTVSessionHeader,
		false)

	// Setup DB connection pool
	pool, err := db.Open()

	if err != nil {
		log.Fatal(err)
	}

	MAX_CONNECTION, err := strconv.Atoi(os.Getenv("MAX_CONNECTIONS"))
	if err != nil {
		MAX_CONNECTION = 0
	}

	IDLE_CONNECTIONS, err := strconv.Atoi(os.Getenv("IDLE_CONNECTIONS"))
	if err != nil {
		IDLE_CONNECTIONS = 0
	}

	MAX_CONN_LIFETIME, err := strconv.Atoi(os.Getenv("MAX_CONN_LIFETIME"))
	if err != nil {
		MAX_CONN_LIFETIME = 60
	}

	pool.DB.SetMaxOpenConns(MAX_CONNECTION)
	pool.DB.SetMaxIdleConns(IDLE_CONNECTIONS)
	pool.DB.SetConnMaxLifetime(time.Duration(MAX_CONN_LIFETIME) * time.Minute)

	// repositories
	module_repository := &db.ModuleRepository{Pool: pool}
	task_repository := &db.TaskRepository{Pool: pool}
	job_repository := &db.JobRepository{Pool: pool}
	owner_repository := &db.ContentOwnerRepository{Pool: pool}
	asset_repository := &db.AssetRepository{Pool: pool}
	admin_repository := &db.AdminRepository{Pool: pool}

	// services
	task_service := sm.NewTaskService(task_repository, job_repository)
	job_service := sm.NewJobService(
		job_repository, task_repository, module_repository, owner_repository)
	module_service := sm.NewModuleService(module_repository)
	owner_service := sm.NewContentOwnerService(owner_repository)
	admin_service := sm.NewAdminService(admin_repository)
	asset_service := sm.NewAssetService(asset_repository, job_repository, task_repository)

	// controllers

	public_controller := PublicApiController{
		sessions:          sessions,
		module_repository: module_repository,
		task_repository:   task_repository,
		job_repository:    job_repository,
		owner_repository:  owner_repository,
		job_service:       job_service,
	}

	user_controller := UserController{
		sessions:         sessions,
		owner_repository: owner_repository,
		admin_repository: admin_repository,
		owner_service:    owner_service,
		admin_service:    admin_service,
	}

	adm_controller := AdminController{
		sessions:          sessions,
		module_repository: module_repository,
		owner_repository:  owner_repository,
		module_service:    module_service,
		owner_service:     owner_service,
	}

	internal_controller := InternalController{
		module_repository: module_repository,
		job_repository:    job_repository,
		task_repository:   task_repository,
		owner_repository:  owner_repository,
		asset_repository:  asset_repository,
		task_service:      task_service,
		job_service:       job_service,
		asset_service:     asset_service,
	}

	// Register routes
	router := chi.NewRouter()

	router.Use(CorsMiddleware)
	router.Use(LoggerMiddleware)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpio.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"code":        sm.CodeNotFound,
			"description": fmt.Sprintf("The endpoint \"%v\" doesn't exist", r.URL.Path),
		})
	})

	/*
	 *	Admin API
	 */
	router.Route("/adm", func(r chi.Router) {
		r.Post("/service", adm_controller.CreateService)
		r.Get("/service", adm_controller.GetServices)
		r.Put("/service/{service_id}", adm_controller.SetAvailability)
		r.Get("/service/{service_id}", adm_controller.GetService)
		r.Post("/user/register", adm_controller.RegisterOwner)
		r.Post("/srt", adm_controller.SrtCommand)
		r.Get("/log", adm_controller.GetLog)
	})

	/*
	 *	Routes for internal API
	 */
	router.Route("/internal", func(r chi.Router) {
		r.Route("/task", func(r chi.Router) {
			r.Get("/", internal_controller.GetTasks)
			r.Post("/", internal_controller.RegisterTask)
			r.Put("/{task_id}", internal_controller.SetTaskAvailability)
			r.Delete("/{task_id}", internal_controller.DeleteTask)
		})

		r.Route("/job", func(r chi.Router) {
			r.Get("/", internal_controller.GetJobs)
			r.Get("/limit/{limit}", internal_controller.GetJobs)
			r.Get("/limit/{limit}/before/{job_id}", internal_controller.GetJobs)

			r.Route("/{job_id}", func(r chi.Router) {
				r.Get("/", internal_controller.GetJob)
				r.Put("/", internal_controller.SetJobStatus)
				r.Delete("/", internal_controller.CancelJob)
				r.Post("/finish", internal_controller.FinishJob)

				r.Route("/asset", func(r chi.Router) {
					r.Get("/", internal_controller.GetAssets)
					r.Post("/", internal_controller.UploadAsset)
				})
			})
		})
	})
	router.Get("/asset/{asset_param}", internal_controller.DownloadAsset)

	/*
	 *	Routes for the public api
	 */
	router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.HandleFunc("/login", user_controller.Login)
			r.HandleFunc("/ping", user_controller.Ping)
			r.HandleFunc("/logout", user_controller.Logout)
			r.Post("/change_password", user_controller.ChangePassword)
		})

		r.Route("/service", func(r chi.Router) {
			r.Get("/", public_controller.GetServices)
			r.Get("/{service_id}", public_controller.GetService)
		})

		r.Route("/job", func(r chi.Router) {
			r.Get("/", public_controller.GetJobs)
			r.Get("/limit/{limit}", public_controller.GetJobs)
			r.Get("/limit/{limit}/before/{job_id}", public_controller.GetJobs)

			r.Get("/{job_id}", public_controller.GetJob)
			r.Post("/", public_controller.PostJob)
			r.Delete("/{job_id}", public_controller.CancelJob)
		})
	})

	// Start server
	var PORT string
	if PORT = os.Getenv("PORT"); PORT == "" {
		PORT = "3000"
	}

	log.Infof("Starting server at port %v", PORT)
	http.ListenAndServe(":"+PORT, router)
}
