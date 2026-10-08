//go:build !js

package reader

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// convert runs an external pre-processor that turns a binary bill into
// layout text. The engine deliberately does not parse PDF itself: no Go
// library handles real statements reliably, and poppler's pdftotext does.
func convert(name, filename string, data []byte) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "pdftotext-layout", "pdftotext":
		return pdftotext(filename, data)
	default:
		return nil, fmt.Errorf("unknown converter %q (supported: pdftotext-layout)", name)
	}
}

func pdftotext(filename string, data []byte) ([]byte, error) {
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return nil, fmt.Errorf("pdftotext not found in PATH; install poppler-utils, or convert the pdf yourself with `pdftotext -layout %s` and import the .txt", filepath.Base(filename))
	}
	// pdftotext needs a file, and the caller may hold only bytes.
	input := filename
	if _, statErr := os.Stat(filename); statErr != nil {
		tmp, err := os.CreateTemp("", "deg-*.pdf")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.Write(data); err != nil {
			tmp.Close()
			return nil, err
		}
		tmp.Close()
		input = tmp.Name()
	}
	cmd := exec.Command(bin, "-layout", "-enc", "UTF-8", input, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftotext %s: %v: %s", filepath.Base(filename), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
