# organization 에 신규로 추가된 admin_id 를 일괄 업데이트 하는 쿼리
# v.4.0.0 deploy 후 아래 쿼리를 실행할 것

UPDATE organizations AS a
SET admin_id = b.id
FROM users b
WHERE b.account_id = 'admin' AND a.id = b.organization_id