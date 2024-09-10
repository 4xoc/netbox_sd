// Package util provides common utilities across the netbox_sd code base and used within different modules.
package util

// NewPtr is a helper function that returns the value v as ptr to itself
func NewPtr[T any](v T) *T {
	return &v
}
