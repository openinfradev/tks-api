CREATE TABLE IF NOT EXISTS public.app_groups
(
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    app_group_type integer,
    cluster_id text ,
    name text ,
    description text ,
    workflow_id text ,
    status integer,
    status_desc text ,
    creator_id uuid,
    updator_id uuid,
    CONSTRAINT app_groups_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.app_serve_app_tasks
(
    id text NOT NULL,
    app_serve_app_id text  NOT NULL,
    version text ,
    status text ,
    output text ,
    artifact_url text ,
    image_url text ,
    executable_path text ,
    profile text ,
    app_config text ,
    app_secret text ,
    extra_env text ,
    port text ,
    resource_spec text ,
    helm_revision integer,
    strategy text ,
    pv_enabled boolean,
    pv_storage_class text ,
    pv_access_mode text ,
    pv_size text ,
    pv_mount_path text ,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT app_serve_app_tasks_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.app_serve_apps
(
    id text  NOT NULL,
    name text ,
    organization_id text ,
    type text ,
    app_type text ,
    endpoint_url text ,
    preview_endpoint_url text ,
    target_cluster_id text ,
    status text ,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT app_serve_apps_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.applications
(
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    app_group_id text ,
    endpoint text ,
    metadata jsonb,
    type integer,
    CONSTRAINT applications_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.cloud_accounts
(
    id text  NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id character varying(36) ,
    name text ,
    description text ,
    resource text ,
    cloud_service text ,
    creator_id uuid,
    updator_id uuid,
    CONSTRAINT cloud_accounts_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.cloud_settings
(
    id text  NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id character varying(36) ,
    name text ,
    description text ,
    resource text ,
    type text ,
    creator_id uuid,
    updator_id uuid,
    cloud_service text ,
    CONSTRAINT cloud_settings_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.clusters
(
    id text  NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text ,
    organization_id character varying(36) ,
    template_id text ,
    ssh_key_name text ,
    region text ,
    num_of_az bigint,
    machine_type text ,
    min_size_per_az bigint,
    max_size_per_az bigint,
    creator text ,
    description text ,
    workflow_id text ,
    status integer,
    status_desc text ,
    cloud_setting_id text ,
    creator_id uuid,
    updator_id uuid,
    stack_template_id text ,
    cloud_account_id text ,
    CONSTRAINT clusters_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.histories
(
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id text ,
    history_type text ,
    project_id text ,
    description text ,
    CONSTRAINT histories_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.organizations
(
    id character varying(36)  NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text ,
    description text ,
    phone_number text ,
    workflow_id text ,
    status integer,
    status_desc text ,
    creator text ,
    phone text ,
    primary_cluster_id text ,
    CONSTRAINT organizations_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.policies
(
    id bigint NOT NULL DEFAULT nextval('policies_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    role_id uuid,
    name text ,
    description text ,
    c boolean,
    create_priviledge text ,
    u boolean,
    update_priviledge text ,
    r boolean,
    read_priviledge text ,
    d boolean,
    delete_priviledge text ,
    creator text ,
    CONSTRAINT policies_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.roles
(
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text ,
    description text ,
    creator text ,
    CONSTRAINT roles_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.stack_templates
(
    id text  NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id character varying(36) ,
    name text ,
    description text ,
    version text ,
    cloud_service text ,
    platform text ,
    template text ,
    creator_id uuid,
    updator_id uuid,
    kube_version text ,
    kube_type text ,
    services jsonb,
    CONSTRAINT stack_templates_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.users
(
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id text ,
    name text ,
    password text ,
    auth_type text ,
    role_id uuid,
    organization_id character varying(36) ,
    creator text ,
    email_address text ,
    department text ,
    description text ,
    email text ,
    CONSTRAINT users_pkey PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.app_groups
    ADD CONSTRAINT fk_app_groups_creator FOREIGN KEY (creator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.app_groups
    ADD CONSTRAINT fk_app_groups_updator FOREIGN KEY (updator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.app_serve_app_tasks
    ADD CONSTRAINT fk_app_serve_apps_app_serve_app_tasks FOREIGN KEY (app_serve_app_id)
    REFERENCES public.app_serve_apps (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_accounts
    ADD CONSTRAINT fk_cloud_accounts_creator FOREIGN KEY (creator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_accounts
    ADD CONSTRAINT fk_cloud_accounts_organization FOREIGN KEY (organization_id)
    REFERENCES public.organizations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_accounts
    ADD CONSTRAINT fk_cloud_accounts_updator FOREIGN KEY (updator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_settings
    ADD CONSTRAINT fk_cloud_settings_creator FOREIGN KEY (creator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_settings
    ADD CONSTRAINT fk_cloud_settings_organization FOREIGN KEY (organization_id)
    REFERENCES public.organizations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.cloud_settings
    ADD CONSTRAINT fk_cloud_settings_updator FOREIGN KEY (updator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_cloud_account FOREIGN KEY (cloud_account_id)
    REFERENCES public.cloud_accounts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_cloud_setting FOREIGN KEY (cloud_setting_id)
    REFERENCES public.cloud_settings (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_creator FOREIGN KEY (creator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_organization FOREIGN KEY (organization_id)
    REFERENCES public.organizations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_stack_template FOREIGN KEY (stack_template_id)
    REFERENCES public.stack_templates (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.clusters
    ADD CONSTRAINT fk_clusters_updator FOREIGN KEY (updator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.policies
    ADD CONSTRAINT fk_policies_role FOREIGN KEY (role_id)
    REFERENCES public.roles (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.stack_templates
    ADD CONSTRAINT fk_stack_templates_creator FOREIGN KEY (creator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.stack_templates
    ADD CONSTRAINT fk_stack_templates_organization FOREIGN KEY (organization_id)
    REFERENCES public.organizations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.stack_templates
    ADD CONSTRAINT fk_stack_templates_updator FOREIGN KEY (updator_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.users
    ADD CONSTRAINT fk_users_organization FOREIGN KEY (organization_id)
    REFERENCES public.organizations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS public.users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id)
    REFERENCES public.roles (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


-- Project table Start
CREATE TABLE IF NOT EXISTS public.projects (
    id text primary key not null,
    organization_id text,
    name text,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE INDEX idx_projects_name ON projects USING btree (name);

CREATE TABLE IF NOT EXISTS public.project_members (
    id text primary key not null,
    project_id text not null,
    user_id text,
    project_role_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    foreign key (project_id) references public.projects (id)
    match simple on update no action on delete no action,
    foreign key (project_role_id) references public.project_roles (id)
    match simple on update no action on delete no action
);

CREATE TABLE IF NOT EXISTS public.project_namesapces (
    id text primary key not null,
    project_id text not null,
    stack_id text,
    namespace text,
    description text,
    status text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    foreign key (project_id) references public.projects (id)
    match simple on update no action on delete no action
);
CREATE UNIQUE INDEX idx_stackid_namespace ON project_namesapces USING btree (stack_id, namespace);

CREATE TABLE IF NOT EXISTS public.project_roles (
    id text primary key not null,
    name text,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
-- Project table End

END;