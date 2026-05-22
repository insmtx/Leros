package steps

import (
	"github.com/insmtx/Leros/backend/internal/agent"
)

func requestRunID(req *agent.RequestContext) string {
	if req == nil {
		return ""
	}
	return req.RunID
}

func requestTraceID(req *agent.RequestContext) string {
	if req == nil {
		return ""
	}
	return req.TraceID
}

func metadataFromResult(result *agent.RunResult) map[string]any {
	if result == nil || result.Metadata == nil {
		return nil
	}
	metadata := make(map[string]any, len(result.Metadata))
	for key, value := range result.Metadata {
		metadata[key] = value
	}
	return metadata
}
