package audit

import (
	"log/slog"
	"regexp"
	"strings"
)

// SensitiveKeyPattern matches logger attribute keys that must never log raw values (T053).
// Case-insensitive on the leaf key; covers token, secret, password, client_secret.
var SensitiveKeyPattern = regexp.MustCompile(`(?i)(token|secret|password|client_secret)`)

// RedactingReplaceAttr returns a ReplaceAttr func suitable for [slog.HandlerOptions].
// It runs after any prior ReplaceAttr in attachRedactingReplaceAttr.
func RedactingReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	key := a.Key
	if len(groups) > 0 {
		key = strings.Join(groups, ".") + "." + a.Key
	}
	if SensitiveKeyPattern.MatchString(key) {
		return slog.String(a.Key, "[REDACTED]")
	}
	return a
}

// attachRedactingReplaceAttr wraps opts.ReplaceAttr to apply redaction first.
func attachRedactingReplaceAttr(opts *slog.HandlerOptions) {
	if opts == nil {
		return
	}
	orig := opts.ReplaceAttr
	opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		a = RedactingReplaceAttr(groups, a)
		if orig != nil {
			a = orig(groups, a)
		}
		return a
	}
}

// NewJSONHandler returns a JSON [slog.Handler] with redacting ReplaceAttr (stdlib slog, zap/zerolog-style scrub).
func NewJSONHandler(w slogWriter, opts *slog.HandlerOptions) slog.Handler {
	o := slog.HandlerOptions{}
	if opts != nil {
		o = *opts
	}
	attachRedactingReplaceAttr(&o)
	return slog.NewJSONHandler(w, &o)
}

// slogWriter is the minimal io.Writer slog's JSON handler needs.
type slogWriter interface {
	Write(p []byte) (n int, err error)
}
