package slack

import (
	"testing"
)

func TestNotice(t *testing.T) {
	Notice("notification", "This is called in test code.")
}