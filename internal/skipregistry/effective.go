package skipregistry

import (
	"context"
	"path/filepath"
	"strings"
)

// Decision is the effective skip state of one validation type.
type Decision struct {
	// Skipped reports whether the validation is skipped.
	Skipped bool
	// Source is the registry entry that skips it; empty when not skipped.
	Source DirectoryPath
}

// EffectiveSkips holds the effective lint and test decisions for a directory.
type EffectiveSkips struct {
	Lint Decision
	Test Decision
}

// Effective resolves which validations are skipped for files in dir. An entry
// on dir or on any ancestor up to and including root applies, so a skip at a
// repository root covers packages nested below it, and the nearest entry is
// reported as the source. When dir is not within root, only dir is consulted.
// Entries that cannot be read skip nothing.
func Effective(ctx context.Context, reader Reader, dir, root DirectoryPath) EffectiveSkips {
	if !withinRoot(dir.String(), root.String()) {
		root = dir
	}

	var skips EffectiveSkips
	for current := dir; ; current = DirectoryPath(filepath.Dir(current.String())) {
		if types, err := reader.GetSkipTypes(ctx, current); err == nil {
			skips.apply(current, types)
		}
		if current == root || current.String() == filepath.Dir(current.String()) {
			return skips
		}
	}
}

// apply records the types registered on source for validations not yet decided.
func (s *EffectiveSkips) apply(source DirectoryPath, types []SkipType) {
	for _, t := range types {
		for _, expanded := range ExpandSkipType(t) {
			s.decide(expanded, source)
		}
	}
}

// decide marks skipType as skipped by source unless a nearer entry already did.
func (s *EffectiveSkips) decide(skipType SkipType, source DirectoryPath) {
	var target *Decision
	switch skipType {
	case SkipTypeLint:
		target = &s.Lint
	case SkipTypeTest:
		target = &s.Test
	case SkipTypeAll:
		return
	}
	if target != nil && !target.Skipped {
		*target = Decision{Skipped: true, Source: source}
	}
}

// withinRoot reports whether path is root itself or lies below it.
func withinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
