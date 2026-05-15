package dm

// 消费者视角

import (
	"errors"
	"fmt"
)

// WorkerTaskTopic 构造 worker 任务 topic，格式为 "org.{org_id}.worker.{worker_id}.task"。
func WorkerTaskTopic(orgid, workerid uint) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if workerid == 0 {
		return "", errors.New("workerid is required")
	}
	orgidStr := fmt.Sprintf("%d", orgid)
	workeridStr := fmt.Sprintf("%d", workerid)
	return topic().Org(orgidStr).Worker(workeridStr).Task().Build(), nil
}

// SessionResultStreamTopic 构造会话结果流 topic，格式为 "org.{org_id}.session.{session_id}.stream"。
func SessionResultStreamTopic(orgid uint, sessionid string) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if sessionid == "" {
		return "", errors.New("sessionid is required")
	}
	orgidStr := fmt.Sprintf("%d", orgid)
	return topic().Org(orgidStr).Session(sessionid).Message().Stream().Build(), nil
}

// SessionCompletedTopic 构造会话完成 topic，格式为 "org.{org_id}.session.{session_id}.completed"。
func SessionCompletedTopic(orgid uint, sessionid string) (string, error) {
	if orgid == 0 {
		return "", errors.New("orgid is required")
	}
	if sessionid == "" {
		return "", errors.New("sessionid is required")
	}
	orgidStr := fmt.Sprintf("%d", orgid)
	return topic().Org(orgidStr).Session(sessionid).Completed().Build(), nil
}

// SessionCompletedTopicWildcard 构造会话完成 topic 的通配符模式，格式为 "org.*.session.*.completed"。
func SessionCompletedTopicWildcard() string {
	return topic().Org(wildcard).Session(wildcard).Completed().Build()
}
