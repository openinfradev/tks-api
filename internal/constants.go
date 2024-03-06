package internal

import "time"

type ContextKey string

const ContextKeyRequestID ContextKey = "REQUEST_ID"

const (
	PasswordExpiredDuration = 30 * 24 * time.Hour
	EmailCodeExpireTime     = 5 * time.Minute
	API_VERSION             = "/1.0"
	API_PREFIX              = "/api"
	ADMINAPI_PREFIX         = "/admin"

	SYSTEM_API_VERSION = "/1.0"
	SYSTEM_API_PREFIX  = "/system-api"
)

// 일단 DB 로 데이터를 관리하지 않고, 하드코딩 처리함.
const SERVICE_LMA = `{
	"name": "Logging,Monitoring,Alerting",
	"type": "LMA",
	"applications": [
	  {
		"name": "thanos",
		"version": "0.30.2",
		"description": "다중클러스터의 모니터링 데이터 통합 질의처리"
	  },
	  {
		"name": "prometheus-stack",
		"version": "v0.66.0",
		"description": "모니터링 데이터 수집/저장 및 질의처리"
	  },
	  {
		"name": "alertmanager",
		"version": "v0.25.0",
		"description": "알람 처리를 위한 노티피케이션 서비스"
	  },
	  {
		"name": "loki",
		"version": "2.6.1",
		"description": "로그데이터 저장 및 질의처리"
	  },
	  {
		"name": "grafana",
		"version": "8.3.3",
		"description": "모니터링/로그 통합대시보드"
	  }
	]
  }`

const SERVICE_SERVICE_MESH = `  {
  "name": "MSA",
  "type": "SERVICE_MESH",
  "applications": [
	{
	  "name": "istio",
	  "version": "v1.17.2",
	  "description": "MSA 플랫폼"
	},
	{
	  "name": "jagger",
	  "version": "1.35.0",
	  "description": "분산 서비스간 트랜잭션 추적을 위한 플랫폼"
	},
	{
	  "name": "kiali",
	  "version": "v1.63.0",
	  "description": "MSA 구조 및 성능을 볼 수 있는 Dashboard"
	},
	{
	  "name": "k8ssandra",
	  "version": "1.6.0",
	  "description": "분산 서비스간 호출 로그를 저장하는 스토리지"
	}
  ]
}`
