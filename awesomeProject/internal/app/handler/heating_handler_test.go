package handler

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func TestParsePositiveID(t *testing.T) {
	if id, ok := parsePositiveID("15"); !ok || id != 15 {
		t.Fatalf("expected id=15, got id=%d ok=%v", id, ok)
	}
	for _, value := range []string{"", "0", "-1", "abc"} {
		if _, ok := parsePositiveID(value); ok {
			t.Fatalf("%q must be rejected", value)
		}
	}
}

func TestValidateMediaFile(t *testing.T) {
	pngHeader := fileHeader("image.png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0})
	if err := validateMediaFile(pngHeader, "image/"); err != nil {
		t.Fatalf("valid png rejected: %v", err)
	}

	textHeader := fileHeader("fake.png", []byte("plain text, not an image"))
	if err := validateMediaFile(textHeader, "image/"); err == nil {
		t.Fatal("non-image content must be rejected")
	}
}

func fileHeader(name string, content []byte) *multipart.FileHeader {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	h.Set("Content-Type", "application/octet-stream")
	part, _ := writer.CreatePart(h)
	_, _ = part.Write(content)
	_ = writer.Close()

	reader := multipart.NewReader(&body, writer.Boundary())
	form, _ := reader.ReadForm(1 << 20)
	return form.File["file"][0]
}
