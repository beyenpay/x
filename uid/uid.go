// Package uid provides helpers for generating unique identifiers.
package uid

import (
	"encoding/hex"

	"github.com/google/uuid"
)

// separator joins a prefix and the identifier body.
const separator = "_"

// New returns a random (UUID v4) identifier as 32 lowercase hex characters,
// without hyphens.
//
//	7daf3804ff224739b925b8a61b5cc550
func New() string {
	u := uuid.New()
	return hex.EncodeToString(u[:])
}

// WithPrefix returns New() prefixed with the given prefix and an underscore.
// An empty prefix is equivalent to New().
//
//	ord_7daf3804ff224739b925b8a61b5cc550
func WithPrefix(prefix string) string {
	if prefix == "" {
		return New()
	}
	return prefix + separator + New()
}

// NewSortable returns a time-ordered (UUID v7) identifier as 32 lowercase hex
// characters, without hyphens. IDs generated later sort after earlier ones,
// which keeps database index inserts mostly sequential.
//
//	019a3c5e7b2a7c4d8e1f0a2b3c4d5e6f
func NewSortable() string {
	u := uuid.Must(uuid.NewV7())
	return hex.EncodeToString(u[:])
}

// WithPrefixSortable returns NewSortable() prefixed with the given prefix and
// an underscore. An empty prefix is equivalent to NewSortable().
//
//	ord_019a3c5e7b2a7c4d8e1f0a2b3c4d5e6f
func WithPrefixSortable(prefix string) string {
	if prefix == "" {
		return NewSortable()
	}
	return prefix + separator + NewSortable()
}
