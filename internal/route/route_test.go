package route

import (
	"net/http"
	"net/url"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	type args struct {
		next *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test case: AuthMiddleware",
			args: args{
				next: &http.Request{
					URL: &url.URL{
						Path: "/api/v1/organizations",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AuthMiddleware(tt.args.next)
		})
	}
}
