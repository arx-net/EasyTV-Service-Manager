package main

import (
	"fmt"

	"gitlab.arx.net/easytv/sm"

	"gitlab.arx.net/easytv/sm/db"
)

func init_db(pool *db.DatabasePool) {
	pool.DB.Query("DROP TABLE IF EXISTS admin_user;")
	pool.DB.Query("DROP TABLE IF EXISTS job_param;")
	pool.DB.Query("DROP TABLE IF EXISTS job_step;")
	pool.DB.Query("DROP TABLE IF EXISTS asset;")
	pool.DB.Query("DROP TABLE IF EXISTS job;")
	pool.DB.Query("DROP TABLE IF EXISTS content_owner;")
	pool.DB.Query("DROP TABLE IF EXISTS task_parameter;")
	pool.DB.Query("DROP TABLE IF EXISTS task;")
	pool.DB.Query("DROP TABLE IF EXISTS module;")

	create_table("AdminUser", `
	create table if not exists admin_user (
		id serial primary key not null,
		username varchar unique not null,
		password varchar unique not null
	)
	`, pool.DB)

	create_table("Module", `
		create table if not exists module (
			id serial primary key not null,
			name varchar unique not null,
			description varchar not null,
			api_key varchar unique not null,
			enabled boolean not null
		)`, pool.DB)

	create_table("ModuleIndex", `
		CREATE INDEX api_key_idx ON module (api_key)
		`, pool.DB)

	create_table("Task", `
		create table if not exists task (
			id serial primary key not null,
			module_id serial references module(id) not null,
			name varchar not null,
			description varchar not null,
			start_url varchar not null,
			cancel_url varchar not null,
			enabled boolean not null,
			deleted boolean not null
		)`, pool.DB)

	create_table("TaskParameter", `
		create table if not exists task_parameter (
			id serial primary key not null,
			task_id serial references task(id) not null,
			name varchar not null,
			data_type smallint not null,
			is_input boolean not null
		)`, pool.DB)

	create_table("ContentOwner", `
		create table if not exists content_owner (
			id serial primary key not null,
			username varchar unique not null,
			password varchar not null,
			email varchar unique not null,
			name varchar unique not null
		)`, pool.DB)

	create_table("Job", `
		create table if not exists job (
			id serial primary key not null,
			is_completed boolean not null,
			is_canceled boolean not null,
			is_expiration_processed boolean not null,
			creation_date timestamp not null,
			completion_date timestamp,
			publication_date timestamp,
			expiration_date timestamp,
			current_step smallint not null,
			owner_id serial references content_owner(id) not null,
			status varchar not null
		)`, pool.DB)

	create_table("Asset", `
		create table if not exists asset (
			id serial primary key not null,
			url_param varchar unique not null,
			job_id serial references job(id) not null,
			path varchar unique not null,
			size int
		)`, pool.DB)

	create_table("JobStep", `
		create table if not exists job_step (
			id serial primary key not null,
			job_id serial references job(id) not null,
			task_id serial references task(id) not null,
			step_order integer not null
		)`, pool.DB)

	create_table("JobParam", `
		create table if not exists job_param (
			job_step_id serial references job_step(id) not null,
			name varchar not null,
			is_input boolean not null,
			data_type smallint not null,
			value text,
			linked_output_name varchar,
			primary key (job_step_id, name, is_input)
		)`, pool.DB)

	fmt.Println("Create admin user")
	service := sm.NewAdminService(&db.AdminRepository{Pool: pool})
	_, err := service.CreateAdminUser("admin", "admin")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("User created")
	}
}
