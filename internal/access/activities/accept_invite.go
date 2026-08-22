package access_activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/BinMunawir/maal_business/internal/access/store"
	"github.com/BinMunawir/maal_business/internal/core"
	"github.com/go-jet/jet/v2/qrm"
)

type AcceptInviteInput struct {
	Invitation core.Invitation
}

// AcceptInvite transitions the Invitation pending → accepted (design §7), the durable
// close-out of the acceptance flow. Skeleton: persist → classify (standard §6.2).
func AcceptInvite(ctx context.Context, in AcceptInviteInput) (core.Invitation, error) {
	inv, err := acceptInvitePersist(ctx, in.Invitation)
	if err != nil {
		return in.Invitation, fmt.Errorf("acceptInvitePersist: %w", err)
	}
	return inv, nil
}

func acceptInvitePersist(ctx context.Context, inv core.Invitation) (core.Invitation, error) {
	updated, err := store.AcceptInvitation(ctx, inv)
	if err == nil {
		return updated, nil
	}
	if errors.Is(err, qrm.ErrNoRows) {
		// Guarded pending→accepted matched no row: already accepted (idempotent replay). Re-read
		// and return it as a no-op success (standard §6.2 classify + §8 no-op).
		current, ferr := store.FetchInvitationByToken(ctx, inv.Token)
		if ferr != nil {
			return inv, fmt.Errorf("acceptInvitePersist: refetch after guarded no-op: %w", ferr)
		}
		if current.Status == core.EnumInvitationStatusAccepted {
			return current, nil
		}
		return inv, fmt.Errorf("acceptInvitePersist: guarded accept matched no row, status is %s: %w", current.Status, err)
	}
	return inv, fmt.Errorf("store.AcceptInvitation: %w", err)
}
