package dm

import "strings"

const (
	streamNameTask    = "TASK_STREAM"
	streamNameSession = "SESSION_STREAM"
)

// StreamSubjects 定义各 Stream 的 NATS subject 匹配模式，使用通配符覆盖所有动态 topic。
var StreamSubjects = map[string][]string{
	streamNameTask:    {"org.*.worker.*.task"},
	streamNameSession: {"org.*.session.*.message.*"},
}

func SessionStream() string {
	return streamNameSession
}

// StreamNameFromTopic 根据 topic 返回 Stream 名称。
func StreamNameFromTopic(topic string) string {
	parts := strings.SplitN(topic, ".", 4)
	if len(parts) < 4 {
		return ""
	}
	subject := parts[2]
	switch subject {
	case "worker":
		return streamNameTask
	case "session":
		return streamNameSession
	default:
		return ""
	}
}
