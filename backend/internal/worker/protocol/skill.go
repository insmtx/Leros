package protocol

// SkillInstallMessage is the message protocol from Server to Worker for skill installation.
// Deprecated: use SkillManagementMessage with Action="install" instead.
type SkillInstallMessage = Envelope[SkillInstallBody]

// SkillInstallBody carries the source hint and skill identifier for installation.
// Deprecated: use SkillManagementBody instead.
type SkillInstallBody struct {
	Source  string `json:"source"`   // "Leros" | "github" | "skills-sh" | "url"
	SkillID string `json:"skill_id"` // the CLI install <identifier> argument
}

// SkillManagementMessage is the unified skill management message from Server to Worker.
type SkillManagementMessage = Envelope[SkillManagementBody]

// SkillManagementBody carries the action and parameters for skill management.
type SkillManagementBody struct {
	Action  string `json:"action"`              // "install" | "list" | "uninstall"
	Source  string `json:"source,omitempty"`    // for install: "Leros" | "github" | "skills-sh" | "url"
	SkillID string `json:"skill_id,omitempty"`  // for install: the CLI install <identifier> argument
	Name    string `json:"name,omitempty"`      // for uninstall: the skill name to remove
	// ReplyTo is the NATS inbox for sending the response back to the server.
	// JetStream does not preserve the NATS Reply header, so the inbox is
	// injected into the body by the server-side Request method.
	ReplyTo string `json:"reply_to,omitempty"`
}

// SkillManagementResponse is the response returned by the worker for skill management requests.
type SkillManagementResponse struct {
	Success bool   `json:"success"`
	Action  string `json:"action"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"` // for list action: []SkillListItem
}

// SkillListItem represents an installed skill in the list response.
type SkillListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Source      string `json:"source"`
	Trust       string `json:"trust"`
}
