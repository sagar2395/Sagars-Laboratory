# Task 078: Certification & training framework

## Phase
M9 — Commercial & Hosted (deferred)

## Type
feature

## Priority
P2

## Description
Build on the existing learning-paths + challenge + results engines (M3) to support
graded certification tracks ("pilot ratings") and structured training content,
deliverable as premium packs. Engine-side: stable interfaces for graded,
proctored, credentialed content; content + exam delivery live in the private repo.

## Implementation Notes
- Reuse pkg/results + challenge grading; add a credential/attestation schema.
- Certification content is premium (private repo); the OSS engine only provides
  the grading + attestation interfaces.
- Exam delivery/credentialing platform is a maintainer manual action (D5).

## Acceptance Criteria
- [ ] Credential/attestation schema + interfaces in the OSS engine
- [ ] A premium certification track runs on the unmodified OSS engine
- [ ] Results/credentials are verifiable

## Dependencies
050, 051, 052, 070, 074
