package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/skill/fetch"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/insmtx/Leros/backend/types"
)

type skillMarketplaceService struct {
	db        *gorm.DB
	publisher eventbus.Publisher
}

// NewSkillMarketplaceService 创建 Skill 市场服务。
func NewSkillMarketplaceService(db *gorm.DB, publisher eventbus.Publisher) contract.SkillMarketplaceService {
	return &skillMarketplaceService{db: db, publisher: publisher}
}

func (s *skillMarketplaceService) SearchSkillMarketplace(ctx context.Context, req *contract.SearchSkillMarketplaceRequest) (*contract.SearchSkillMarketplaceResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 80
	}
	if req.Limit > 200 {
		req.Limit = 200
	}

	// 决定查询哪些源
	queryBuiltin, queryExternal := s.resolveSources(req.SourceTypes)

	keyword := strings.TrimSpace(req.Keyword)
	var externalQuery string

	if keyword == "" {
		if req.Category != "" {
			externalQuery = req.Category
		} else {
			externalQuery = "office"
		}
	} else {
		externalQuery = keyword
	}

	var (
		mu       sync.Mutex
		allItems []contract.SkillMarketplaceItemView
		warnings []contract.SkillSourceWarning
		wg       sync.WaitGroup
	)

	// 内置源：优先排在前面
	if queryBuiltin {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := s.searchBuiltin(ctx, req.Keyword, req.Category, req.Limit)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				warnings = append(warnings, contract.SkillSourceWarning{
					SourceType: "Leros",
					Message:    err.Error(),
				})
			} else {
				allItems = append(allItems, items...)
			}
		}()
	}

	// 外部源（skills.sh）
	if queryExternal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metas, err := fetch.NewSkillsShSource().Search(ctx, externalQuery, req.Limit)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				warnings = append(warnings, contract.SkillSourceWarning{
					SourceType: "Skills.sh",
					Message:    err.Error(),
				})
			} else {
				for _, meta := range metas {
					allItems = append(allItems, metaToView(meta))
				}
			}
		}()
	}

	wg.Wait()

	// 首屏聚合：内置源优先，截断至 limit。
	if len(allItems) > req.Limit {
		allItems = allItems[:req.Limit]
	}

	return &contract.SearchSkillMarketplaceResponse{
		Items:    allItems,
		Warnings: warnings,
	}, nil
}

// resolveSources 根据 source_types 决定查询哪些源。
func (s *skillMarketplaceService) resolveSources(sourceTypes []string) (builtin, external bool) {
	if len(sourceTypes) == 0 {
		return true, true
	}
	for _, t := range sourceTypes {
		switch t {
		case "Leros":
			builtin = true
		case "Skills.sh":
			external = true
		}
	}
	return
}

// searchBuiltin 从数据库查询内置 Skill。
func (s *skillMarketplaceService) searchBuiltin(ctx context.Context, keyword, category string, limit int) ([]contract.SkillMarketplaceItemView, error) {
	items, err := infradb.SearchBuiltinSkills(ctx, s.db, keyword, category, limit)
	if err != nil {
		return nil, err
	}

	result := make([]contract.SkillMarketplaceItemView, 0, len(items))
	for _, item := range items {
		result = append(result, builtinItemToView(item))
	}
	return result, nil
}

func builtinItemToView(item types.BuiltinSkillMarketplaceItem) contract.SkillMarketplaceItemView {
	return contract.SkillMarketplaceItemView{
		SourceType:  "Leros",
		SkillID:     item.SkillID,
		Name:        item.Name,
		Description: item.Description,
		Version:     item.Version,
		Author:      item.Author,
		Category:    item.Category,
		Tags:        []string(item.Tags),
		Icon:        item.Icon,
		Installs:    item.Installs,
	}
}

func metaToView(meta fetch.SkillMeta) contract.SkillMarketplaceItemView {
	return contract.SkillMarketplaceItemView{
		SourceType:  meta.Source,
		SkillID:     meta.SkillID,
		Name:        meta.Name,
		Description: meta.Description,
		Version:     meta.Version,
		Author:      meta.Author,
		Category:    meta.Category,
		Tags:        meta.Tags,
		Icon:        meta.Icon,
		Installs:    meta.Installs,
	}
}

func (s *skillMarketplaceService) DownloadBuiltinSkill(ctx context.Context, skillID string) (*contract.SkillPackageDownload, error) {
	item, err := infradb.GetBuiltinSkillByID(ctx, s.db, skillID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("skill not found")
	}

	serverDir, err := infradb.ResolveSkillsServerDir()
	if err != nil {
		return nil, fmt.Errorf("resolve skills server dir: %w", err)
	}

	skillDir := filepath.Join(serverDir, skillID)
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill not found")
	}

	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(zipSkillDir(ctx, pw, skillDir))
	}()

	return &contract.SkillPackageDownload{
		Reader:   pr,
		FileName: skillID + ".zip",
	}, nil
}

func zipSkillDir(ctx context.Context, w io.Writer, skillDir string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.ToSlash(relPath)

		f, err := zw.Create(zipPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		file.Close()
		return err
	})
}

func (s *skillMarketplaceService) InstallSkill(ctx context.Context, req *contract.InstallSkillRequest) (*contract.InstallSkillResponse, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}

	workerID := uint(1)

	topic, err := dm.WorkerSkillInstallSubject(caller.OrgID, workerID)
	if err != nil {
		return nil, fmt.Errorf("build skill install topic: %w", err)
	}

	msg := protocol.SkillInstallMessage{
		ID:        fmt.Sprintf("%d-%d-%d", caller.OrgID, workerID, time.Now().UnixNano()),
		Type:      protocol.MessageTypeSkillInstall,
		CreatedAt: time.Now(),
		Route: protocol.RouteContext{
			OrgID:    caller.OrgID,
			WorkerID: workerID,
		},
		Body: protocol.SkillInstallBody{
			Source:  strings.TrimSpace(req.Source),
			SkillID: strings.TrimSpace(req.SkillID),
		},
	}

	if err := s.publisher.Publish(ctx, topic, msg); err != nil {
		return nil, fmt.Errorf("publish skill install: %w", err)
	}

	return &contract.InstallSkillResponse{
		Status:  "accepted",
		Message: fmt.Sprintf("Skill install request queued for org %d, worker %d", caller.OrgID, workerID),
	}, nil
}

var _ contract.SkillMarketplaceService = (*skillMarketplaceService)(nil)
