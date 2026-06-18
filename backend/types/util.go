package types

// AuthState 认证状态
type AuthState int

const (
	AuthStateNil    AuthState = 0
	AuthStateSucc   AuthState = 1
	AuthStateFailed AuthState = 2
)

// CallerKind identifies the authenticated principal type.
type CallerKind string

const (
	CallerKindUser   CallerKind = "user"
	CallerKindWorker CallerKind = "worker"
	CallerKindSystem CallerKind = "system"
)

// Caller 定义了一个执行身份，包含用户 ID、租户 ID 和认证状态。
type Caller struct {
	Uin      uint       `json:"uin"`
	OrgID    uint       `json:"org_id"`
	WorkerID uint       `json:"worker_id,omitempty"`
	Kind     CallerKind `json:"kind,omitempty"`
	State    AuthState  `json:"state"`
}

// Trace 定义了一个跟踪信息结构体，用于在请求链路中传递跟踪标识符，帮助进行分布式追踪和日志关联。
type Trace struct {
	RequestID string   `json:"request_id"`
	TraceID   string   `json:"trace_id"`
	SpanID    []string `json:"span_id"`
}

// IdentityContext 携带执行身份与跟踪信息。
type IdentityContext struct {
	Caller *Caller `json:"caller"`
	Trace  *Trace  `json:"trace"`
}

// SystemIdentity 返回一个预定义的系统身份，通常用于系统内部调用或没有特定用户上下文的场景。
func SystemIdentity() *Caller {
	return &Caller{
		Uin:   SystemUin,
		OrgID: SystemOrgID,
		Kind:  CallerKindSystem,
		State: AuthStateSucc,
	}
}

// 分页默认值
const (
	DefaultPageSize = 10
	PageMaxCount    = 150
)

// Pagination pagination request base
type Pagination struct {
	Offset  int  `json:"offset,omitempty"`
	Limit   int  `json:"limit,omitempty"`
	ListAll bool `json:"list_all,omitempty"`
}

// Fill 设置分页默认值
func (p *Pagination) Fill() {
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Limit <= 0 || p.Limit > PageMaxCount {
		p.Limit = DefaultPageSize
	}
	if p.ListAll {
		p.Limit = PageMaxCount
	}
}

// PageQuery 通用分页查询参数
type PageQuery struct {
	Filters    []Filter `json:"filters,omitempty"`
	OrderBy    []string `json:"order_by,omitempty"`
	Caller     `json:"-"`
	Pagination `json:",inline"`
}

// Filter 过滤条件
type Filter struct {
	Field      string   `json:"field"`
	Value      []string `json:"value"`
	ExactMatch bool     `json:"exact_match"`
}

// NewPageQuery 快速创建 PageQuery，接受 Caller 和分页参数。
func NewPageQuery(caller Caller, offset, limit int) *PageQuery {
	return &PageQuery{
		Caller:     caller,
		Pagination: Pagination{Offset: offset, Limit: limit},
	}
}

// AddFilter 链式添加过滤条件。
func (pq *PageQuery) AddFilter(field string, values ...string) *PageQuery {
	pq.Filters = append(pq.Filters, Filter{Field: field, Value: values})
	return pq
}

// AddExactFilter 链式添加精确过滤条件。
func (pq *PageQuery) AddExactFilter(field string, values ...string) *PageQuery {
	pq.Filters = append(pq.Filters, Filter{Field: field, Value: values, ExactMatch: true})
	return pq
}
