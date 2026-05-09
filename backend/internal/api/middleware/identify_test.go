package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	localauth "github.com/insmtx/SingerOS/backend/internal/api/auth"
	ygauth "github.com/ygpkg/yg-go/apis/runtime/auth"
)

const testJWTSecret = "test-secret-key"

func generateTestJWT(uin uint, issuer string) (string, error) {
	claims := &ygauth.UserClaims{
		Uin:       uin,
		Issuer:    issuer,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(testJWTSecret))
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	return ctx, w
}

func TestCallerMiddleware_NoAuthHeader(t *testing.T) {
	ctx, _ := setupTestContext()
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.Uin != 0 {
		t.Errorf("expected Uin 0, got %d", caller.Uin)
	}
	if caller.State != localauth.AuthStateNil {
		t.Errorf("expected State AuthStateNil, got %v", caller.State)
	}
}

func TestCallerMiddleware_EmptyAuthHeader(t *testing.T) {
	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "")
	ctx.Request = req

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.State != localauth.AuthStateNil {
		t.Errorf("expected State AuthStateNil, got %v", caller.State)
	}
}

func TestCallerMiddleware_InvalidToken(t *testing.T) {
	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	ctx.Request = req

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.State != localauth.AuthStateFailed {
		t.Errorf("expected State AuthStateFailed, got %v", caller.State)
	}
}

func TestCallerMiddleware_ValidToken(t *testing.T) {
	testUin := uint(12345)
	token, err := generateTestJWT(testUin, "test-issuer")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx.Request = req

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.Uin != testUin {
		t.Errorf("expected Uin %d, got %d", testUin, caller.Uin)
	}
	if caller.State != localauth.AuthStateSucc {
		t.Errorf("expected State AuthStateSucc, got %v", caller.State)
	}
}

func TestCallerMiddleware_RequestIDAndTraceID(t *testing.T) {
	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(headerKeyRequestID, "test-request-id")
	req.Header.Set(headerKeyTraceID, "test-trace-id")
	ctx.Request = req

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	_, trace := localauth.FromGinContext(ctx)
	if trace == nil {
		t.Fatal("trace should not be nil")
	}
	if trace.RequestID != "test-request-id" {
		t.Errorf("expected RequestID test-request-id, got %s", trace.RequestID)
	}
	if trace.TraceID != "test-trace-id" {
		t.Errorf("expected TraceID test-trace-id, got %s", trace.TraceID)
	}
}

func TestCallerMiddleware_AutoGenerateRequestID(t *testing.T) {
	ctx, _ := setupTestContext()
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	_, trace := localauth.FromGinContext(ctx)
	if trace == nil {
		t.Fatal("trace should not be nil")
	}
	if trace.RequestID == "" {
		t.Error("RequestID should be auto-generated when not provided")
	}
	if trace.TraceID != trace.RequestID {
		t.Error("TraceID should equal RequestID when not provided")
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{
			name:       "Bearer token",
			authHeader: "Bearer my-token",
			want:       "my-token",
		},
		{
			name:       "Bearer token with spaces",
			authHeader: "Bearer   my-token  ",
			want:       "my-token",
		},
		{
			name:       "Direct token",
			authHeader: "my-direct-token",
			want:       "my-direct-token",
		},
		{
			name:       "Empty header",
			authHeader: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTokenFromHeader(tt.authHeader)
			if got != tt.want {
				t.Errorf("extractTokenFromHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallerMiddleware_WrongSecret(t *testing.T) {
	token, _ := generateTestJWT(12345, "test-issuer")

	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx.Request = req

	middleware := CallerMiddleware("wrong-secret")
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.State != localauth.AuthStateFailed {
		t.Errorf("expected State AuthStateFailed with wrong secret, got %v", caller.State)
	}
}

func TestCallerMiddleware_ZeroUin(t *testing.T) {
	token, _ := generateTestJWT(0, "test-issuer")

	ctx, _ := setupTestContext()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx.Request = req

	middleware := CallerMiddleware(testJWTSecret)
	middleware(ctx)

	caller, _ := localauth.FromGinContext(ctx)
	if caller == nil {
		t.Fatal("caller should not be nil")
	}
	if caller.State != localauth.AuthStateFailed {
		t.Errorf("expected State AuthStateFailed for zero Uin, got %v", caller.State)
	}
}
