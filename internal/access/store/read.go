package store

import (
	"context"
	"fmt"

	"github.com/BinMunawir/client/internal/adapters/jet"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/table"
	"github.com/BinMunawir/client/internal/core"
	"github.com/go-jet/jet/v2/postgres"
)

func FetchInvitationByToken(ctx context.Context, token string) (core.Invitation, error) {
	stmt := table.Invitations.
		SELECT(table.Invitations.AllColumns).
		WHERE(table.Invitations.Token.EQ(postgres.String(token)))

	var m model.Invitations
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Invitation{}, fmt.Errorf("pg: fetch invitation by token: %w", err)
	}
	return InvitationFromModel(m), nil
}

func FetchInvitationByCorrID(ctx context.Context, corrID string) (core.Invitation, error) {
	stmt := table.Invitations.
		SELECT(table.Invitations.AllColumns).
		WHERE(table.Invitations.CorrelationID.EQ(postgres.String(corrID)))

	var m model.Invitations
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Invitation{}, fmt.Errorf("pg: fetch invitation by corr_id %q: %w", corrID, err)
	}
	return InvitationFromModel(m), nil
}

func FetchMembership(ctx context.Context, id string) (core.Membership, error) {
	stmt := table.Memberships.
		SELECT(table.Memberships.AllColumns).
		WHERE(table.Memberships.ID.EQ(postgres.String(id)))

	var m model.Memberships
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Membership{}, fmt.Errorf("pg: fetch membership %q: %w", id, err)
	}
	return MembershipFromModel(m), nil
}

func FetchMembershipByCorrID(ctx context.Context, corrID string) (core.Membership, error) {
	stmt := table.Memberships.
		SELECT(table.Memberships.AllColumns).
		WHERE(table.Memberships.CorrelationID.EQ(postgres.String(corrID)))

	var m model.Memberships
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Membership{}, fmt.Errorf("pg: fetch membership by corr_id %q: %w", corrID, err)
	}
	return MembershipFromModel(m), nil
}

func FetchServiceAccountByCorrID(ctx context.Context, corrID string) (core.ServiceAccount, error) {
	stmt := table.ServiceAccounts.
		SELECT(table.ServiceAccounts.AllColumns).
		WHERE(table.ServiceAccounts.CorrelationID.EQ(postgres.String(corrID)))

	var m model.ServiceAccounts
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.ServiceAccount{}, fmt.Errorf("pg: fetch service_account by corr_id %q: %w", corrID, err)
	}
	return ServiceAccountFromModel(m), nil
}

// FetchMembershipsByBusiness returns a page of the operators of a business — "an account has
// many operators" (design §6.1). Count is derived from the base statement (standard §6.3).
func FetchMembershipsByBusiness(ctx context.Context, businessID string, limit, offset int) (pg.Pagination[core.Membership], error) {
	stmt := postgres.SELECT(table.Memberships.AllColumns).
		FROM(table.Memberships).
		WHERE(table.Memberships.BusinessID.EQ(postgres.String(businessID)))

	page := pg.Pagination[core.Membership]{}
	stmtCount := postgres.SELECT(postgres.COUNT(postgres.STAR).AS("total")).
		FROM(stmt.AsTable(table.Memberships.TableName()))
	if err := stmtCount.QueryContext(ctx, pg.DB(), &page); err != nil {
		return pg.Pagination[core.Membership]{}, fmt.Errorf("pg: count memberships: %w", err)
	}
	page.Limit, page.Offset = page.NormalizePaging(limit, offset)

	if page.Total == 0 || int64(offset) >= page.Total {
		return page, nil
	}

	stmtRows := stmt.
		ORDER_BY(table.Memberships.CreatedAt.DESC()).
		LIMIT(int64(page.Limit)).
		OFFSET(int64(page.Offset))

	var mdls []model.Memberships
	if err := stmtRows.QueryContext(ctx, pg.DB(), &mdls); err != nil {
		return pg.Pagination[core.Membership]{}, fmt.Errorf("pg: list memberships: %w", err)
	}

	page.Items = make([]core.Membership, len(mdls))
	for i, m := range mdls {
		page.Items[i] = MembershipFromModel(m)
	}
	return page, nil
}
