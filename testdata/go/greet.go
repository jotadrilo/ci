// Package greet is the fixture the reusable Go workflow is tested against.
package greet

import "fmt"

// Hello returns a greeting for name.
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
