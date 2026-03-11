package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v78/github"
	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/ygpkg/yg-go/logs"
)

func (c *GitHubConnector) HandleWebhook(
	ctx *gin.Context,
) {
	var (
		w = ctx.Writer
		r = ctx.Request
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Errorf("Failed to read GitHub webhook payload: %v", err)
		http.Error(w, "bad request", 400)
		return
	}

	if !c.verifySignature(r, payload) {
		logs.Warnf("Invalid GitHub webhook signature for request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "invalid signature", 401)
		return
	}

	eventType := github.WebHookType(r)

	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		logs.Errorf("Failed to parse GitHub webhook event (type: %s): %v", eventType, err)
		http.Error(w, "parse error", 400)
		return
	}

	interactionEvent := c.convertEvent(eventType, event)

	c.publisher.Publish(ctx, interaction.TopicGithubIssueComment, interactionEvent)

	w.WriteHeader(200)
}

func (c *GitHubConnector) verifySignature(
	r *http.Request,
	payload []byte,
) bool {

	signature := r.Header.Get("X-Hub-Signature-256")

	mac := hmac.New(sha256.New, []byte(c.config.WebhookSecret))
	mac.Write(payload)

	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
