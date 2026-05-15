package dm

// 消费者视角

import (
	"errors"
	"fmt"
)

// WorkerTaskSubject 构造 worker 任务 topic，格式为 "org.{org_id}.worker.{worker_id}.task"。
func WorkerTaskSubject(orgid, workerid uint) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if workerid == 0 {
		return "", errors.New("workerid is required")
	}
	workeridStr := fmt.Sprintf("%d", workerid)
	return topic().Org(orgid).Worker(workeridStr).Task().Build(), nil
}

// SessionResultStreamSubject 构造会话结果流 topic，格式为 "org.{org_id}.session.{session_id}.stream"。
func SessionResultStreamSubject(orgid uint, sessionid string) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if sessionid == "" {
		return "", errors.New("sessionid is required")
	}
	return topic().Org(orgid).Session(sessionid).Message().Stream().Build(), nil
}

// SessionCompletedSubject 构造会话完成 topic，格式为 "org.{org_id}.session.{session_id}.completed"。
func SessionCompletedSubject(orgid uint, sessionid string) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if sessionid == "" {
		return "", errors.New("sessionid is required")
	}
	return topic().Org(orgid).Session(sessionid).Completed().Build(), nil
}

// SessionCompletedWildcardSubject 构造会话完成 topic 的通配符模式，格式为 "org.*.session.*.completed"。
func SessionCompletedWildcardSubject() string {
	return "org.*.session.*.completed"
}
