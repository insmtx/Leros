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
	return Topic().Org(orgidStr).Worker(workeridStr).Task().Build(), nil
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
	return Topic().Org(orgidStr).Session(sessionid).Message().Stream().Build(), nil
}
