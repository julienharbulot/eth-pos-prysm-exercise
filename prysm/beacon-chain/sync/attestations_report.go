package sync

import (
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/time/slots"
	"github.com/sirupsen/logrus"
)

type AttestationReport struct {
	lock      sync.RWMutex
	lastEpoch primitives.Epoch
	nOk       uint64
	nErr      uint64
	errs      map[string]uint64
}

func NewAttestationReport() *AttestationReport {
	return &AttestationReport{}
}

func (s *Service) ReportAttestationValidationOutcome(r pubsub.ValidationResult, err error) {
	currentEpoch := slots.ToEpoch(s.cfg.clock.CurrentSlot())

	s.cfg.attestationReport.lock.Lock()
	defer s.cfg.attestationReport.lock.Unlock()

	report := s.cfg.attestationReport

	if r == pubsub.ValidationAccept {
		report.nOk += 1
	} else {
		report.nErr += 1
		report.errs[err.Error()] += 1
	}

	if currentEpoch > report.lastEpoch {
		report.lastEpoch = currentEpoch
		go report.LogAttestationReport()
	}
}

func (r *AttestationReport) LogAttestationReport() {
	r.lock.RLock()
	defer r.lock.RUnlock()

	log.WithFields(logrus.Fields{
		"epoch": r.lastEpoch,
		"nOk":   r.nOk,
		"nErr":  r.nErr,
		"errs":  r.errs,
	}).Debug("Attestation Validation Report")
}
