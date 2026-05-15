// nats 提供基于 NATS JetStream 的事件总线实现
//
// 该部分实现了 mq 包中的 Publisher 和 Subscriber 接口，
// 使用 NATS JetStream 作为消息中间件来实现事件的发布和订阅。
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

// natsPublisher 表示一个 NATS 客户端，实现 Publisher 和 Subscriber 接口
type natsPublisher struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	closed bool
	mu     sync.Mutex
}

const defaultRealtimeFlushTimeout = 5 * time.Second

// NewPublisher 创建一个新的 NATS JetStream 发布者实例
// 在初始化阶段创建所有预配置的 Streams
func NewPublisher(url string) (*natsPublisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		logs.Errorf("Failed to connect to NATS: %v", err)
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		logs.Errorf("Failed to create JetStream context: %v", err)
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	publisher := &natsPublisher{
		conn:   conn,
		js:     js,
		closed: false,
	}

	if err := publisher.initStreams(); err != nil {
		conn.Close()
		logs.Errorf("Failed to initialize NATS streams: %v", err)
		return nil, fmt.Errorf("failed to initialize NATS streams: %w", err)
	}

	logs.Infof("Successfully connected to NATS at %s with JetStream", url)
	return publisher, nil
}

// initStreams 在初始化阶段创建或更新所有预配置的 Stream
func (p *natsPublisher) initStreams() error {
	// 先清理可能冲突的旧 streams
	existingStreamsCh := p.js.StreamNames()
	var existingStreams []string
	for name := range existingStreamsCh {
		existingStreams = append(existingStreams, name)
	}
	for _, name := range existingStreams {
		info, err := p.js.StreamInfo(name)
		if err != nil {
			continue
		}
		// 检查是否与预配置的 stream 冲突（同 subjects 但不同名）
		isConfigured := false
		for streamName, subjects := range dm.StreamSubjects {
			if name == streamName {
				isConfigured = true
				break
			}
			// 判断 subjects 是否重叠
			if hasOverlap(info.Config.Subjects, subjects) {
				logs.Warnf("Deleting conflicting stream '%s' (subjects: %v)", name, info.Config.Subjects)
				if err := p.js.DeleteStream(name); err != nil {
					logs.Warnf("Failed to delete conflicting stream '%s': %v", name, err)
				}
			}
		}
		// 如果 stream 已存在但不是我们配置的，检查其 subjects
		if !isConfigured {
			for _, subj := range info.Config.Subjects {
				for _, cfgSubjects := range dm.StreamSubjects {
					if hasOverlap([]string{subj}, cfgSubjects) {
						logs.Warnf("Deleting conflicting stream '%s' with subject %q", name, subj)
						_ = p.js.DeleteStream(name)
						break
					}
				}
			}
		}
	}

	for streamName, subjects := range dm.StreamSubjects {
		_, addErr := p.js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: subjects,
			Storage:  nats.FileStorage,
		})
		if addErr == nil {
			logs.Infof("Created JetStream stream '%s' with subjects: %v", streamName, subjects)
			continue
		}

		// AddStream 失败，尝试 UpdateStream
		if _, err := p.js.UpdateStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: subjects,
			Storage:  nats.FileStorage,
		}); err != nil {
			return fmt.Errorf("failed to initialize stream '%s': AddStream=%v, UpdateStream=%w", streamName, addErr, err)
		}
		logs.Infof("Updated JetStream stream '%s' with subjects: %v", streamName, subjects)
	}
	return nil
}

// hasOverlap 检查两组 subjects 是否有重叠
func hasOverlap(a, b []string) bool {
	for _, s1 := range a {
		for _, s2 := range b {
			if subjectsOverlap(s1, s2) {
				return true
			}
		}
	}
	return false
}

// subjectsOverlap 检查两个 NATS subject 模式是否重叠
func subjectsOverlap(s1, s2 string) bool {
	if s1 == s2 {
		return true
	}
	p1 := strings.Split(s1, ".")
	p2 := strings.Split(s2, ".")

	// 不同长度的固定部分不重叠
	if len(p1) != len(p2) {
		// 通配符情况: org.*.worker.*.task 与 org.1.worker.2.task 重叠
		return partialMatch(p1, p2)
	}

	for i := range p1 {
		if p1[i] == "*" || p2[i] == "*" {
			continue
		}
		if p1[i] == ">" || p2[i] == ">" {
			return true
		}
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}

// partialMatch 检查不同长度的 subject 是否可能重叠（通配符场景）
func partialMatch(p1, p2 []string) bool {
	// 处理 > 通配符
	for i, s := range p1 {
		if s == ">" {
			return true
		}
		if i < len(p2) && p1[i] != "*" && p2[i] != p1[i] {
			return false
		}
	}
	for i, s := range p2 {
		if s == ">" {
			return true
		}
		if i < len(p1) && p2[i] != "*" && p1[i] != p2[i] {
			return false
		}
	}
	// 短的是长的前缀且短的以 * 结尾可能重叠
	return false
}

// PublishWithContext 在给定上下文环境中发布消息到指定主题
func (p *natsPublisher) PublishWithContext(ctx context.Context, topic string, message any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("NATS client is closed")
	}

	// 将消息序列化为 JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	_, err = p.js.Publish(topic, body, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to publish message to topic '%s': %w", topic, err)
	}

	return nil
}

// PublishRealtimeWithContext 在给定上下文环境中发布实时消息，不声明 JetStream Stream。
func (p *natsPublisher) PublishRealtimeWithContext(ctx context.Context, topic string, message any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("NATS client is closed")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.conn.Publish(topic, body); err != nil {
		return fmt.Errorf("failed to publish message to subject '%s': %w", topic, err)
	}

	return nil
}

// SubscribeWithContext 在给定上下文环境中订阅特定主题的消息。
// 该函数会阻塞直到 context 被取消或订阅返回错误。
func (p *natsPublisher) SubscribeWithContext(ctx context.Context, topic string, handler func(msg *nats.Msg)) error {
	// 使用 OrderedConsumer 不依赖 durable consumer，避免重启时 "consumer already bound" 错误。
	// OrderedConsumer 使用 AckNone 策略，无需手动 ack。
	sub, err := p.js.Subscribe(topic, func(msg *nats.Msg) {
		handler(msg)
	}, nats.OrderedConsumer(), nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic '%s': %w", topic, err)
	}

	// 阻塞直到 context 取消。
	<-ctx.Done()

	if err := sub.Unsubscribe(); err != nil {
		logs.WarnContextf(ctx, "Failed to unsubscribe from topic '%s': %v", topic, err)
	}
	logs.InfoContextf(ctx, "Unsubscribed from topic: %s", topic)

	return ctx.Err()
}

// Publish implements the eventbus.Publisher interface
func (p *natsPublisher) Publish(ctx context.Context, topic string, event any) error {
	return p.PublishWithContext(ctx, topic, event)
}

// PublishRealtime implements the eventbus.RealtimePublisher interface.
func (p *natsPublisher) PublishRealtime(ctx context.Context, topic string, event any) error {
	return p.PublishRealtimeWithContext(ctx, topic, event)
}

// FlushRealtime implements the eventbus.RealtimePublisher interface.
func (p *natsPublisher) FlushRealtime(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("NATS client is closed")
	}
	p.mu.Unlock()

	flushCtx, cancel := contextWithDefaultDeadline(ctx, defaultRealtimeFlushTimeout)
	defer cancel()
	if err := p.conn.FlushWithContext(flushCtx); err != nil {
		return fmt.Errorf("failed to flush realtime messages: %w", err)
	}

	return nil
}

// Subscribe implements the eventbus.Subscriber interface
func (p *natsPublisher) Subscribe(ctx context.Context, topic string, handler func(msg *nats.Msg)) error {
	return p.SubscribeWithContext(ctx, topic, handler)
}

// SubscribeRealtime implements the eventbus.RealtimeSubscriber interface
func (p *natsPublisher) SubscribeRealtime(ctx context.Context, topic string, handler func(msg *nats.Msg)) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("NATS client is closed")
	}
	p.mu.Unlock()

	// 使用 Core NATS 订阅，不依赖 JetStream
	sub, err := p.conn.Subscribe(topic, func(msg *nats.Msg) {
		handler(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to realtime topic '%s': %w", topic, err)
	}

	// 在 context 取消时清理订阅
	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			logs.WarnContextf(ctx, "Failed to unsubscribe from realtime topic '%s': %v", topic, err)
		}
		logs.InfoContextf(ctx, "Unsubscribed from realtime topic: %s", topic)
	}()

	return nil
}

// Close 关闭 NATS 连接并释放资源
func (p *natsPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	p.conn.Close()

	return nil
}

func contextWithDefaultDeadline(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), timeout)
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
