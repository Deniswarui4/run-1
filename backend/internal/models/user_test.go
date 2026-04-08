package models

import (
	"testing"
)

func TestUserHasPermission(t *testing.T) {
	tests := []struct {
		name         string
		userRole     Role
		requiredRole Role
		expected     bool
	}{
		{"Admin has admin permission", RoleAdmin, RoleAdmin, true},
		{"Admin has moderator permission", RoleAdmin, RoleModerator, true},
		{"Admin has organizer permission", RoleAdmin, RoleOrganizer, true},
		{"Admin has attendee permission", RoleAdmin, RoleAttendee, true},
		{"Moderator has moderator permission", RoleModerator, RoleModerator, true},
		{"Moderator has organizer permission", RoleModerator, RoleOrganizer, true},
		{"Moderator has attendee permission", RoleModerator, RoleAttendee, true},
		{"Moderator does not have admin permission", RoleModerator, RoleAdmin, false},
		{"Organizer has organizer permission", RoleOrganizer, RoleOrganizer, true},
		{"Organizer has attendee permission", RoleOrganizer, RoleAttendee, true},
		{"Organizer does not have moderator permission", RoleOrganizer, RoleModerator, false},
		{"Attendee has attendee permission", RoleAttendee, RoleAttendee, true},
		{"Attendee does not have organizer permission", RoleAttendee, RoleOrganizer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.userRole}
			result := user.HasPermission(tt.requiredRole)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUserIsAdmin(t *testing.T) {
	adminUser := &User{Role: RoleAdmin}
	if !adminUser.IsAdmin() {
		t.Error("Admin user should return true for IsAdmin()")
	}

	moderatorUser := &User{Role: RoleModerator}
	if moderatorUser.IsAdmin() {
		t.Error("Moderator user should return false for IsAdmin()")
	}
}

func TestUserIsModerator(t *testing.T) {
	adminUser := &User{Role: RoleAdmin}
	if !adminUser.IsModerator() {
		t.Error("Admin user should return true for IsModerator()")
	}

	moderatorUser := &User{Role: RoleModerator}
	if !moderatorUser.IsModerator() {
		t.Error("Moderator user should return true for IsModerator()")
	}

	organizerUser := &User{Role: RoleOrganizer}
	if organizerUser.IsModerator() {
		t.Error("Organizer user should return false for IsModerator()")
	}
}

func TestUserIsOrganizer(t *testing.T) {
	adminUser := &User{Role: RoleAdmin}
	if !adminUser.IsOrganizer() {
		t.Error("Admin user should return true for IsOrganizer()")
	}

	moderatorUser := &User{Role: RoleModerator}
	if !moderatorUser.IsOrganizer() {
		t.Error("Moderator user should return true for IsOrganizer()")
	}

	organizerUser := &User{Role: RoleOrganizer}
	if !organizerUser.IsOrganizer() {
		t.Error("Organizer user should return true for IsOrganizer()")
	}

	attendeeUser := &User{Role: RoleAttendee}
	if attendeeUser.IsOrganizer() {
		t.Error("Attendee user should return false for IsOrganizer()")
	}
}
