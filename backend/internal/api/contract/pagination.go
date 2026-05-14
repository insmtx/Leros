package contract

import (
	apiobj "github.com/ygpkg/yg-go/apis/apiobj"
)

// Pagination pagination request base
type Pagination struct {
	Offset  int `json:"offset,omitempty"`
	Limit   int `json:"limit,omitempty"`
	ListAll bool `json:"list_all,omitempty"`
}

// Fill 设置分页默认值，行为同 apiobj.PageQuery.Fill
func (p *Pagination) Fill() {
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Limit <= 0 || p.Limit > apiobj.PageMaxCount {
		p.Limit = apiobj.DefaultPageSize
	}
	if p.ListAll {
		p.Limit = apiobj.PageMaxCount
	}
}
