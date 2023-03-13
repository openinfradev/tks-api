insert into roles ( id, name, description, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'tks-admin', 'tks-admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b2b689f0-ceeb-46c2-b280-0bc06896acd1', 'admin', 'admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'd3015140-2b12-487a-9516-cdeed7c17735', 'project-leader', 'project-leader', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'f6637d3d-3a0e-4db0-9086-c1b6dc9d433d', 'project-member', 'project-member', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b7ac7e7d-d8bc-470d-b6b2-3e0cc8ba55cc', 'project-viewer', 'project-viewer', now(), now() );


insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'organization', 'organization', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'project', 'project', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'user', 'user', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'cluster', 'cluster', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'service', 'service', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values 
( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'k8s_resources', 'resources of k8s', 'f', '', 'f', '', 'f', '', 'f', '', '', now(), now() );


insert into users ( id, account_id, name, password, created_at, updated_at  ) values ( 'bf67de40-ce15-4dc0-b6c2-17f053ca504f', 'admin', 'admin', '$2a$10$zNkTnh/luJXeV6uJIQNnxOuJEfuQ4r3x6XZE27RUvYKjoPGy6.I2q<nil>', now(), now() );

insert into user_roles ( user_id, role_id) values ( 'bf67de40-ce15-4dc0-b6c2-17f053ca504f', '2ea4415c-9748-493f-91ba-4a64506b7be8');

#insert into organization_users ( organization_id, user_id) values ( 'pqrjjzko8', 'bf67de40-ce15-4dc0-b6c2-17f053ca504f');
