package builder

// We are testing internal functions, so we use package builder instead of builder_test
// This allows access to chunkContent helper.

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChunker_Golden(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Chunk
	}{
		{
			name:    "Small file (1 chunk)",
			content: "line1\nline2",
			expected: []Chunk{
				{FilePath: "test.txt", StartLine: 1, EndLine: 2, Content: "line1\nline2"},
			},
		},
		{
			name:    "Exact 200 lines",
			content: strings.Repeat("line\n", 199) + "line", // 200 lines
			expected: []Chunk{
				{FilePath: "test.txt", StartLine: 1, EndLine: 200, Content: strings.Repeat("line\n", 199) + "line"},
			},
		},
		{
			name:    "201 lines (2 chunks)",
			content: strings.Repeat("line\n", 200) + "line201",
			expected: []Chunk{
				{FilePath: "test.txt", StartLine: 1, EndLine: 200, Content: strings.Repeat("line\n", 199) + "line"},
				{FilePath: "test.txt", StartLine: 201, EndLine: 201, Content: "line201"},
			},
		},
		{
			name:    "CRLFs normalized",
			content: "line1\r\nline2",
			expected: []Chunk{
				{FilePath: "test.txt", StartLine: 1, EndLine: 2, Content: "line1\nline2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chunkContent("test.txt", tt.content)
			if len(got) != len(tt.expected) {
				t.Errorf("got %d chunks, want %d", len(got), len(tt.expected))
				return
			}
			for i, c := range got {
				if c.StartLine != tt.expected[i].StartLine {
					t.Errorf("chunk[%d] StartLine=%d, want %d", i, c.StartLine, tt.expected[i].StartLine)
				}
				if c.EndLine != tt.expected[i].EndLine {
					t.Errorf("chunk[%d] EndLine=%d, want %d", i, c.EndLine, tt.expected[i].EndLine)
				}
				if c.Content != tt.expected[i].Content {
					// Truncate long content for logging
					gotC := c.Content
					if len(gotC) > 20 {
						gotC = gotC[:20] + "..."
					}
					wantC := tt.expected[i].Content
					if len(wantC) > 20 {
						wantC = wantC[:20] + "..."
					}
					t.Errorf("chunk[%d] Content mismatch. Got %q, want %q", i, gotC, wantC)
				}
				// Check JSON marshalling stability?
				_, err := json.Marshal(c)
				if err != nil {
					t.Errorf("chunk[%d] failed to marshal: %v", i, err)
				}
			}
		})
	}
}

func TestIsText(t *testing.T) {
	if !isText([]byte("hello world")) {
		t.Error("expected text to be text")
	}
	if isText([]byte("hello\x00world")) {
		t.Error("expected null byte to be non-text")
	}
	// Invalid UTF8
	if isText([]byte{0xff, 0xfe, 0xfd}) {
		t.Error("expected invalid utf8 to be non-text")
	}
}
