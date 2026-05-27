package steps

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/agent"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
	agentworkspace "github.com/insmtx/Leros/backend/internal/workspace"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
)

// ArtifactRecorder persists declared artifacts and returns public event payloads.
type ArtifactRecorder interface {
	Record(ctx context.Context, req *agent.RequestContext) ([]events.ArtifactPayload, error)
}

// ArtifactStep records manifest artifacts before the terminal run event is emitted.
type ArtifactStep struct {
	Recorder ArtifactRecorder
}

func (ArtifactStep) Name() string {
	return "artifact"
}

func (s ArtifactStep) Run(ctx context.Context, state *State) error {
	if state == nil || state.Err != nil || state.Journal == nil || s.Recorder == nil {
		return nil
	}
	artifacts, err := s.Recorder.Record(ctx, state.Request)
	if err != nil {
		return err
	}
	for _, artifact := range artifacts {
		if strings.TrimSpace(artifact.ArtifactID) == "" {
			continue
		}
		if err := state.Journal.Append(ctx, events.NewArtifactDeclared(artifact)); err != nil {
			return err
		}
	}
	return nil
}

// DBArtifactRecorder persists artifacts declared in the run workspace manifest.
type DBArtifactRecorder struct {
	db *gorm.DB
}

// NewDBArtifactRecorder creates a manifest-backed artifact recorder.
func NewDBArtifactRecorder(db *gorm.DB) *DBArtifactRecorder {
	return &DBArtifactRecorder{db: db}
}

// Record persists final manifest artifacts for one run.
func (r *DBArtifactRecorder) Record(ctx context.Context, req *agent.RequestContext) ([]events.ArtifactPayload, error) {
	if r == nil || r.db == nil || req == nil {
		return nil, nil
	}
	plan, ok, err := agentworkspace.FromAgentRequest(req)
	if err != nil || !ok {
		return nil, err
	}
	records, err := agentworkspace.CollectFinalArtifacts(ctx, plan)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	sessionID := strings.TrimSpace(req.Conversation.ID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required for artifact declaration")
	}
	session, err := infradb.GetSessionByPublicID(ctx, r.db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find session %s: %w", sessionID, err)
	}
	if session == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if session.ProjectID == nil || *session.ProjectID == 0 {
		return nil, fmt.Errorf("session project_id is required for artifact declaration")
	}
	if session.TaskID == nil || *session.TaskID == 0 {
		return nil, fmt.Errorf("session task_id is required for artifact declaration")
	}

	payloads := make([]events.ArtifactPayload, 0, len(records))
	for _, record := range records {
		artifact, err := r.persistRecord(ctx, session, record)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, artifactPayloadFromRecord(artifact, record))
	}
	return payloads, nil
}

func (r *DBArtifactRecorder) persistRecord(ctx context.Context, session *types.Session, record agentworkspace.ArtifactRecord) (*types.Artifact, error) {
	publicID := fmt.Sprintf("art_%s", snowflake.GenerateIDBase58())
	artifact := &types.Artifact{
		PublicID:     publicID,
		OrgID:        session.OrgID,
		OwnerID:      session.Uin,
		TaskID:       *session.TaskID,
		ProjectID:    *session.ProjectID,
		SessionID:    &session.ID,
		Title:        artifactTitle(record),
		Filename:     artifactFilename(record),
		Description:  strings.TrimSpace(record.Description),
		ArtifactType: artifactType(record.ArtifactType),
		FileURL:      "/v1/artifacts/" + publicID + "/download",
		MimeType:     strings.TrimSpace(record.MimeType),
		FileSize:     record.FileSize,
		RelativePath: strings.TrimSpace(record.RelativePath),
		StorageKey:   strings.TrimSpace(record.StorageKey),
		Sha256:       strings.TrimSpace(record.Sha256),
		Source:       artifactSource(record.Source),
		Status:       artifactStatus(record.Status),
	}
	if err := infradb.CreateArtifact(ctx, r.db, artifact); err != nil {
		return nil, err
	}
	return artifact, nil
}

func artifactPayloadFromRecord(artifact *types.Artifact, record agentworkspace.ArtifactRecord) events.ArtifactPayload {
	return events.ArtifactPayload{
		ArtifactID:   artifact.PublicID,
		Title:        artifact.Title,
		Filename:     artifact.Filename,
		MimeType:     artifact.MimeType,
		ArtifactType: artifact.ArtifactType,
	}
}

func artifactTitle(record agentworkspace.ArtifactRecord) string {
	title := strings.TrimSpace(record.Title)
	if title != "" {
		return title
	}
	return strings.TrimSpace(record.RelativePath)
}

func artifactFilename(record agentworkspace.ArtifactRecord) string {
	filename := strings.TrimSpace(record.Filename)
	if filename != "" {
		return filename
	}
	return strings.TrimSpace(record.RelativePath)
}

func artifactType(value string) string {
	if strings.TrimSpace(value) == "" {
		return string(types.ArtifactTypeFile)
	}
	return strings.TrimSpace(value)
}

func artifactSource(value string) string {
	if strings.TrimSpace(value) == "" {
		return string(types.ArtifactSourceAgentDeclared)
	}
	return strings.TrimSpace(value)
}

func artifactStatus(value string) string {
	if strings.TrimSpace(value) == "" {
		return string(types.ArtifactStatusCompleted)
	}
	return strings.TrimSpace(value)
}

var _ ArtifactRecorder = (*DBArtifactRecorder)(nil)
