package protocol

// SkillInstallMessage is the message protocol from Server to Worker for skill installation.
type SkillInstallMessage = Envelope[SkillInstallBody]

// SkillInstallBody carries the source hint and skill identifier for installation.
type SkillInstallBody struct {
	Source  string `json:"source"`   // "Leros" | "github" | "skills-sh" | "url"
	SkillID string `json:"skill_id"` // the CLI install <identifier> argument
}
