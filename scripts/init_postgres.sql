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

insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '49901092-be76-4d4f-94e9-b84525f560b5', 'master', 'AWS Standard (x86)', 'included LMA', 'v1', 'AWS', 'x86', 'aws-reference', 'STANDARD', 'v1.25', 'AWS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '44d5e76b-63db-4dd0-a16e-11bd3f6054cf', 'master', 'AWS MSA Standard (x86)', 'included LMA, SERVICE MESH', 'v1', 'AWS', 'x86', 'aws-msa-reference', 'MSA', 'v1.25', 'AWS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}, {"name": "MSA", "type": "SERVICE_MESH", "applications": [{"name": "istio", "version": "v1.13.1", "description": "MSA 플랫폼"}, {"name": "jagger", "version": "v2.27.1", "description": "분산 서비스간 트랜잭션 추적을 위한 로깅 플랫폼"}, {"name": "kiali", "version": "v1.45.1", "description": "MSA 통합 모니터링포탈"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( 'fe1d97e0-7428-4be6-9c69-310a88b4ff46', 'master', 'AWS Standard (arm)', 'included LMA', 'v2', 'AWS', 'arm', 'aws-arm-reference', 'STANDARD', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '3696cb38-4da0-4235-97eb-b6eb15962bd1', 'master', 'AWS Standard (arm)', 'included LMA, SERVICE_MESH', 'v2', 'AWS', 'arm', 'aws-arm-msa-reference', 'MSA', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}, {"name": "MSA", "type": "SERVICE_MESH", "applications": [{"name": "istio", "version": "v1.13.1", "description": "MSA 플랫폼"}, {"name": "jagger", "version": "v2.27.1", "description": "분산 서비스간 트랜잭션 추적을 위한 로깅 플랫폼"}, {"name": "kiali", "version": "v1.45.1", "description": "MSA 통합 모니터링포탈"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( 'c8a4658d-d5a6-4191-8a91-e26f6aee007f', 'master', 'EKS Standard (x86)', 'included LMA', 'v1', 'AWS', 'x86', 'eks-reference', 'STANDARD', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '39f18a09-5b94-4772-bdba-e4c32ee002f7', 'master', 'EKS MSA Standard (x86)', 'included LMA, SERVICE MESH', 'v1', 'AWS', 'x86', 'eks-msa-reference', 'MSA', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]},{"name":"MSA","type":"SERVICE_MESH","applications":[{"name":"istio","version":"v1.17.2","description":"MSA 플랫폼"},{"name":"jagger","version":"1.35.0","description":"분산 서비스간 트랜잭션 추적을 위한 플랫폼"},{"name":"kiali","version":"v1.63.0","description":"MSA 구조 및 성능을 볼 수 있는 Dashboard"},{"name":"k8ssandra","version":"1.6.0","description":"분산 서비스간 호출 로그를 저장하는 스토리지"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '5678bf11-256f-4d2c-a673-f2fedb82de5b', 'master', 'BYOH Standard', 'included LMA', 'v1', 'AWS', 'x86', 'eks-reference', 'STANDARD', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '92f5e5ce-7ffd-4c3e-aff6-9b7fb03dd881', 'master', 'BYOH MSA Standard', 'included LMA, SERVICE MESH', 'v1', 'AWS', 'x86', 'eks-msa-reference', 'MSA', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]},{"name":"MSA","type":"SERVICE_MESH","applications":[{"name":"istio","version":"v1.17.2","description":"MSA 플랫폼"},{"name":"jagger","version":"1.35.0","description":"분산 서비스간 트랜잭션 추적을 위한 플랫폼"},{"name":"kiali","version":"v1.63.0","description":"MSA 구조 및 성능을 볼 수 있는 Dashboard"},{"name":"k8ssandra","version":"1.6.0","description":"분산 서비스간 호출 로그를 저장하는 스토리지"}]}]' );

insert into stack_templates ( id, organization_id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( 'f4c74dec-3444-481b-9597-bacb7596e7e7', 'master', 'BYOH MSA Standard (BTV STG)', 'included LMA', 'v1', 'BYOH', 'x86', 'byoh-stage-reference', 'MSA', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );

