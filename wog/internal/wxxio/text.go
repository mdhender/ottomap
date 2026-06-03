package wxxio

import (
	"strings"

	"github.com/mdhender/ottomap"
)

// Escape XML-escapes attribute values.
func Escape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// XMLChars escapes character data (less aggressive than attribute escaping).
func XMLChars(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return r.Replace(s)
}

// SplitTags parses Worldographer's comma-separated tag attribute into a
// slice. Empty input yields a nil slice.
func SplitTags(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ScopeFor maps Worldographer's isContinent/isKingdom/isProvince trio onto
// the domain Scope. The most-specific true flag wins.
func ScopeFor(continent, kingdom, province bool) ottomap.Scope {
	switch {
	case province:
		return ottomap.ScopeProvince
	case kingdom:
		return ottomap.ScopeKingdom
	case continent:
		return ottomap.ScopeContinent
	default:
		return ottomap.ScopeWorld
	}
}

// ScopeFlag reports whether a label at the given Scope should be visible at
// the named zoom level ("world", "continent", "kingdom", "province").
// Worldographer's nested model means a Continent-scope label is visible at
// World and Continent.
func ScopeFlag(s ottomap.Scope, level string) bool {
	switch s {
	case ottomap.ScopeProvince:
		return level == "province" || level == "kingdom" || level == "continent" || level == "world"
	case ottomap.ScopeKingdom:
		return level == "kingdom" || level == "continent" || level == "world"
	case ottomap.ScopeContinent:
		return level == "continent" || level == "world"
	default:
		return level == "world"
	}
}

// ScopeFlagString returns "true" or "false" for use directly inside an
// attribute value.
func ScopeFlagString(s ottomap.Scope, level string) string {
	if ScopeFlag(s, level) {
		return "true"
	}
	return "false"
}
