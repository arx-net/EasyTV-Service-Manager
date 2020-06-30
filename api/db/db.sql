--
--  Configuration tables
--

create table if not exists module (
    id serial primary key not null,
    name varchar unique not null,
    description varchar not null,
    api_key varchar unique not null,
    enabled boolean not null
)

create table if not exists tasks (
    id serial primary key not null,
    module_id serial references module(id) not null,
    name varchar not null,
    description varchar not null,
    start_url varchar not null,
    cancel_url varchar not null,
    enabled boolean not null,
    deleted boolean not null
)

create table if not exists task_parameter (
    id serial primary key not null,
    task_id serial references task(id) not null,
    name varchar not null,
    data_type smallint not null,
    is_input boolean not null
)

--
--  Content Owner and more data
--
create table if not exists content_owner (
    id serial primary key not null,
    username varchar unique not null,
    password varchar not null,
    email varchar unique not null,
    name varchar not null
)

create table if not exists job (
    id serial primary key not null,
    is_completed boolean not null,
    is_canceled boolean not null,
    creation_date timestamp not null,
    completion_date timestamp,
    publication_date timestamp,
    expiration_date timestamp,
    current_step smallint not null,
    owner_id serial references content_owner(id) not null,
    status varchar not null 
)

create table if not exists asset (
    id serial primary key not null,
    job_id serial references job(id) not null,
    url varchar unique not null
)

create table if not exists job_step (
    id serial primary key not null,
    job_id serial references job(id) not null,
    task_id serial references task(id) not null,
    next_step_id serial references job_sted(id),
    order smallint not null
)

create table if not exists job_param (
    id serial primary key not null,
    job_step_id serial references job_step(id) not null,
    param_id serial references task_parameter(id) not null,
    value varchar,
    linked_id serial references job_param(id) not null
)