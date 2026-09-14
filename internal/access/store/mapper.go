package store

import (
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/core"
)

func ActorToModel(a core.Actor) model.Actors {
	return model.Actors{
		ID:        a.ID,
		Type:      string(a.Type),
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.CreatedAt,
	}
}

func ActorFromModel(m model.Actors) core.Actor {
	return core.Actor{
		ID:        m.ID,
		Type:      core.EnumActorType(m.Type),
		UpdatedAt: m.UpdatedAt,
		CreatedAt: m.CreatedAt,
	}
}

func MembershipToModel(mb core.Membership) model.Memberships {
	return model.Memberships{
		ID:            mb.ID,
		CorrelationID: mb.CorrID,
		ActorID:       mb.ActorID,
		BusinessID:    mb.BusinessID,
		KeycloakSub:   mb.KeycloakSub,
		PersonID:      mb.PersonID, // nullable reconciliation seam
		Role:          string(mb.Role),
		Status:        string(mb.Status),
		UpdatedAt:     mb.UpdatedAt,
		CreatedAt:     mb.CreatedAt,
	}
}

func MembershipFromModel(m model.Memberships) core.Membership {
	return core.Membership{
		ID:          m.ID,
		CorrID:      m.CorrelationID,
		ActorID:     m.ActorID,
		BusinessID:  m.BusinessID,
		KeycloakSub: m.KeycloakSub,
		PersonID:    m.PersonID,
		Role:        core.EnumMembershipRole(m.Role),
		Status:      core.EnumMembershipStatus(m.Status),
		UpdatedAt:   m.UpdatedAt,
		CreatedAt:   m.CreatedAt,
	}
}

func InvitationToModel(iv core.Invitation) model.Invitations {
	return model.Invitations{
		ID:            iv.ID,
		CorrelationID: iv.CorrID,
		BusinessID:    iv.BusinessID,
		InvitedEmail:  iv.InvitedEmail,
		IntendedRole:  string(iv.IntendedRole),
		Token:         iv.Token,
		Status:        string(iv.Status),
		InvitedBy:     iv.InvitedBy,
		ExpiresAt:     iv.ExpiresAt,
		UpdatedAt:     iv.UpdatedAt,
		CreatedAt:     iv.CreatedAt,
	}
}

func InvitationFromModel(m model.Invitations) core.Invitation {
	return core.Invitation{
		ID:           m.ID,
		CorrID:       m.CorrelationID,
		BusinessID:   m.BusinessID,
		InvitedEmail: m.InvitedEmail,
		IntendedRole: core.EnumMembershipRole(m.IntendedRole),
		Token:        m.Token,
		Status:       core.EnumInvitationStatus(m.Status),
		InvitedBy:    m.InvitedBy,
		ExpiresAt:    m.ExpiresAt,
		UpdatedAt:    m.UpdatedAt,
		CreatedAt:    m.CreatedAt,
	}
}

func ServiceAccountToModel(sa core.ServiceAccount) model.ServiceAccounts {
	return model.ServiceAccounts{
		ID:            sa.ID,
		CorrelationID: sa.CorrID,
		ActorID:       sa.ActorID,
		BusinessID:    sa.BusinessID,
		Label:         sa.Label,
		Status:        string(sa.Status),
		CredentialRef: sa.CredentialRef,
		UpdatedAt:     sa.UpdatedAt,
		CreatedAt:     sa.CreatedAt,
	}
}

func ServiceAccountFromModel(m model.ServiceAccounts) core.ServiceAccount {
	return core.ServiceAccount{
		ID:            m.ID,
		CorrID:        m.CorrelationID,
		ActorID:       m.ActorID,
		BusinessID:    m.BusinessID,
		Label:         m.Label,
		Status:        core.EnumServiceAccountStatus(m.Status),
		CredentialRef: m.CredentialRef,
		UpdatedAt:     m.UpdatedAt,
		CreatedAt:     m.CreatedAt,
	}
}
