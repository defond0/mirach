// Package plugin houses resources shared by the plugins defined in its
// subpackages.
//
// The only resources shared currently, are the error handling bits that let you
// create custom exceptions so that callers can behave differently in expected
// conditions.
package plugin

// InfoGroup is an interface for getting data and marshaling to json.
type InfoGroup interface {
	GetInfo()
	String() string
}
