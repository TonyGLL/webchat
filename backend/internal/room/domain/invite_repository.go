
package domain

import (
	"context"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *Invite) (*Invite, error)
}
