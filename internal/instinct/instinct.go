// Package instinct manages atomic learned behaviors with confidence scoring.
package instinct

import "time"

// Instinct represents a single learned behavior.
type Instinct struct {
	ID         string    `json:"id"                    yaml:"id"`
	Trigger    string    `json:"trigger"               yaml:"trigger"`
	Confidence float64   `json:"confidence"            yaml:"confidence"`
	Domain     string    `json:"domain"                yaml:"domain"`
	Source     string    `json:"source"                yaml:"source"`
	SourceRepo string    `json:"source_repo,omitempty" yaml:"source_repo,omitempty"`
	Content    string    `json:"content,omitempty"     yaml:"-"`
	CreatedAt  time.Time `json:"created_at"            yaml:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"            yaml:"updated_at"`
}

// ListOptions filters instinct listing.
type ListOptions struct {
	Domain        string
	MinConfidence float64
	Source        string
}
