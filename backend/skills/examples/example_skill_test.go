package examples

import (
	"context"
	"testing"
)

func TestExampleSkill(t *testing.T) {
	skill := NewExampleSkill()

	// 测试基本信息获取
	if skill.GetID() != "example.hello_world" {
		t.Errorf("Expected ID 'example.hello_world', got '%s'", skill.GetID())
	}

	if skill.GetName() != "Hello World Skill" {
		t.Errorf("Expected name 'Hello World Skill', got '%s'", skill.GetName())
	}

	// 测试执行功能
	ctx := context.Background()
	input := map[string]interface{}{
		"name":     "World",
		"greeting": "Hi",
	}

	result, err := skill.Execute(ctx, input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedMessage := "Hi, World!"
	if result["message"] != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result["message"])
	}

	// 测试缺少必需参数的情况
	badInput := map[string]interface{}{
		"greeting": "Hi",
	}

	_, err = skill.Execute(ctx, badInput)
	if err == nil {
		t.Error("Expected error when 'name' parameter is missing")
	}
}
