package policytemplate

import "regexp"

// (?m)은 멀티라인 모드로 각 라인의 시작이 ^레 매칭되도록 처리
// general_violation 등 violation을 포함하지만 violation이 아닌 정책을 매칭하지 않기 위해 멀티라인 모드 필요함
// OPA 포맷팅하면 violation rule은 공백없이 violation[ 으로 시작하므로 개행 문자 전까지 매칭
const violation_regex_pattern = `(?m)^violation\[[^\n\r]+[\n\r]+`

var violation_regex = regexp.MustCompile(violation_regex_pattern)

// violation 정책 헤드 매칭 후 다음에 삽입할 주석 및 가드 정책
const tks_guard_rego_rulename = `  # Do not delete following line, added by TKS
  ___not_tks_triggered_request___

`

// 가드 정책의 내용
// 해당 정책이 undefined로 빠지면 violation의 뒷 부분이 평가되지 않음
// 처음 블럭은 userInfo가 설정되지 않은 audit 모드에서 정책 평가가 스킵되는 것을 방지하기 위한처리
// 그 다음 블럭은 username이 tks_users 목록에 없고, tks_groups와 groups의 교집합 크기가 0인 경우에 true이며 그 외는 undefined
// 죽 username 및 groups가 정의된 리스트와 매칭되는 것이 하나라도 있으면 정책이 undefined가 됨
const tks_guard_rego_rulelogic = `
# Do not delete or edit following rule, managed by TKS
___not_tks_triggered_request___ {
  not input.review.userInfo
} {
  tks_users := {"kubernetes-admin","system:serviceaccount:kube-system:argocd-manager"}
  tks_groups := {"system:masters"}

  not tks_users[input.review.userInfo.username]

  count({g |g := input.review.userInfo.groups[_]; tks_groups[g]}) == 0
}
# Do not delete or edit end`

// violation 정책에 가드 정책 추가
func AddTksGuardToRego(rego string) string {
	// 매칭되는 violation 정책의 바디 첫부분에 가드 정책 체크를 추가하고 rego 코드 맨 끝에 실제 코드 내용 추가함
	return violation_regex.ReplaceAllString(rego, `${0}`+tks_guard_rego_rulename) +
		tks_guard_rego_rulelogic
}
