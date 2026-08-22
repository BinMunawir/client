package access_activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BinMunawir/maal_business/internal/access/store"
	"github.com/BinMunawir/maal_business/internal/core"
	"github.com/go-jet/jet/v2/qrm"
)

type ValidateInvitationInput struct {
	Token string
}

// ValidateInvitation loads the invitation by its single-use token and checks it is pending
// and unexpired — the guard before a user is created. A missing token surfaces as
// qrm.ErrNoRows and is classified explicitly (standard §6.2).
func ValidateInvitation(ctx context.Context, in ValidateInvitationInput) (core.Invitation, error) {
	inv, err := store.FetchInvitationByToken(ctx, in.Token)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return core.Invitation{}, fmt.Errorf("ValidateInvitation: no invitation for token: %w", err)
		}
		return core.Invitation{}, fmt.Errorf("store.FetchInvitationByToken: %w", err)
	}
	if inv.Status != core.EnumInvitationStatusPending {
		return inv, fmt.Errorf("ValidateInvitation: invitation is %s, not pending", inv.Status)
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		return inv, fmt.Errorf("ValidateInvitation: invitation expired at %s", inv.ExpiresAt.Format(time.RFC3339))
	}
	return inv, nil
}
