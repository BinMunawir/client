package store

import (
	"context"
	"fmt"
	"time"

	"github.com/BinMunawir/client/internal/adapters/jet"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/table"
	"github.com/BinMunawir/client/internal/core"
	"github.com/go-jet/jet/v2/postgres"
)

// --- Inserts ---------------------------------------------------------------

func InsertActor(ctx context.Context, a core.Actor) (core.Actor, error) {
	stmt := table.Actors.
		INSERT(table.Actors.AllColumns).
		MODEL(ActorToModel(a)).
		RETURNING(table.Actors.AllColumns)

	var m model.Actors
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return a, fmt.Errorf("stmt.QueryContext %q: %w", a.ID, err)
	}
	return ActorFromModel(m), nil
}

func InsertMembership(ctx context.Context, mb core.Membership) (core.Membership, error) {
	stmt := table.Memberships.
		INSERT(table.Memberships.AllColumns).
		MODEL(MembershipToModel(mb)).
		RETURNING(table.Memberships.AllColumns)

	var m model.Memberships
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return mb, fmt.Errorf("stmt.QueryContext %q: %w", mb.ID, err)
	}
	return MembershipFromModel(m), nil
}

func InsertInvitation(ctx context.Context, iv core.Invitation) (core.Invitation, error) {
	stmt := table.Invitations.
		INSERT(table.Invitations.AllColumns).
		MODEL(InvitationToModel(iv)).
		RETURNING(table.Invitations.AllColumns)

	var m model.Invitations
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return iv, fmt.Errorf("stmt.QueryContext %q: %w", iv.ID, err)
	}
	return InvitationFromModel(m), nil
}

func InsertServiceAccount(ctx context.Context, sa core.ServiceAccount) (core.ServiceAccount, error) {
	stmt := table.ServiceAccounts.
		INSERT(table.ServiceAccounts.AllColumns).
		MODEL(ServiceAccountToModel(sa)).
		RETURNING(table.ServiceAccounts.AllColumns)

	var m model.ServiceAccounts
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return sa, fmt.Errorf("stmt.QueryContext %q: %w", sa.ID, err)
	}
	return ServiceAccountFromModel(m), nil
}

// --- Invitation transitions (design §7) ------------------------------------

func AcceptInvitation(ctx context.Context, iv core.Invitation) (core.Invitation, error) {
	return updateInvitationStatus(ctx, iv, core.EnumInvitationStatusAccepted)
}

func ExpireInvitation(ctx context.Context, iv core.Invitation) (core.Invitation, error) {
	return updateInvitationStatus(ctx, iv, core.EnumInvitationStatusExpired)
}

func RevokeInvitation(ctx context.Context, iv core.Invitation) (core.Invitation, error) {
	return updateInvitationStatus(ctx, iv, core.EnumInvitationStatusRevoked)
}

func updateInvitationStatus(ctx context.Context, iv core.Invitation, to core.EnumInvitationStatus) (core.Invitation, error) {
	stmt := table.Invitations.UPDATE().SET(
		table.Invitations.Status.SET(postgres.String(string(to))),
		table.Invitations.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	).WHERE(
		table.Invitations.ID.EQ(postgres.String(iv.ID)).
			AND(table.Invitations.Status.NOT_EQ(postgres.String(string(to)))),
	).RETURNING(table.Invitations.AllColumns)

	var m model.Invitations
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return iv, fmt.Errorf("QueryContext: %w", err)
	}
	return InvitationFromModel(m), nil
}

// --- Membership transitions (design §7) ------------------------------------

func SuspendMembership(ctx context.Context, mb core.Membership) (core.Membership, error) {
	return updateMembershipStatus(ctx, mb, core.EnumMembershipStatusSuspended)
}

func ReactivateMembership(ctx context.Context, mb core.Membership) (core.Membership, error) {
	return updateMembershipStatus(ctx, mb, core.EnumMembershipStatusActive)
}

func RevokeMembership(ctx context.Context, mb core.Membership) (core.Membership, error) {
	return updateMembershipStatus(ctx, mb, core.EnumMembershipStatusRevoked)
}

func updateMembershipStatus(ctx context.Context, mb core.Membership, to core.EnumMembershipStatus) (core.Membership, error) {
	stmt := table.Memberships.UPDATE().SET(
		table.Memberships.Status.SET(postgres.String(string(to))),
		table.Memberships.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	).WHERE(
		table.Memberships.ID.EQ(postgres.String(mb.ID)).
			AND(table.Memberships.Status.NOT_EQ(postgres.String(string(to)))),
	).RETURNING(table.Memberships.AllColumns)

	var m model.Memberships
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return mb, fmt.Errorf("QueryContext: %w", err)
	}
	return MembershipFromModel(m), nil
}

// --- ServiceAccount transitions --------------------------------------------

func SuspendServiceAccount(ctx context.Context, sa core.ServiceAccount) (core.ServiceAccount, error) {
	return updateServiceAccountStatus(ctx, sa, core.EnumServiceAccountStatusSuspended)
}

func RevokeServiceAccount(ctx context.Context, sa core.ServiceAccount) (core.ServiceAccount, error) {
	return updateServiceAccountStatus(ctx, sa, core.EnumServiceAccountStatusRevoked)
}

func updateServiceAccountStatus(ctx context.Context, sa core.ServiceAccount, to core.EnumServiceAccountStatus) (core.ServiceAccount, error) {
	stmt := table.ServiceAccounts.UPDATE().SET(
		table.ServiceAccounts.Status.SET(postgres.String(string(to))),
		table.ServiceAccounts.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	).WHERE(
		table.ServiceAccounts.ID.EQ(postgres.String(sa.ID)).
			AND(table.ServiceAccounts.Status.NOT_EQ(postgres.String(string(to)))),
	).RETURNING(table.ServiceAccounts.AllColumns)

	var m model.ServiceAccounts
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return sa, fmt.Errorf("QueryContext: %w", err)
	}
	return ServiceAccountFromModel(m), nil
}
