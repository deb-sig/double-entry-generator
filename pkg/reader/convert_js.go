//go:build js

package reader

import "fmt"

// In the browser there is no process to run. The web front end converts
// PDF to layout text with pdf.js and hands the text to the reader.
func convert(name, filename string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("converter %q is not available in the browser; convert the pdf to text first", name)
}
