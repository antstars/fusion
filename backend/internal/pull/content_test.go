package pull

import "testing"

func TestHasReadableItemContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "empty string", content: "", want: false},
		{name: "whitespace", content: " \n\t ", want: false},
		{name: "empty html", content: "<p></p><br><hr>", want: false},
		{name: "html text", content: "<p>Hello <strong>world</strong></p>", want: true},
		{name: "plain text", content: "Hello world", want: true},
		{name: "html entity text", content: "<p>&nbsp;</p><p>&amp;</p>", want: true},
		{name: "image only", content: `<p><img src="https://example.com/a.jpg"></p>`, want: true},
		{name: "video only", content: `<video src="https://example.com/a.mp4"></video>`, want: true},
		{name: "script text ignored", content: "<script>alert('x')</script>", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasReadableItemContent(tt.content); got != tt.want {
				t.Fatalf("hasReadableItemContent(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestBuildReadableItemInputsSkipsEmptyContent(t *testing.T) {
	inputs, skipped := buildReadableItemInputs([]*ParsedItem{
		{GUID: "empty", Title: "Empty", Link: "https://example.com/empty", Content: "<p></p>"},
		{GUID: "text", Title: "Text", Link: "https://example.com/text", Content: "<p>Body</p>", PubDate: 123},
		{GUID: "image", Title: "Image", Link: "https://example.com/image", Content: `<img src="https://example.com/a.jpg">`, PubDate: 456},
	})

	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1", skipped)
	}
	if len(inputs) != 2 {
		t.Fatalf("len(inputs) = %d, want 2", len(inputs))
	}
	if inputs[0].GUID != "text" || inputs[1].GUID != "image" {
		t.Fatalf("unexpected inputs: %#v", inputs)
	}
}
