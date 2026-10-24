package permissions

import "testing"

func member(role RoomRole) Actor { return Actor{Room: role, IsMember: true, Global: GlobalUser} }

func TestRankCombinesRoomAndGlobalRoles(t *testing.T) {
	cases := []struct {
		name string
		a    Actor
		want Rank
	}{
		{"not a member", Actor{Global: GlobalUser}, RankNone},
		{"member", member(RoomMember), RankMember},
		{"room moderator", member(RoomModerator), RankModerator},
		{"room admin", member(RoomAdmin), RankAdmin},
		{"owner", member(RoomOwner), RankOwner},
		{"global moderator outside the room", Actor{Global: GlobalModerator}, RankModerator},
		{"global moderator never lowers a room admin", Actor{Room: RoomAdmin, IsMember: true, Global: GlobalModerator}, RankAdmin},
		{"global admin outranks an owner", Actor{Room: RoomOwner, IsMember: true, Global: GlobalAdmin}, RankSuper},
		{"unknown room role is nothing", Actor{Room: "wizard", IsMember: true, Global: GlobalUser}, RankNone},
	}
	for _, c := range cases {
		if got := c.a.Rank(); got != c.want {
			t.Errorf("%s: rank = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestCapabilities(t *testing.T) {
	cases := []struct {
		name                            string
		a                               Actor
		del, resolve, pin, invite, read bool
	}{
		{"stranger", Actor{Global: GlobalUser}, false, false, false, false, false},
		{"member", member(RoomMember), false, false, false, false, true},
		{"moderator", member(RoomModerator), true, true, false, false, true},
		{"admin", member(RoomAdmin), true, true, true, true, true},
		{"owner", member(RoomOwner), true, true, true, true, true},
		{"global moderator", Actor{Global: GlobalModerator}, true, true, false, false, true},
		{"global admin", Actor{Global: GlobalAdmin}, true, true, true, true, true},
	}
	for _, c := range cases {
		got := []bool{c.a.CanDeleteAnyMessage(), c.a.CanResolveReports(), c.a.CanPin(), c.a.CanInvite(), c.a.CanRead()}
		want := []bool{c.del, c.resolve, c.pin, c.invite, c.read}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("%s: capability #%d = %v, want %v", c.name, i, got[i], want[i])
			}
		}
	}
}

func TestCanModerateRequiresOutranking(t *testing.T) {
	if member(RoomMember).CanModerate(RankNone) {
		t.Error("a plain member cannot moderate anyone")
	}
	if !member(RoomModerator).CanModerate(RankMember) {
		t.Error("moderator should moderate a member")
	}
	if member(RoomModerator).CanModerate(RankModerator) || member(RoomModerator).CanModerate(RankAdmin) {
		t.Error("moderator must not moderate peers or superiors")
	}
	if !member(RoomAdmin).CanModerate(RankModerator) || member(RoomAdmin).CanModerate(RankOwner) {
		t.Error("admin moderates below, never the owner")
	}
	if !(Actor{Global: GlobalAdmin}).CanModerate(RankOwner) {
		t.Error("global admin outranks owners")
	}
}

func TestCanAssignRole(t *testing.T) {
	cases := []struct {
		name   string
		a      Actor
		target Rank
		role   RoomRole
		want   bool
	}{
		{"member cannot assign", member(RoomMember), RankNone, RoomModerator, false},
		{"moderator cannot assign", member(RoomModerator), RankMember, RoomModerator, false},
		{"admin makes a moderator", member(RoomAdmin), RankMember, RoomModerator, true},
		{"admin demotes a moderator", member(RoomAdmin), RankModerator, RoomMember, true},
		{"admin cannot make an admin", member(RoomAdmin), RankMember, RoomAdmin, false},
		{"admin cannot touch another admin", member(RoomAdmin), RankAdmin, RoomMember, false},
		{"owner makes an admin", member(RoomOwner), RankMember, RoomAdmin, true},
		{"owner demotes an admin", member(RoomOwner), RankAdmin, RoomMember, true},
		{"nobody assigns owner", member(RoomOwner), RankMember, RoomOwner, false},
		{"global admin makes an admin", Actor{Global: GlobalAdmin}, RankMember, RoomAdmin, true},
		{"unknown role refused", member(RoomOwner), RankMember, "wizard", false},
	}
	for _, c := range cases {
		if got := c.a.CanAssignRole(c.target, c.role); got != c.want {
			t.Errorf("%s: = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestValidRoles(t *testing.T) {
	if !ValidRoomRole("moderator") || ValidRoomRole("god") || !ValidGlobalRole("admin") || ValidGlobalRole("owner") {
		t.Error("role validation is wrong")
	}
}
