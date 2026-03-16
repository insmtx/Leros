package skills

import (
	"context"
	"testing"
)

// mockSkill 用于测试目的的技能模拟实现
type mockSkill struct {
	BaseSkill
	executeFunc func(context.Context, map[string]interface{}) (map[string]interface{}, error)
}

func (m *mockSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return map[string]interface{}{"result": "mock"}, nil
}

func TestSkillInterface(t *testing.T) {
	// 测试基本接口可用性
	info := &SkillInfo{
		ID:           "test.skill",
		Name:         "Test Skill",
		Description:  "A skill for testing",
		Version:      "1.0.0",
		Category:     "utility",
		SkillType:    LocalSkill,
		Permissions:  []Permission{{Resource: "test", Action: "execute"}},
		InputSchema:  InputSchema{},
		OutputSchema: OutputSchema{},
	}

	skill := &mockSkill{
		BaseSkill: BaseSkill{
			InfoData: info,
		},
	}

	// 测试ID获取
	if skill.GetID() != "test.skill" {
		t.Errorf("Expected ID 'test.skill', got '%s'", skill.GetID())
	}

	// 测试名称获取
	if skill.GetName() != "Test Skill" {
		t.Errorf("Expected name 'Test Skill', got '%s'", skill.GetName())
	}

	// 测试描述获取
	if skill.GetDescription() != "A skill for testing" {
		t.Errorf("Expected description 'A skill for testing', got '%s'", skill.GetDescription())
	}

	// 测试Info获取
	if skill.Info().ID != "test.skill" {
		t.Errorf("Expected Info ID 'test.skill', got '%s'", skill.Info().ID)
	}

	// 测试执行功能
	ctx := context.Background()
	input := map[string]interface{}{"param": "value"}
	result, err := skill.Execute(ctx, input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}
