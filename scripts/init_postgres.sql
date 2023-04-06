insert into roles ( id, name, description, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'tks-admin', 'tks-admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b2b689f0-ceeb-46c2-b280-0bc06896acd1', 'admin', 'admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'd3015140-2b12-487a-9516-cdeed7c17735', 'project-leader', 'project-leader', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'f6637d3d-3a0e-4db0-9086-c1b6dc9d433d', 'project-member', 'project-member', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b7ac7e7d-d8bc-470d-b6b2-3e0cc8ba55cc', 'project-viewer', 'project-viewer', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'ff4187a2-f3c1-46b3-8448-03a4b5e132e7', 'user', 'user', now(), now() );

insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'organization', 'organization', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'project', 'project', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'user', 'user', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'cluster', 'cluster', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'service', 'service', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'k8s_resources', 'resources of k8s', 'f', '', 'f', '', 'f', '', 'f', '', '', now(), now() );

insert into organizations ( id, name, description, created_at, updated_at ) values ( 'master', 'master', 'tks', now(), now() );
insert into users ( id, account_id, name, password, organization_id, role_id, created_at, updated_at  ) values ( 'bf67de40-ce15-4dc0-b6c2-17f053ca504f', 'admin', 'admin', '$2a$10$Akf03nbLHk93sTtozm35XuINXkJeNX7A1T9o/Pxpg9R2B2PToBPOO', 'master', 'b2b689f0-ceeb-46c2-b280-0bc06896acd1', now(), now() );

insert into cloud_settings ( id, name, description, organization_id, type, resource, created_at, updated_at ) 
values ( 'ce9e0387-01cb-4f37-a22a-fb91b6338434', 'aws', 'aws_description', 'master', 'aws', 'result', now(), now() );

insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, created_at, updated_at )
values ( '49901092-be76-4d4f-94e9-b84525f560b5', 'master', 'AWS Standard', 'included LMA over AWS', "v1", "AWS", "x86 | arm", "aws-reference", now(), now() )

