package security

import (
	"regexp"
	"strings"
)

// DangerousPattern 危险命令模式
type DangerousPattern struct {
	Regex    *regexp.Regexp
	Name     string
	Severity string
}

// ApprovalResult 命令审批结果
type ApprovalResult struct {
	Approved    bool
	Status      string
	PatternKey  string
	Description string
	Message     string
}

var dangerousPatterns = []DangerousPattern{
	{Regex: regexp.MustCompile(`\brm\s+(-[^\s]*\s+)*/`), Name: "delete in root path", Severity: "high"},
	{Regex: regexp.MustCompile(`\brm\s+-[^\s]*r\b`), Name: "recursive delete", Severity: "high"},
	{Regex: regexp.MustCompile(`\brm\s+--recursive\b`), Name: "recursive delete (long flag)", Severity: "high"},
	{Regex: regexp.MustCompile(`\bchmod\s+(-[^\s]*\s+)*(777|666|o\+[rwx]*w|a\+[rwx]*w)\b`), Name: "world-writable permissions", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bchown\s+(-[^\s]*)?R\s+root`), Name: "recursive chown to root", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bmkfs\b`), Name: "format filesystem", Severity: "high"},
	{Regex: regexp.MustCompile(`\bdd\s+.*if=`), Name: "disk copy", Severity: "high"},
	{Regex: regexp.MustCompile(`>\s*/dev/sd`), Name: "write to block device", Severity: "high"},
	{Regex: regexp.MustCompile(`(?i)\bDROP\s+(TABLE|DATABASE)\b`), Name: "SQL DROP", Severity: "high"},
	{Regex: regexp.MustCompile(`(?i)\bDELETE\s+FROM\b\s*;?\s*$`), Name: "SQL DELETE without WHERE", Severity: "high"},
	{Regex: regexp.MustCompile(`(?i)\bTRUNCATE\s+(TABLE)?\s*\w`), Name: "SQL TRUNCATE", Severity: "high"},
	{Regex: regexp.MustCompile(`>\s*/etc/`), Name: "overwrite system config", Severity: "high"},
	{Regex: regexp.MustCompile(`(?i)\bsystemctl\s+(stop|disable|mask)\b`), Name: "stop/disable system service", Severity: "high"},
	{Regex: regexp.MustCompile(`\bkill\s+-9\s+-1\b`), Name: "kill all processes", Severity: "high"},
	{Regex: regexp.MustCompile(`\bpkill\s+-9\b`), Name: "force kill processes", Severity: "high"},
	{Regex: regexp.MustCompile(`:\(\)\s*\{\s*:\s*\|\s*:\s*&\s*\}\s*;`), Name: "fork bomb", Severity: "high"},
	{Regex: regexp.MustCompile(`\b(bash|sh|zsh|ksh)\s+-[^\s]*c(\s+|$)`), Name: "shell command via -c flag", Severity: "medium"},
	{Regex: regexp.MustCompile(`\b(python[23]?|perl|ruby|node)\s+-[ec]\s+`), Name: "script execution via -e/-c flag", Severity: "medium"},
	{Regex: regexp.MustCompile(`\b(curl|wget)\b.*\|\s*(ba)?sh\b`), Name: "pipe remote content to shell", Severity: "high"},
	{Regex: regexp.MustCompile(`\bxargs\s+.*\brm\b`), Name: "xargs with rm", Severity: "high"},
	{Regex: regexp.MustCompile(`\bfind\b.*-exec\s+(/\S*/)?rm\b`), Name: "find -exec rm", Severity: "high"},
	{Regex: regexp.MustCompile(`\bfind\b.*-delete\b`), Name: "find -delete", Severity: "high"},
	{Regex: regexp.MustCompile(`\b(pkill|killall)\b.*\b(hermes|gateway|cli\.py)\b`), Name: "kill hermes/gateway process", Severity: "high"},
	{Regex: regexp.MustCompile(`\bkill\b.*\$\(\s*pgrep\b`), Name: "kill process via pgrep expansion", Severity: "high"},
	{Regex: regexp.MustCompile(`\b(cp|mv|install)\b.*\s/etc/`), Name: "copy/move file into /etc/", Severity: "high"},
	{Regex: regexp.MustCompile(`\bsed\s+-[^\s]*i.*\s/etc/`), Name: "in-place edit of system config", Severity: "high"},
	{Regex: regexp.MustCompile(`\bgit\s+reset\s+--hard\b`), Name: "git reset --hard", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bgit\s+push\b.*(--force|-f)\b`), Name: "git force push", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bgit\s+clean\s+-[^\s]*f`), Name: "git clean with force", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bgit\s+branch\s+-D\b`), Name: "git branch force delete", Severity: "medium"},
	{Regex: regexp.MustCompile(`\bchmod\s+\+x\b.*[;&|]+\s*\./`), Name: "chmod +x followed by execution", Severity: "medium"},
}

// CheckDangerousCommand 检查命令是否危险
func CheckDangerousCommand(command string) ApprovalResult {
	command = strings.TrimSpace(command)
	if command == "" {
		return ApprovalResult{Approved: true, Status: "approved"}
	}

	for _, pattern := range dangerousPatterns {
		if pattern.Regex.MatchString(command) {
			return ApprovalResult{
				Approved:    false,
				Status:      "blocked",
				PatternKey:  pattern.Name,
				Description: "命令存在风险: " + pattern.Name,
				Message:     "BLOCKED: 命令被阻止 (危险模式: " + pattern.Name + ")",
			}
		}
	}

	return ApprovalResult{Approved: true, Status: "approved"}
}
