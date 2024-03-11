package lockfile

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	entry := LockVersion{
		Version: "1.0.0",
		Image:   "test-image",
	}

	err := Write(&buf, entry)
	if err != nil {
		t.Errorf("error writing to file: %v", err)
	}

	expected := `{"version":"1.0.0","image":"test-image"}
`

	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}

func TestRead(t *testing.T) {
	reader := bytes.NewBufferString(`{"version":"1.0.0","image":"test-image"}
{"version":"2.0.0","image":"test-image-2"}`)

	data, err := Read(reader)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	want := []LockVersion{
		{Version: "1.0.0", Image: "test-image"},
		{Version: "2.0.0", Image: "test-image-2"},
	}

	if len(data) != len(want) {
		t.Errorf("Expected %v, got %v", want, data)
	}

	for i, d := range data {
		if d != want[i] {
			t.Errorf("at %v: expected %v, got %v", i, want[i], d)
		}
	}
}

func TestVersionExists(t *testing.T) {
	input := bytes.NewBufferString(`{"version":"1.0.0","image":"test-image"}
{"version":"2.0.0","image":"test-image-2"}`)

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "Existing version", version: "1.0.0", want: true},
		{name: "Non-existing version", version: "3.0.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VersionExists(input, tt.version)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("VersionExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getwd() (dir string, err error) {
	return "/mock/dir", nil
}

func TestLockPath(t *testing.T) {
	expected := "/mock/dir/.goship/goship-lock.json"
	lockpath, err := lockPath(getwd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if lockpath != expected {
		t.Errorf("Expected %v, got %s", expected, lockpath)
	}
}
