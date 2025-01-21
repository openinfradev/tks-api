ALTER DATABASE tks SET timezone = 'Asia/Seoul';

## Roles
insert into roles ( id, name, description, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'tks-admin', 'tks-admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b2b689f0-ceeb-46c2-b280-0bc06896acd1', 'admin', 'admin', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'd3015140-2b12-487a-9516-cdeed7c17735', 'project-leader', 'project-leader', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'f6637d3d-3a0e-4db0-9086-c1b6dc9d433d', 'project-member', 'project-member', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'b7ac7e7d-d8bc-470d-b6b2-3e0cc8ba55cc', 'project-viewer', 'project-viewer', now(), now() );
insert into roles ( id, name, description, created_at, updated_at ) values ( 'ff4187a2-f3c1-46b3-8448-03a4b5e132e7', 'user', 'user', now(), now() );

## Policies
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'organization', 'organization', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'project', 'project', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'user', 'user', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'cluster', 'cluster', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'service', 'service', 't', '', 't', '', 't', '', 't', '', '', now(), now() );
insert into policies ( role_id, name, description, c, create_priviledge, u, update_priviledge, r, read_priviledge, d, delete_priviledge, creator, created_at, updated_at ) values ( '2ea4415c-9748-493f-91ba-4a64506b7be8', 'k8s_resources', 'resources of k8s', 'f', '', 'f', '', 'f', '', 'f', '', '', now(), now() );

## Organizations
insert into organizations ( id, name, description, created_at, updated_at ) values ( 'master', 'master', 'tks', now(), now() );

## Users
insert into users ( id, account_id, name, organization_id, created_at, updated_at  ) values ( 'bf67de40-ce15-4dc0-b6c2-17f053ca504f', 'admin', 'admin', 'master', now(), now() );

## StackTemplates
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '49901092-be76-4d4f-94e9-b84525f560b5', 'AWS Standard (x86)', 'included LMA', 'v1', 'AWS', 'x86', 'aws-reference', 'STANDARD', 'v1.25', 'AWS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '44d5e76b-63db-4dd0-a16e-11bd3f6054cf', 'AWS MSA Standard (x86)', 'included LMA, SERVICE MESH', 'v1', 'AWS', 'x86', 'aws-msa-reference', 'MSA', 'v1.25', 'AWS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}, {"name": "MSA", "type": "SERVICE_MESH", "applications": [{"name": "istio", "version": "v1.13.1", "description": "MSA 플랫폼"}, {"name": "jagger", "version": "v2.27.1", "description": "분산 서비스간 트랜잭션 추적을 위한 로깅 플랫폼"}, {"name": "kiali", "version": "v1.45.1", "description": "MSA 통합 모니터링포탈"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( 'fe1d97e0-7428-4be6-9c69-310a88b4ff46', 'AWS Standard (arm)', 'included LMA', 'v2', 'AWS', 'arm', 'aws-arm-reference', 'STANDARD', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '3696cb38-4da0-4235-97eb-b6eb15962bd1', 'AWS Standard (arm)', 'included LMA, SERVICE_MESH', 'v2', 'AWS', 'arm', 'aws-arm-msa-reference', 'MSA', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}, {"name": "MSA", "type": "SERVICE_MESH", "applications": [{"name": "istio", "version": "v1.13.1", "description": "MSA 플랫폼"}, {"name": "jagger", "version": "v2.27.1", "description": "분산 서비스간 트랜잭션 추적을 위한 로깅 플랫폼"}, {"name": "kiali", "version": "v1.45.1", "description": "MSA 통합 모니터링포탈"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( 'c8a4658d-d5a6-4191-8a91-e26f6aee007f', 'EKS Standard (x86)', 'included LMA', 'v1', 'AWS', 'x86', 'eks-reference', 'STANDARD', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '39f18a09-5b94-4772-bdba-e4c32ee002f7', 'EKS MSA Standard (x86)', 'included LMA, SERVICE MESH', 'v1', 'AWS', 'x86', 'eks-msa-reference', 'MSA', 'v1.25', 'AWS', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]},{"name":"MSA","type":"SERVICE_MESH","applications":[{"name":"istio","version":"v1.17.2","description":"MSA 플랫폼"},{"name":"jagger","version":"1.35.0","description":"분산 서비스간 트랜잭션 추적을 위한 플랫폼"},{"name":"kiali","version":"v1.63.0","description":"MSA 구조 및 성능을 볼 수 있는 Dashboard"},{"name":"k8ssandra","version":"1.6.0","description":"분산 서비스간 호출 로그를 저장하는 스토리지"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '5678bf11-256f-4d2c-a673-f2fedb82de5b', 'BYOH Standard', 'included LMA', 'v1', 'BYOH', 'x86', 'byoh-reference', 'STANDARD', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
values ( '92f5e5ce-7ffd-4c3e-aff6-9b7fb03dd881', 'BYOH MSA Standard', 'included LMA, SERVICE MESH', 'v1', 'BYOH', 'x86', 'byoh-msa-reference', 'MSA', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]},{"name":"MSA","type":"SERVICE_MESH","applications":[{"name":"istio","version":"v1.17.2","description":"MSA 플랫폼"},{"name":"jagger","version":"1.35.0","description":"분산 서비스간 트랜잭션 추적을 위한 플랫폼"},{"name":"kiali","version":"v1.63.0","description":"MSA 구조 및 성능을 볼 수 있는 Dashboard"},{"name":"k8ssandra","version":"1.6.0","description":"분산 서비스간 호출 로그를 저장하는 스토리지"}]}]' );
# BTV
#insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
#values ( '2526ec49-28a2-4be9-8d18-2c39fc0993fd', 'BYOH Admin Standard (BTV)', 'included LMA', 'v1', 'BYOH', 'x86', 'tks-admin', 'STANDARD', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
#insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
#values ( 'a76b5c97-7d55-46d8-9248-9952bfaff62c', 'BYOH MSA Standard (BTV SSU)', 'included LMA', 'v1', 'BYOH', 'x86', 'byoh-ssu-reference', 'MSA', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
#insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
#values ( 'b5bbd6ea-5bf3-4d88-bb06-4a4c64c73c15', 'BYOH MSA Standard (BTV SUY)', 'included LMA', 'v1', 'BYOH', 'x86', 'byoh-suy-reference', 'MSA', 'v1.25', 'BYOH', now(), now(), '[{"name":"Logging,Monitoring,Alerting","type":"LMA","applications":[{"name":"thanos","version":"0.30.2","description":"다중클러스터의 모니터링 데이터 통합 질의처리"},{"name":"prometheus-stack","version":"v0.66.0","description":"모니터링 데이터 수집/저장 및 질의처리"},{"name":"alertmanager","version":"v0.25.0","description":"알람 처리를 위한 노티피케이션 서비스"},{"name":"loki","version":"2.6.1","description":"로그데이터 저장 및 질의처리"},{"name":"grafana","version":"8.3.3","description":"모니터링/로그 통합대시보드"}]}]' );
# PSNM
#insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
#values ( 'c3396c68-03ec-4d41-991c-69e4a2ac16aa', 'psnm-backend-reference', 'included LMA', 'v1', 'AWS', 'x86', 'psnm-backend-reference', 'STANDARD', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );
#insert into stack_templates ( id, name, description, version, cloud_service, platform, template, template_type, kube_version, kube_type, created_at, updated_at, services )
#values ( '23b07a65-1cb3-4609-9bba-e88c15e2e192', 'psnm-frontend-reference', 'included LMA', 'v1', 'AWS', 'x86', 'psnm-frontend-reference', 'STANDARD', 'v1.25', 'EKS', now(), now(), '[{"name": "Logging,Monitoring,Alerting", "type": "LMA", "applications": [{"name": "prometheus-stack", "version": "v.44.3.1", "description": "통계데이터 제공을 위한 backend  플랫폼"}, {"name": "elastic-system", "version": "v1.8.0", "description": "로그 데이터 적재를 위한 Storage"}, {"name": "alertmanager", "version": "v0.23.0", "description": "Alert 관리를 위한 backend 서비스"}, {"name": "grafana", "version": "v6.50.7", "description": "모니터링 통합 포탈"}]}]' );

## Projects
insert into project_roles ( id, name, description, created_at, updated_at ) values ( 'f4358b4e-adc3-447a-8ad9-c111c4b9a974', 'project-leader', 'project-leader', now(), now() );
insert into project_roles ( id, name, description, created_at, updated_at ) values ( '2071bd6f-26b3-4c1a-a3ab-439bc89f0011', 'project-member', 'project-member', now(), now() );
insert into project_roles ( id, name, description, created_at, updated_at ) values ( 'f62c16e1-316c-4d7f-9cfa-dbe4ed7dfa17', 'project-viewer', 'project-viewer', now(), now() );

## SystemNotificationTemplates
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('d42d716f-dd2e-429b-897d-b602f6382790', 'node-cpu-high-load', 'node-cpu-high-load', true, 'SYSTEM_NOTIFICATION', '(avg by (taco_cluster, instance) (rate(node_cpu_seconds_total{mode="idle"}[60s])))', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('f11eefa4-5a16-44fc-8dae-4662e7fba023', 'node-memory-high-utilization', 'node-memory-high-utilization', true, 'SYSTEM_NOTIFICATION', '(node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes)', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('1ec08b58-2fe1-49c5-bbab-3544ec8ce330', 'node-disk-full', 'node-disk-full', true, 'SYSTEM_NOTIFICATION', 'predict_linear(node_filesystem_free_bytes{mountpoint="/"}[6h], 24*3600)', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('68dcb92d-91cc-47d0-9b2f-2285d74f157f', 'pvc-full', 'pvc-full', true, 'SYSTEM_NOTIFICATION','predict_linear(kubelet_volume_stats_available_bytes[6h], 24*3600)', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('46e9e216-364a-4a3f-9182-85b2c4c34f77', 'pod-restart-frequently', 'pod-restart-frequently', true, 'SYSTEM_NOTIFICATION','increase(kube_pod_container_status_restarts_total{namespace!="kube-system"}[60m:])', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('7355d0f9-7c14-4f70-92ea-a9868624ff82', 'policy-warning', 'policy-warning', true, 'POLICY_NOTIFICATION', 'opa_scorecard_constraint_violations{namespace!="kube-system|taco-system|gatekeeper-system", violation_enforcement="warn"}', null, null, now(), now() );
insert into system_notification_templates ( id, name, description, is_system, notification_type, metric_query, creator_id, updator_id, created_at, updated_at )
values ('792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'policy-blocked', 'policy-blocked', true, 'POLICY_NOTIFICATION', 'opa_scorecard_constraint_violations{namespace!="kube-system|taco-system|gatekeeper-system",violation_enforcement=""}', null, null, now(), now() );

## SystemNotificationTemplates -> SystemNotificationMetricParameters
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, 'd42d716f-dd2e-429b-897d-b602f6382790', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, 'd42d716f-dd2e-429b-897d-b602f6382790', 'INSTANCE', '$labels.instance', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, 'f11eefa4-5a16-44fc-8dae-4662e7fba023', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, 'f11eefa4-5a16-44fc-8dae-4662e7fba023', 'INSTANCE', '$labels.instance', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, '1ec08b58-2fe1-49c5-bbab-3544ec8ce330', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, '1ec08b58-2fe1-49c5-bbab-3544ec8ce330', 'INSTANCE', '$labels.instance', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, '68dcb92d-91cc-47d0-9b2f-2285d74f157f', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, '68dcb92d-91cc-47d0-9b2f-2285d74f157f', 'PVC', '$labels.persistentvolumeclaim', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, '46e9e216-364a-4a3f-9182-85b2c4c34f77', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, '46e9e216-364a-4a3f-9182-85b2c4c34f77', 'POD', '$labels.pod', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 2, '46e9e216-364a-4a3f-9182-85b2c4c34f77', 'NAMESPACE', '$labels.namespace', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'NAME', '$labels.name', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 2, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'KIND', '$labels.kind', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 3, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'VIOLATING_KIND', '$labels.violating_kind', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 4, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'VIOLATING_NAMESPACE', '$labels.violating_namespace', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 5, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'VIOLATING_NAME', '$labels.violating_name', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 6, '7355d0f9-7c14-4f70-92ea-a9868624ff82', 'VIOLATION_MSG', '$labels.violation_msg', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 0, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'STACK', '$labels.taco_cluster', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 1, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'NAME', '$labels.name', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 2, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'KIND', '$labels.kind', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 3, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'VIOLATING_KIND', '$labels.violating_kind', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 4, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'VIOLATING_NAMESPACE', '$labels.violating_namespace', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 5, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'VIOLATING_NAME', '$labels.violating_name', now(), now() );
insert into system_notification_metric_parameters ( "order", system_notification_template_id, key, value, created_at, updated_at )
values ( 6, '792ca0c6-b98f-4493-aa17-548de9eb9a4e', 'VIOLATION_MSG', '$labels.violation_msg', now(), now() );
