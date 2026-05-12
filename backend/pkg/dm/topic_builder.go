package dm

import (
	"strings"
	"unicode"
)

const wildcard = "*"
const defaultSeparator = "."
const underscoreSeparator = "_"
const unknownSegment = "unknown"

// topicBuilder 基于有序片段构造领域消息 topic。
type topicBuilder struct {
	segments  []string
	separator string
}

// topic 创建一个空的 topic builder。
func topic() topicBuilder {
	return topicBuilder{separator: defaultSeparator}
}

// Add 追加一个或多个普通 topic 片段。
func (b topicBuilder) Add(segments ...string) topicBuilder {
	next := b.clone()
	for _, segment := range segments {
		next.segments = append(next.segments, cleanSegment(segment))
	}
	return next
}

// Org 追加组织字段片段。
func (b topicBuilder) Org(orgID string) topicBuilder {
	return b.Add("org", orgID)
}

// Session 追加会话字段片段。
func (b topicBuilder) Session(sessionID string) topicBuilder {
	return b.Add("session", sessionID)
}

// Worker 追加 Worker 字段片段。
func (b topicBuilder) Worker(workerID string) topicBuilder {
	return b.Add("worker", workerID)
}

// Message 追加 message 字段片段。
func (b topicBuilder) Message() topicBuilder {
	return b.Add("message")
}

// Stream 追加 stream 字段片段。
func (b topicBuilder) Stream() topicBuilder {
	return b.Add("stream")
}

// Task 追加 task 字段片段。
func (b topicBuilder) Task() topicBuilder {
	return b.Add("task")
}

// Wildcard 追加单层通配符片段。
func (b topicBuilder) Wildcard() topicBuilder {
	next := b.clone()
	next.segments = append(next.segments, wildcard)
	return next
}

// WithSeparator 返回使用指定连接符的新 topic builder。
func (b topicBuilder) WithSeparator(separator string) topicBuilder {
	next := b.clone()
	next.separator = separator
	return next
}

// WithUnderscoreSeparator 返回使用下划线连接符的新 topic builder。
func (b topicBuilder) WithUnderscoreSeparator() topicBuilder {
	return b.WithSeparator(underscoreSeparator)
}

// Build 返回使用当前连接符连接后的最终 topic，默认连接符为点号。
func (b topicBuilder) Build() string {
	separator := b.separator
	if separator == "" {
		separator = defaultSeparator
	}
	return strings.Join(b.segments, separator)
}

func (b topicBuilder) clone() topicBuilder {
	segments := make([]string, len(b.segments))
	copy(segments, b.segments)
	return topicBuilder{segments: segments, separator: b.separator}
}

func cleanSegment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		switch {
		case r == '.' || r == '*' || r == '>':
			return '_'
		case unicode.IsSpace(r):
			return '_'
		default:
			return r
		}
	}, value)
	if value == "" {
		return unknownSegment
	}
	return value
}
