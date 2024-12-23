package sync

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type AttestationReport struct{}

func NewAttestationReport() *AttestationReport {
	return &AttestationReport{}
}

func (s *Service) ReportAttestationValidationOutcome(r pubsub.ValidationResult, err error) {
}
