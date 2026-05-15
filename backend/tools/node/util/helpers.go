package util

import (
	"fmt"
	"strings"
)

// StringValue 从 map 中获取字符串值并去除首尾空格
func StringValue(input map[string]interface{}, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}

// IntValue 将 interface{} 转换为 int，支持多种数值类型
func IntValue(value interface{}) (int, error) {
	switch typed := value.(type) {
	case nil:
		return 0, nil
	case int:
		return typed, nil
	case int32:
		return int(typed), nil
	case int64:
		return int(typed), nil
	case float64:
		return int(typed), nil
	default:
		return 0, fmt.Errorf("invalid integer value")
	}
}

// BoolValue 将 interface{} 转换为 bool
func BoolValue(value interface{}) (bool, error) {
	switch typed := value.(type) {
	case nil:
		return false, nil
	case bool:
		return typed, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}

// ClampInt 将值限制在 min 和 max 之间
func ClampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

// TruncateOutput 截断输出，超过最大行数时只显示最后部分
func TruncateOutput(output string, maxLines int) (string, bool, int) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", false, 0
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output, false, len(lines)
	}

	return fmt.Sprintf("[输出共 %d 行，显示最后 %d 行]\n%s", len(lines), maxLines, strings.Join(lines[len(lines)-maxLines:], "\n")), true, len(lines)
}

// CombineOutput 合并 stdout 和 stderr，优先返回 stdout
func CombineOutput(stdout string, stderr string) string {
	stdout = strings.TrimSpace(stdout)
	stderr = strings.TrimSpace(stderr)
	switch {
	case stdout != "" && stderr != "":
		return stdout + "\n" + stderr
	case stdout != "":
		return stdout
	default:
		return stderr
	}
}

// CountContentLines 计算内容的行数
func CountContentLines(content string) int {
	if content == "" {
		return 0
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	trimmed := strings.TrimSuffix(normalized, "\n")
	if trimmed == "" {
		return 1
	}

	return len(strings.Split(trimmed, "\n"))
}
