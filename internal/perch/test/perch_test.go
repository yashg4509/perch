package perch_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/perch"
)

func TestModulePath(t *testing.T) {
	if perch.ModulePath != "github.com/yashg4509/perch" {
		t.Fatalf("ModulePath = %q, want github.com/yashg4509/perch", perch.ModulePath)
	}
}
