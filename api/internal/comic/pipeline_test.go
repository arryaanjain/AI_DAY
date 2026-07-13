package comic

import (
	"testing"
)

func TestCleanJSONString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "```json\n{\n  \"key\": \"value\"\n}\n```",
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			input:    "```{\n  \"key\": \"value\"\n}```",
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			input:    "  {\n  \"key\": \"value\"\n}  ",
			expected: "{\n  \"key\": \"value\"\n}",
		},
	}

	for _, test := range tests {
		result := cleanJSONString(test.input)
		if result != test.expected {
			t.Errorf("expected %q, got %q", test.expected, result)
		}
	}
}

func TestDetectImageType(t *testing.T) {
	pngBytes := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	jpgBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	unknownBytes := []byte{0x00, 0x01, 0x02, 0x03}

	if detectImageType(pngBytes) != "PNG" {
		t.Errorf("expected PNG")
	}
	if detectImageType(jpgBytes) != "JPG" {
		t.Errorf("expected JPG")
	}
	if detectImageType(unknownBytes) != "PNG" { // fallback
		t.Errorf("expected PNG fallback")
	}
}
