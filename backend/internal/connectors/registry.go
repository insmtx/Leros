package connectors

import "github.com/gin-gonic/gin"

type Registry struct {
	connectors map[string]Connector
}

func NewRegistry() *Registry {
	return &Registry{
		connectors: map[string]Connector{},
	}
}

func (r *Registry) Register(c Connector) {
	r.connectors[c.ChannelCode()] = c
}

func (r *Registry) RegisterRoutes(router gin.IRouter) {
	for _, c := range r.connectors {
		c.RegisterRoutes(router)
	}
}
