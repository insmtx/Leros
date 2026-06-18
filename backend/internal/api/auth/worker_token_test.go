package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseWorkerToken(t *testing.T) {
	token, expiredAt, err := GenerateWorkerToken(3, 7, "secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateWorkerToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if expiredAt == 0 {
		t.Fatal("expiredAt should not be zero")
	}

	claims, err := ParseWorkerToken(token, "secret")
	if err != nil {
		t.Fatalf("ParseWorkerToken() error = %v", err)
	}
	if claims.OrgID != 3 || claims.WorkerID != 7 {
		t.Fatalf("claims identity = org %d worker %d, want org 3 worker 7", claims.OrgID, claims.WorkerID)
	}
	if claims.Kind != WorkerTokenKind || claims.Scope != WorkerTokenScopeAPI {
		t.Fatalf("claims kind/scope = %q/%q", claims.Kind, claims.Scope)
	}
}

func TestParseWorkerTokenWrongSecret(t *testing.T) {
	token, _, err := GenerateWorkerToken(3, 7, "secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateWorkerToken() error = %v", err)
	}

	if _, err := ParseWorkerToken(token, "wrong"); err == nil {
		t.Fatal("ParseWorkerToken() should fail with wrong secret")
	}
}
