package cli

import "testing"

func TestParseLink(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantID  string
		wantKey string
		wantErr bool
	}{
		{
			name:    "valid link",
			link:    "http://localhost:8080/s/abc123#keydata",
			wantID:  "abc123",
			wantKey: "keydata",
		},
		{
			name:    "valid link with long key",
			link:    "https://kip.dev/s/xY9kL2mN3pQ4#3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
			wantID:  "xY9kL2mN3pQ4",
			wantKey: "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
		},
		{
			name:    "escaped hash from shell",
			link:    `http://localhost:8080/s/abc123\#keydata`,
			wantID:  "abc123",
			wantKey: "keydata",
		},
		{
			name:    "missing fragment",
			link:    "http://localhost:8080/s/abc123",
			wantErr: true,
		},
		{
			name:    "empty key",
			link:    "http://localhost:8080/s/abc123#",
			wantErr: true,
		},
		{
			name:    "missing /s/ path",
			link:    "http://localhost:8080/abc123#key",
			wantErr: true,
		},
		{
			name:    "empty ID",
			link:    "http://localhost:8080/s/#key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, key, err := parseLink(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id != tt.wantID {
					t.Errorf("parseLink() id = %q, want %q", id, tt.wantID)
				}
				if key != tt.wantKey {
					t.Errorf("parseLink() key = %q, want %q", key, tt.wantKey)
				}
			}
		})
	}
}
