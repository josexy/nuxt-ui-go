package nui

import "testing"

func TestFormatFileSize(t *testing.T) {
	for bytes, want := range map[int64]string{0: "0B", 512: "512B", 1023: "1023B", 1024: "1KB", 1536: "2KB", 12 * 1024: "12KB", 3 << 20: "3MB", 5 << 40: "5TB", 5 << 50: "5120TB"} {
		if got := formatFileSize(bytes); got != want {
			t.Errorf("formatFileSize(%d) = %q, want %q", bytes, got, want)
		}
	}
}
