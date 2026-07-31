package berita

import (
	"regexp"
	"strings"
)

var imageRefPattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)(?:[^)]*)\)`)

func isInternalImageKey(ref string) bool {
	return strings.HasPrefix(ref, "berita/") && !strings.Contains(ref, "://")
}

func eachImageRef(content string, fn func(ref string) (string, bool)) string {
	matches := imageRefPattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var sb strings.Builder
	last := 0
	for _, m := range matches {
		fullStart, fullEnd := m[0], m[1]
		refStart, refEnd := m[2], m[3]
		ref := content[refStart:refEnd]

		sb.WriteString(content[last:fullStart])
		if repl, ok := fn(ref); ok {
			sb.WriteString(content[fullStart:refStart])
			sb.WriteString(repl)
			sb.WriteString(content[refEnd:fullEnd])
		} else {
			sb.WriteString(content[fullStart:fullEnd])
		}
		last = fullEnd
	}
	sb.WriteString(content[last:])
	return sb.String()
}

func extractImageKeys(content string) []string {
	var keys []string
	for _, m := range imageRefPattern.FindAllStringSubmatch(content, -1) {
		ref := m[1]
		if isInternalImageKey(ref) {
			keys = append(keys, ref)
		}
	}
	return keys
}

func normalizeImageRefs(content string) (string, error) {
	var firstErr error
	out := eachImageRef(content, func(ref string) (string, bool) {
		path := ref
		if strings.Contains(ref, "://") {
			p, err := extractObjectPath(ref)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return "", false
			}
			path = p
		}
		if !strings.HasPrefix(path, "berita/") {
			return "", false
		}
		return path, true
	})
	return out, firstErr
}

func resolveImageRefs(content string, sign func(key string) (string, error)) string {
	return eachImageRef(content, func(ref string) (string, bool) {
		if !isInternalImageKey(ref) {
			return "", false
		}
		signed, err := sign(ref)
		if err != nil {
			return "", false
		}
		return signed, true
	})
}
