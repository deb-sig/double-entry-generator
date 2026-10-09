package reader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"path/filepath"
	"strings"
)

// The eml reader is a container: it picks one MIME part of a saved e-mail
// and hands its bytes to an inner reader. `part` is a media type
// ("text/html", "text/plain") or "attachment:<glob>" matched against the
// attachment file name ("attachment:*.pdf").
func readEML(data []byte, cfg Config) (Table, error) {
	if cfg.Inner == nil {
		return Table{}, fmt.Errorf("eml reader requires `inner` (the reader for the selected part)")
	}
	if strings.TrimSpace(cfg.Part) == "" {
		return Table{}, fmt.Errorf("eml reader requires `part` (text/html, text/plain or attachment:<glob>)")
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return Table{}, fmt.Errorf("parse eml: %w", err)
	}
	var found *mimePart
	err = walkParts(msg.Header.Get("Content-Type"), msg.Header.Get("Content-Transfer-Encoding"), msg.Header.Get("Content-Disposition"), msg.Body, func(p mimePart) bool {
		if matchPart(cfg.Part, p) {
			found = &p
			return true
		}
		return false
	})
	if err != nil {
		return Table{}, err
	}
	if found == nil {
		return Table{}, fmt.Errorf("eml: no part matches %q", cfg.Part)
	}
	inner := *cfg.Inner
	if inner.Encoding == "" {
		inner.Encoding = found.charset
	}
	name := found.filename
	if name == "" {
		name = "part." + extensionFor(found.mediaType, inner.Format)
	}
	return ReadBytes(name, found.body, inner)
}

type mimePart struct {
	mediaType string
	charset   string
	filename  string
	body      []byte
}

// walkParts descends multipart bodies depth-first and calls visit on each
// leaf. visit returns true to stop.
func walkParts(contentType, transferEncoding, disposition string, body io.Reader, visit func(mimePart) bool) error {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = "text/plain"
		params = map[string]string{}
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return fmt.Errorf("eml: multipart without boundary")
		}
		mr := multipart.NewReader(body, boundary)
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("eml: read part: %w", err)
			}
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, p); err != nil {
				return err
			}
			// multipart.Part already decodes quoted-printable; base64 it does not.
			encoding := p.Header.Get("Content-Transfer-Encoding")
			if strings.EqualFold(encoding, "quoted-printable") {
				encoding = ""
			}
			stop := false
			sub := func(mp mimePart) bool { stop = visit(mp); return stop }
			if err := walkParts(p.Header.Get("Content-Type"), encoding, p.Header.Get("Content-Disposition"), &buf, sub); err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
	}
	raw, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	decoded, err := decodeTransfer(raw, transferEncoding)
	if err != nil {
		return err
	}
	filename := params["name"]
	if _, dparams, err := mime.ParseMediaType(disposition); err == nil && dparams["filename"] != "" {
		filename = dparams["filename"]
	}
	if decoded, err := new(mime.WordDecoder).DecodeHeader(filename); err == nil {
		filename = decoded
	}
	visit(mimePart{mediaType: mediaType, charset: params["charset"], filename: filename, body: decoded})
	return nil
}

func decodeTransfer(raw []byte, encoding string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		clean := bytes.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, raw)
		out := make([]byte, base64.StdEncoding.DecodedLen(len(clean)))
		n, err := base64.StdEncoding.Decode(out, clean)
		if err != nil {
			// Tolerate missing padding.
			n, err = base64.RawStdEncoding.Decode(out, bytes.TrimRight(clean, "="))
			if err != nil {
				return nil, fmt.Errorf("eml: base64: %w", err)
			}
		}
		return out[:n], nil
	case "quoted-printable":
		return io.ReadAll(quotedprintable.NewReader(bytes.NewReader(raw)))
	default:
		return raw, nil
	}
}

func matchPart(selector string, p mimePart) bool {
	selector = strings.TrimSpace(selector)
	if glob, ok := strings.CutPrefix(selector, "attachment:"); ok {
		if p.filename == "" {
			return false
		}
		matched, err := filepath.Match(strings.ToLower(strings.TrimSpace(glob)), strings.ToLower(p.filename))
		return err == nil && matched
	}
	return strings.EqualFold(selector, p.mediaType) && p.filename == ""
}

func extensionFor(mediaType, format string) string {
	switch {
	case strings.Contains(mediaType, "html"):
		return "html"
	case strings.Contains(mediaType, "pdf"):
		return "pdf"
	case format != "":
		return NormalizeFormat(format)
	default:
		return "txt"
	}
}
