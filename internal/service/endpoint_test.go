package service

import "testing"

func TestBuildEndpointURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "adds default path to bare host",
			in:   "example.com",
			want: "http://example.com/api.php/provide/vod",
		},
		{
			name: "adds default path to root url",
			in:   "https://example.com/",
			want: "https://example.com/api.php/provide/vod",
		},
		{
			name: "preserves explicit api path",
			in:   "https://example.com/custom/api",
			want: "https://example.com/custom/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEndpointURL(tt.in)
			if got != tt.want {
				t.Fatalf("BuildEndpointURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
