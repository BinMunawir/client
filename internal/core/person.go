package core

// Person and BusinessRole are the DEFERRED compliance-identity spine (design §5). They
// are intentionally NOT built in the starter scope. This file exists so the seam in the
// access layer has a defined other end, and so adding KYB later is *additive* rather
// than a rewrite.
//
// The one thing the starter MUST preserve is the reconciliation seam: Membership.PersonID
// is a nullable link from the operator edge to a natural person. Build the nullable column
// now (see the memberships migration and core.Membership.PersonID); populate it only once
// Person exists. Skipping the seam is the expensive mistake — skipping the entities is fine.
//
// When built (design §5.1, §5.2), the entities would be:
//
//	Person — the natural person, holding a human's identity ONCE, decoupled from any single
//	  Business, so the same human can control one Business and operate another without
//	  duplication. Would hold: national_id (the identity anchor, Nafath in the full system),
//	  full_name, nationality, date_of_birth — plus deferred KYC machinery (is_pep,
//	  screening_status, verification provenance). It is a hub: one human → one screening
//	  record → every relationship linked by construction. It is also the clean answer under
//	  PDPL — erasure/rectification touches one Person, not scattered copies.
//
//	BusinessRole — the typed control edge from a Person to a Business recording *how they
//	  control it*. Would hold: person_id, business_id; role ∈ {ubo, signatory, director,
//	  legal_representative}; role-specific attributes (e.g. ownership_pct + control_basis for
//	  a UBO; signing authority for a signatory); verification state. One Person may hold
//	  several BusinessRoles to the same Business (a founder is ubo + signatory + director —
//	  three edges), each re-verified on its own trigger. Ownership % stays on this edge;
//	  screening stays on the Person.
//
// The four canonical human scenarios the model must pass (design §6.2) all resolve on the
// access side today; PersonID reserves the link for scenarios 1 and 4:
//
//	1. Founder            — control: BusinessRole[ubo, signatory]; operate: Membership[owner];  PersonID set
//	2. Silent UBO         — control: BusinessRole[ubo];            operate: (none);              PersonID —
//	3. Outsourced clerk   — control: (none);                       operate: Membership[finance]; PersonID null
//	4. Approving signatory— control: BusinessRole[signatory];      operate: Membership[checker]; PersonID set
