package auth

import (
	"slices"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
)

// TestButtonCodes guards the seam between stored button codes and the codes the
// frontend compares in hasAuth (add-auth="manage:user:add" etc.): the seed data
// stores button codes with the API version prefix ("v1:manage:user:add"), while
// the frontend expects bare codes, so the prefix must be stripped before the
// codes are returned to the UI.
func TestButtonCodes(t *testing.T) {
	menus := []*manage_menu.ManageMenu{
		{ButtonCode: "v1:manage:user:add"},
		{ButtonCode: "v1:manage:user:edit"},
		// codes stored without the prefix must pass through unchanged
		{ButtonCode: "manage:role:add"},
	}

	got := buttonCodes(menus)
	want := []string{"manage:user:add", "manage:user:edit", "manage:role:add"}

	if !slices.Equal(got, want) {
		t.Errorf("buttonCodes() = %v, want %v", got, want)
	}
}
