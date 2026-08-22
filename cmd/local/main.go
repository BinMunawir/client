package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/BinMunawir/client/config"
	accessworkflows "github.com/BinMunawir/client/internal/access/workflows"
	workflows "github.com/BinMunawir/client/internal/client/workflows"
)

// local drives the pure-function twins directly — no Temporal, no worker (standard §3, §6.1).
// It walks the whole domain end-to-end against just Postgres (Keycloak falls back to dev
// stubs when unconfigured): onboard a Business, invite an operator, accept the invite, and
// provision a machine operator. This is the fast local-iteration / smoke-run seam.
func main() {
	fmt.Println("local app is running...")
	config.Load()

	ctx := context.Background()
	run := strconv.Itoa(rand.IntN(99999999))

	// 1. Onboard a Business: Register (draft) → ProvisionOrg (Keycloak org) → Activate.
	onb, err := workflows.OnboardLocal(ctx, workflows.OnboardInput{
		CorrID:      "onboard-" + run,
		LegalName:   "Acme Trading Company LLC",
		TradeName:   "Acme",
		CRNumber:    "1010101010",
		LegalForm:   "Company",
		SizeSegment: "sme",
		ServiceTier: "gold",
	})
	if err != nil {
		log.Fatalf("OnboardLocal: %v", err)
	}
	biz := onb.Business
	fmt.Printf("1) onboarded business  : %s (%s) status=%s keycloak_org=%s\n",
		biz.ID, biz.LegalName, biz.Status, deref(biz.KeycloakOrgID))

	// 2. Invite a finance operator (the outsourced-accountant scenario: operates, owns nothing).
	inv, err := accessworkflows.InviteLocal(ctx, accessworkflows.InviteInput{
		CorrID:       "invite-" + run,
		BusinessID:   biz.ID,
		InvitedEmail: "clerk-" + run + "@example.com",
		IntendedRole: "FinanceOperator",
		InvitedBy:    "act-system",
	})
	if err != nil {
		log.Fatalf("InviteLocal: %v", err)
	}
	fmt.Printf("2) issued invitation   : %s → %s (token %s…)\n",
		inv.Invitation.ID, inv.Invitation.InvitedEmail, inv.Invitation.Token[:8])

	// 3. Accept: Validate → ProvisionUser (Keycloak sub) → MaterializeMembership → AcceptInvite.
	acc, err := accessworkflows.AcceptInvitationLocal(ctx, accessworkflows.AcceptInvitationInput{
		Token: inv.Invitation.Token,
	})
	if err != nil {
		log.Fatalf("AcceptInvitationLocal: %v", err)
	}
	fmt.Printf("3) materialized member : %s actor=%s role=%s sub=%s person_id=%s (seam)\n",
		acc.Membership.ID, acc.Membership.ActorID, acc.Membership.Role, acc.Membership.KeycloakSub, deref(acc.Membership.PersonID))

	// 4. Provision a machine operator (topology 1): pure persistence, no external calls.
	sa, err := accessworkflows.ProvisionServiceAccountLocal(ctx, accessworkflows.ProvisionServiceAccountInput{
		CorrID:     "svc-" + run,
		BusinessID: biz.ID,
		Label:      "nightly-payments-bot",
	})
	if err != nil {
		log.Fatalf("ProvisionServiceAccountLocal: %v", err)
	}
	fmt.Printf("4) provisioned machine : %s actor=%s status=%s\n",
		sa.ServiceAccount.ID, sa.ServiceAccount.ActorID, sa.ServiceAccount.Status)

	fmt.Println("local end-to-end complete.")
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
