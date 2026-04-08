package react

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/skills"
)

// CodeReviewSkill 代码审查技能
type CodeReviewSkill struct {
	skills.BaseSkill
}

// NewCodeReviewSkill 创建新的代码审查技能
func NewCodeReviewSkill() *CodeReviewSkill {
	return &CodeReviewSkill{
		BaseSkill: skills.BaseSkill{
			InfoData: &skills.SkillInfo{
				ID:          "code.review",
				Name:        "Code Review Skill",
				Description: "代码审查技能，提供最佳实践建议和潜在问题识别",
				Version:     "1.0.0",
				Category:    "programming",
				SkillType:   skills.LocalSkill,
				InputSchema: skills.InputSchema{
					Type:     "object",
					Required: []string{"code"},
					Properties: map[string]*skills.Property{
						"code": {
							Type:        "string",
							Title:       "代码",
							Description: "需要审查的代码",
						},
						"language": {
							Type:        "string",
							Title:       "编程语言",
							Description: "代码的语言（如Go、JavaScript、Java等）",
						},
						"context": {
							Type:        "string",
							Title:       "上下文",
							Description: "代码上下文或说明",
						},
					},
				},
				OutputSchema: skills.OutputSchema{
					Type:     "object",
					Required: []string{"reviews"},
					Properties: map[string]*skills.Property{
						"reviews": {
							Type:        "array",
							Title:       "审查结果",
							Description: "发现的问题列表",
						},
						"summary": {
							Type:        "string",
							Title:       "摘要",
							Description: "审查结果的摘要",
						},
					},
				},
			},
		},
	}
}

// Execute 执行代码审查技能
func (s *CodeReviewSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code, ok := input["code"].(string)
	if !ok || code == "" {
		return nil, errors.New("code parameter is required and should be a string")
	}

	language, exists := input["language"].(string)
	if !exists {
		language = "Unknown" // 默认语言
	}

	// 分析代码并提供建议
	reviewResults := analyzeCode(code, language)

	result := map[string]interface{}{
		"reviews":       reviewResults,
		"summary":       generateReviewSummary(reviewResults),
		"code_language": language,
	}

	return result, nil
}

// analyzeCode 分析代码并返回问题列表
func analyzeCode(code, language string) []map[string]interface{} {
	var reviews []map[string]interface{}

	// 基础代码健康检查
	lines := strings.Split(code, "\n")

	// 检查过于长的行 (通常超过120字符）
	longLines := []int{}
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) > 120 {
			longLines = append(longLines, i+1)
		}
	}

	if len(longLines) > 0 {
		reviews = append(reviews, map[string]interface{}{
			"type":     "style",
			"category": "line_length",
			"message": fmt.Sprintf("发现了 %d 行过长(>120字符): 行 %v. 建议每行不超过120字符",
				len(longLines), longLines),
			"severity":   "medium",
			"suggestion": "考虑缩短这些行或将它们拆分为多行",
		})
	}

	// 检查未使用的变量（基于简单模式识别）
	unusedVars := findPotentialUnusedVariables(code)
	for _, v := range unusedVars {
		reviews = append(reviews, map[string]interface{}{
			"type":       "potential_issue",
			"category":   "unused_variable",
			"message":    fmt.Sprintf("潜在未使用的变量: %s", v),
			"severity":   "medium",
			"suggestion": "检查变量是否真的需要，或者是否忘记使用它",
		})
	}

	// 语言特有的建议
	switch strings.ToLower(language) {
	case "go":
		goSpecificChecks(code, &reviews)
	case "javascript", "js":
		jsSpecificChecks(code, &reviews)
	case "python", "py":
		pythonSpecificChecks(code, &reviews)
	}

	return reviews
}

// findPotentialUnusedVariables 查找潜在的未使用变量（简化版实现）
func findPotentialUnusedVariables(code string) []string {
	var unusedVars []string

	// 这是一个简化的实现，只是查找可能的变量声明而不检查实际使用
	// 在实际应用中，这将需要一个完整的语法分析器
	// 仅为演示目的实现一个非常基础的版本

	lines := strings.Split(code, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 基础正则表达式样式的匹配（实际需要一个完整的解析器）
		if strings.Contains(line, ":=") {
			// 形如 varName := value 的赋值语句
			parts := strings.Split(line, ":=")
			if len(parts) > 0 {
				left := strings.TrimSpace(parts[0])
				// 从左侧获取变量名
				words := strings.Fields(left)
				if len(words) > 0 {
					varName := words[0]
					if !isUsed(varName, code) && isValidVariableName(varName) {
						unusedVars = append(unusedVars, varName)
					}
				}
			}
		}
	}

	return unusedVars
}

// isUsed 检查变量是否在代码中被使用（简化实现）
func isUsed(varName, code string) bool {
	// 简单判断：如果在 = / := / ( 之前出现了多次，则可能被使用了
	count := strings.Count(code, varName)
	return count > 1
}

// isValidVariableName 检查是否是有效的变量名
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// 检查是否是有效标识符（字母开头，后续可以是字母数字下划线）
	for i, char := range name {
		if i == 0 && !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
			return false
		}
		if i > 0 && !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// goSpecificChecks Go语言特定检查
func goSpecificChecks(code string, reviews *[]map[string]interface{}) {
	// 检查是否所有的error都被检查了（简化版本）
	if strings.Contains(code, "= ") && strings.Contains(code, "error") {
		// 这将需要更复杂的AST分析来真正确认
		// 此处为简化示例
		*reviews = append(*reviews, map[string]interface{}{
			"type":       "best_practice",
			"category":   "error_handling",
			"message":    "检测到可能未检查的错误（请使用更完整的lint工具确认）",
			"severity":   "high",
			"suggestion": "确保所有error值都被检查或使用_显式忽略，例如: if err != nil { return err } 或 _ = ignoredError",
		})
	}
}

// jsSpecificChecks JavaScript特定检查
func jsSpecificChecks(code string, reviews *[]map[string]interface{}) {
	if strings.Contains(code, "var ") {
		*reviews = append(*reviews, map[string]interface{}{
			"type":       "best_practice",
			"category":   "variable_declaration",
			"message":    "建议使用let或const替代var声明变量",
			"severity":   "medium",
			"suggestion": "var在ES6+中不太推荐，建议在合适的场景使用let或const",
		})
	}
}

// pythonSpecificChecks Python特定检查
func pythonSpecificChecks(code string, reviews *[]map[string]interface{}) {
	if strings.Contains(code, "print(") {
		*reviews = append(*reviews, map[string]interface{}{
			"type":       "production_readiness",
			"category":   "debug_print",
			"message":    "检测到 print 语句。生产代码中应该移除调试打印",
			"severity":   "medium",
			"suggestion": "使用适当的日志框架替换print语句",
		})
	}
}

// generateReviewSummary 生成审查结果摘要
func generateReviewSummary(results []map[string]interface{}) string {
	if len(results) == 0 {
		return "代码看起来很棒！没有发现明显的问题。"
	}

	severityCount := make(map[string]int)
	categoryCount := make(map[string]int)

	for _, result := range results {
		if severity, ok := result["severity"].(string); ok {
			severityCount[severity]++
		}
		if category, ok := result["category"].(string); ok {
			categoryCount[category]++
		}
	}

	summary := fmt.Sprintf("完成了代码审查，发现 %d 个问题:", len(results))

	if highCount, exists := severityCount["high"]; exists && highCount > 0 {
		summary += fmt.Sprintf(" %d 个高严重性,", highCount)
	}
	if medCount, exists := severityCount["medium"]; exists && medCount > 0 {
		summary += fmt.Sprintf(" %d 个中等严重性,", medCount)
	}
	if lowCount, exists := severityCount["low"]; exists && lowCount > 0 {
		summary += fmt.Sprintf(" %d 个低严重性问题", lowCount)
	}

	// Remove trailing comma and add period
	if strings.HasSuffix(summary, ",") {
		summary = summary[:len(summary)-1] + "."
	}

	return summary
}

// PRAnalysisSkill GitHub PR分析技能
type PRAnalysisSkill struct {
	skills.BaseSkill
}

// NewPRAnalysisSkill 创建新的PR分析技能
func NewPRAnalysisSkill() *PRAnalysisSkill {
	return &PRAnalysisSkill{
		BaseSkill: skills.BaseSkill{
			InfoData: &skills.SkillInfo{
				ID:          "github.pr_analysis",
				Name:        "GitHub PR Analysis Skill",
				Description: "分析GitHub Pull Request，提供摘要和反馈",
				Version:     "1.0.0",
				Category:    "programming",
				SkillType:   skills.LocalSkill,
				InputSchema: skills.InputSchema{
					Type:     "object",
					Required: []string{"repo", "pr_number", "diff"},
					Properties: map[string]*skills.Property{
						"repo": {
							Type:        "string",
							Title:       "代码仓库",
							Description: "仓库名称(用户名/仓库名)",
						},
						"pr_number": {
							Type:        "integer",
							Title:       "PR编号",
							Description: "Pull Request的编号",
						},
						"diff": {
							Type:        "string",
							Title:       "差异内容",
							Description: "PR的代码差异",
						},
						"title": {
							Type:        "string",
							Title:       "PR标题",
							Description: "Pull Request的标题",
						},
						"description": {
							Type:        "string",
							Title:       "描述",
							Description: "Pull Request的描述",
						},
					},
				},
				OutputSchema: skills.OutputSchema{
					Type:     "object",
					Required: []string{"summary", "suggestions"},
					Properties: map[string]*skills.Property{
						"summary": {
							Type:        "string",
							Title:       "摘要",
							Description: "PR的分析摘要",
						},
						"suggestions": {
							Type:        "array",
							Title:       "建议",
							Description: "改进建议列表",
						},
						"risk_factors": {
							Type:        "array",
							Title:       "风险因素",
							Description: "检测到的风险因素",
						},
					},
				},
			},
		},
	}
}

// Execute 执行PR分析技能
func (s *PRAnalysisSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	repo, ok := input["repo"].(string)
	if !ok || repo == "" {
		return nil, errors.New("repo参数是必需的")
	}

	prNumber, ok := input["pr_number"].(float64) // JSON解码为float64
	if !ok {
		return nil, errors.New("pr_number参数是必需的且必须是数字")
	}

	diff, ok := input["diff"].(string)
	if !ok || diff == "" {
		return nil, errors.New("diff参数是必需的")
	}

	// 可选参数
	title, _ := input["title"].(string)
	description, _ := input["description"].(string)

	// 分析PR并生成结果
	analysis := analyzePR(repo, int(prNumber), title, description, diff)

	result := map[string]interface{}{
		"summary":      analysis.Summary,
		"suggestions":  analysis.Suggestions,
		"risk_factors": analysis.RiskFactors,
		"pr_details": map[string]interface{}{
			"repo":                repo,
			"pr_number":           int(prNumber),
			"title":               title,
			"changed_files_count": analysis.ChangedFilesCount,
		},
	}

	return result, nil
}

// PRAnalysisResult PR分析结果结构
type PRAnalysisResult struct {
	Summary           string                   `json:"summary"`
	Suggestions       []map[string]interface{} `json:"suggestions"`
	RiskFactors       []map[string]interface{} `json:"risk_factors"`
	ChangedFilesCount int                      `json:"changed_files_count"`
}

// analyzePR 分析PR并返回结构化结果
func analyzePR(repo string, prNumber int, title, description, diff string) *PRAnalysisResult {
	result := &PRAnalysisResult{
		ChangedFilesCount: countChangedFiles(diff),
	}

	// 生成摘要
	summaries := []string{}
	if title != "" {
		summaries = append(summaries, fmt.Sprintf("PR #%d: '%s' 提交到了 %s", prNumber, title, repo))
	} else {
		summaries = append(summaries, fmt.Sprintf("PR #%d 提交到了 %s", prNumber, repo))
	}

	summaries = append(summaries, fmt.Sprintf("共变更了 %d 个文件", result.ChangedFilesCount))

	// 执行具体分析
	reviewResults := analyzeDiff(diff)

	if len(reviewResults) > 0 {
		summaries = append(summaries, fmt.Sprintf("发现了 %d 个值得注目的问题", len(reviewResults)))
	}

	result.Summary = strings.Join(summaries, ". ")

	// 建议
	result.Suggestions = generateSuggestions(diff)

	// 风险因素
	result.RiskFactors = identifyRiskFactors(diff)

	return result
}

// countChangedFiles 计算变更的文件数
func countChangedFiles(diff string) int {
	// 简化计算：计算文件分割符 'diff --git'
	count := strings.Count(diff, "diff --git")
	if count == 0 {
		// 如果没有git格式的diff标记，使用另一种策略
		count = strings.Count(diff, "+++")
		if count == 0 {
			return 1 // 至少一个文件
		}
	}
	return count
}

// analyzeDiff 分析diff内容
func analyzeDiff(diff string) []map[string]interface{} {
	// 此处的实现只是一个示例
	// 在实际应用中，这将需要一个完整的diff解析器
	return []map[string]interface{}{}
}

// generateSuggestions 基于diff生成建议
func generateSuggestions(diff string) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	// 检查是否有大量删除
	if strings.Count(diff, "-") > 50 && float64(strings.Count(diff, "+"))/float64(strings.Count(diff, "-")) < 0.5 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":     "refactoring",
			"category": "significant_deletion",
			"text":     "检测到大量的删除操作，确认这不是意外造成的",
		})
	}

	// 检查是否有新增测试文件
	if strings.Contains(strings.ToLower(diff), "test") || strings.Contains(strings.ToLower(diff), "_test") {
		suggestions = append(suggestions, map[string]interface{}{
			"type":     "testing",
			"category": "tests_added",
			"text":     "很好的做法！添加了新的测试以验证功能",
		})
	}

	// 确保建议的多样化
	if len(suggestions) == 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":     "general",
			"category": "positive_feedback",
			"text":     "PR结构良好，变更清晰明了。感谢您的贡献！",
		})
	}

	return suggestions
}

// identifyRiskFactors 识别风险因素
func identifyRiskFactors(diff string) []map[string]interface{} {
	riskFactors := []map[string]interface{}{}

	// 搜索关键风险指标
	if strings.Contains(diff, "TODO") || strings.Contains(diff, "FIXME") || strings.Contains(diff, "HACK") {
		riskFactors = append(riskFactors, map[string]interface{}{
			"type":   "technical_debt",
			"level":  "medium",
			"detail": "在代码中发现了TODO/FIXME/HACK注释，可能代表遗留的技术债",
		})
	}

	// 检查配置文件的变更（通常被认为是高风险）
	configExtensions := []string{".yaml", ".yml", ".json", ".toml", ".ini", ".conf", ".env"}
	for _, ext := range configExtensions {
		if strings.Contains(diff, ext) {
			riskFactors = append(riskFactors, map[string]interface{}{
				"type":   "configuration_change",
				"level":  "high",
				"detail": "检测到配置文件的变更，这类更改往往影响较大请谨慎测试",
			})
			break
		}
	}

	// 检索敏感关键词
	sensitivePatterns := []string{"password", "secret", "token", "key", "credential"}
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(diff), pattern) {
			riskFactors = append(riskFactors, map[string]interface{}{
				"type":   "security_risk",
				"level":  "high",
				"detail": fmt.Sprintf("在变更中检测到敏感词 '%s'。请确保不会暴露实际密钥", pattern),
			})
		}
	}

	return riskFactors
}
