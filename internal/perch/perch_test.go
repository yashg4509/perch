package perch

import "testing"

func TestModulePath(t *testing.T) {
	if ModulePath != "github.com/yashg4509/perch" {
		t.Fatalf("ModulePath = %q, want github.com/yashg4509/perch", ModulePath)
	}
}
