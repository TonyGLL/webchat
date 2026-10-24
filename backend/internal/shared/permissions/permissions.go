// Package permissions is the single place that decides who may do what.
//
// Room roles, from least to most powerful: member, moderator, admin, owner.
// Global roles: user (no extra power), moderator (acts as a moderator in every room) and
// admin (acts above every room owner). Everything is expressed as a numeric rank so a rule
// reads "the actor must outrank the target".
package permissions

// RoomRole is a user's role inside one room.
type RoomRole string

const (
	RoomMember    RoomRole = "member"
	RoomModerator RoomRole = "moderator"
	RoomAdmin     RoomRole = "admin"
	RoomOwner     RoomRole = "owner"
)

// GlobalRole is a user's role in the whole application.
type GlobalRole string

const (
	GlobalUser      GlobalRole = "user"
	GlobalModerator GlobalRole = "moderator"
	GlobalAdmin     GlobalRole = "admin"
)

// Rank is the numeric power of an actor in a room. Higher outranks lower.
type Rank int

const (
	RankNone      Rank = 0 // not a member and no global power
	RankMember    Rank = 1
	RankModerator Rank = 2
	RankAdmin     Rank = 3
	RankOwner     Rank = 4
	RankSuper     Rank = 5 // global admin: above every room owner
)

// ValidRoomRole reports whether s is a known room role.
func ValidRoomRole(s string) bool {
	switch RoomRole(s) {
	case RoomMember, RoomModerator, RoomAdmin, RoomOwner:
		return true
	}
	return false
}

// ValidGlobalRole reports whether s is a known global role.
func ValidGlobalRole(s string) bool {
	switch GlobalRole(s) {
	case GlobalUser, GlobalModerator, GlobalAdmin:
		return true
	}
	return false
}

func roomRank(role RoomRole) Rank {
	switch role {
	case RoomOwner:
		return RankOwner
	case RoomAdmin:
		return RankAdmin
	case RoomModerator:
		return RankModerator
	case RoomMember:
		return RankMember
	}
	return RankNone
}

// Actor is someone acting in a room: their room role (if they are a member) and global role.
type Actor struct {
	Room     RoomRole
	IsMember bool
	Global   GlobalRole
}

// Rank returns the actor's effective power in the room: the higher of their room role and
// what their global role grants.
func (a Actor) Rank() Rank {
	rank := RankNone
	if a.IsMember {
		rank = roomRank(a.Room)
	}
	switch a.Global {
	case GlobalAdmin:
		return RankSuper
	case GlobalModerator:
		if rank < RankModerator {
			rank = RankModerator
		}
	}
	return rank
}

// TargetRank is the rank of the user being acted upon (global roles included).
func TargetRank(room RoomRole, isMember bool, global GlobalRole) Rank {
	return Actor{Room: room, IsMember: isMember, Global: global}.Rank()
}

// CanRead: any member, or a global moderator/admin (to review reports).
func (a Actor) CanRead() bool { return a.Rank() >= RankMember }

// CanDeleteAnyMessage lets moderators and above delete other people's messages.
func (a Actor) CanDeleteAnyMessage() bool { return a.Rank() >= RankModerator }

// CanResolveReports lets moderators and above handle reports of a room.
func (a Actor) CanResolveReports() bool { return a.Rank() >= RankModerator }

// CanPin lets admins and above pin messages.
func (a Actor) CanPin() bool { return a.Rank() >= RankAdmin }

// CanInvite lets admins and above create invitations.
func (a Actor) CanInvite() bool { return a.Rank() >= RankAdmin }

// CanModerate reports whether the actor may kick or ban a user of the given rank: they need to
// be at least a moderator and strictly outrank the target.
func (a Actor) CanModerate(target Rank) bool {
	return a.Rank() >= RankModerator && a.Rank() > target
}

// CanAssignRole reports whether the actor may set a member of rank target to newRole.
// Admins can create moderators; only owners (or global admins) can create admins; the owner
// role is never assigned this way; nobody can change someone who is not below them.
func (a Actor) CanAssignRole(target Rank, newRole RoomRole) bool {
	rank := a.Rank()
	if rank < RankAdmin || rank <= target {
		return false
	}
	switch newRole {
	case RoomMember, RoomModerator:
		return true
	case RoomAdmin:
		return rank >= RankOwner
	}
	return false
}
