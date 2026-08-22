package access_activities

import (
	"context"
	"fmt"
	"time"

	"github.com/BinMunawir/client/internal/access/store"
	"github.com/BinMunawir/client/internal/core"
)

type CreateServiceAccountInput struct {
	CorrID     string
	BusinessID string
	Label      string
}

// CreateServiceAccount provisions a machine operator scoped to one Business — topology 1
// (design §4.5). It creates the Actor (machine concrete form, §4.1) then the ServiceAccount
// that backs it. The credential secret is minted/stored elsewhere; this domain holds only a
// reference to it. Skeleton: map → persist → classify (standard §6.2).
func CreateServiceAccount(ctx context.Context, in CreateServiceAccountInput) (core.ServiceAccount, error) {
	actor, sa := createServiceAccountEntity(in)
	sa, err := createServiceAccountPersist(ctx, actor, sa)
	if err != nil {
		return core.ServiceAccount{}, fmt.Errorf("createServiceAccountPersist: %w", err)
	}
	return sa, nil
}

func createServiceAccountEntity(in CreateServiceAccountInput) (core.Actor, core.ServiceAccount) {
	now := time.Now().UTC()
	actor := core.Actor{
		ID:        core.ActorID(),
		Type:      core.EnumActorTypeServiceAccount,
		UpdatedAt: now,
		CreatedAt: now,
	}
	// A reference to secret material minted outside this domain (the secret itself is never
	// stored here). A real build resolves this from a secrets manager / Keycloak client.
	credentialRef := "cred-ref-" + actor.ID
	sa := core.ServiceAccount{
		ID:            core.ServiceAccountID(),
		CorrID:        in.CorrID,
		ActorID:       actor.ID,
		BusinessID:    in.BusinessID,
		Label:         in.Label,
		Status:        core.EnumServiceAccountStatusActive,
		CredentialRef: &credentialRef,
		UpdatedAt:     now,
		CreatedAt:     now,
	}
	return actor, sa
}

func createServiceAccountPersist(ctx context.Context, actor core.Actor, sa core.ServiceAccount) (core.ServiceAccount, error) {
	if _, err := store.InsertActor(ctx, actor); err != nil {
		if !isUniqueViolation(err) {
			return sa, fmt.Errorf("store.InsertActor: %w", err)
		}
	}

	inserted, err := store.InsertServiceAccount(ctx, sa)
	if err != nil {
		if isUniqueViolation(err) {
			// Idempotent replay: already provisioned for this correlation key (standard §8).
			existing, ferr := store.FetchServiceAccountByCorrID(ctx, sa.CorrID)
			if ferr != nil {
				return sa, fmt.Errorf("createServiceAccountPersist: reconcile: %w", ferr)
			}
			return existing, nil
		}
		return sa, fmt.Errorf("store.InsertServiceAccount: %w", err)
	}
	return inserted, nil
}
