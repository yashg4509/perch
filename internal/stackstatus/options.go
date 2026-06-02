package stackstatus

import "github.com/yashg4509/perch/internal/credentials"

// CollectOptions configures [Collect].
type CollectOptions struct {
	CredStore        *credentials.Store
	ProjectEnv       map[string]string // optional .env beside perch.yaml (non-secret keys only, e.g. INNGEST_DEV)
	ProbeAPI         bool              // call provider status endpoints in parallel (default true)
	ProbeConcurrency int               // max concurrent probes; 0 = 8
}
