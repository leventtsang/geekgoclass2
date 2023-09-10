package middleware

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func generateJWTToken(claims web.UserClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(web.JWTKey)
	if err != nil {
		panic(err)
	}
	return ss
}

func TestJWTLoginMiddlewareBuilder_Build(t *testing.T) {
	validClaims := web.UserClaims{
		UserAgent: "testAgent",
		// Add more fields if needed
	}
	validClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))

	validToken := generateJWTToken(validClaims)

	tests := []struct {
		name           string
		requestPath    string
		authHeader     string
		userAgent      string
		expectStatus   int
		expectNewToken bool
	}{
		{
			"Public path - no auth required",
			"/users/signup",
			"",
			"",
			http.StatusOK,
			false,
		},
		{
			"Missing Authorization header",
			"/private-path",
			"",
			"",
			http.StatusUnauthorized,
			false,
		},
		{
			"Invalid Authorization format",
			"/private-path",
			"InvalidFormat",
			"",
			http.StatusUnauthorized,
			false,
		},
		{
			"Valid token",
			"/private-path",
			fmt.Sprintf("Bearer %s", validToken),
			validClaims.UserAgent,
			http.StatusOK,
			false,
		},
		// Add more cases as necessary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			builder := NewJWTLoginMiddlewareBuilder()
			r.Use(builder.Build())
			r.GET(tt.requestPath, func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			req, _ := http.NewRequest(http.MethodGet, tt.requestPath, nil)
			req.Header.Set("Authorization", tt.authHeader)
			req.Header.Set("User-Agent", tt.userAgent)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, but got %d", tt.expectStatus, w.Code)
			}

			_, exists := w.HeaderMap["x-jwt-token"]
			if exists != tt.expectNewToken {
				t.Errorf("Expected new token %v, but got %v", tt.expectNewToken, exists)
			}
		})
	}
}
