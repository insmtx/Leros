package dm

import "strings"

const (
	streamNameTask    = "TASK_STREAM"
	streamNameSession = "SESSION_STREAM"
)

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
