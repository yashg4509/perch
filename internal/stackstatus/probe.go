package stackstatus

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"golang.org/x/sync/errgroup"
)

const (
	defaultProbeConcurrency = 8
	probeDetailMaxLen       = 80
)

// SourceAPI means a live vendor status endpoint was called.
const SourceAPI = "api"

// SourceAppEnv means the node is satisfied by project .env (e.g. INNGEST_DEV, DATABASE_URL).
const SourceAppEnv = "app_env"

type probeJob struct {
	index      int
	name       string
	spec       *provider.Spec
	node       config.Node
	token      string
	projectEnv map[string]string
}

// probeJobsParallel runs vendor status checks concurrently and writes results into rows.
func probeJobsParallel(ctx context.Context, jobs []probeJob, rows []NodeReport, concurrency int) {
	if len(jobs) == 0 {
		return
	}
	if concurrency <= 0 {
		concurrency = defaultProbeConcurrency
	}
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	client := provider.HTTPClientForAPI()

	for _, job := range jobs {
		job := job
		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-gctx.Done():
				return gctx.Err()
			}
			defer func() { <-sem }()

			healthy, detail := runProbe(gctx, client, job)
			mu.Lock()
			rows[job.index].Healthy = healthy
			rows[job.index].StatusSource = SourceAPI
			rows[job.index].Detail = detail
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
}

func runProbe(ctx context.Context, client *http.Client, job probeJob) (bool, string) {
	switch job.spec.StatusProbeMode() {
	case "signed":
		return probeSigned(ctx, client, job)
	default:
		return probeREST(ctx, client, job.spec, job.node, job.token, job.projectEnv)
	}
}

func probeREST(ctx context.Context, client *http.Client, spec *provider.Spec, n config.Node, token string, env map[string]string) (bool, string) {
	if spec == nil || strings.TrimSpace(token) == "" {
		return false, "missing credential"
	}
	if env == nil {
		env = map[string]string{}
	}
	proj := config.EffectiveProject(n, spec, env)
	if proj == "" {
		proj = strings.TrimSpace(n.Project)
	}
	vars := map[string]string{
		"token":   token,
		"project": proj,
		"service": config.EffectiveService(n, spec, env),
	}
	err := provider.DoGETJSON(ctx, client, spec, "status", vars, nil)
	if err != nil {
		return false, truncateProbeDetail(err.Error())
	}
	return true, ""
}

func probeSigned(ctx context.Context, client *http.Client, job probeJob) (bool, string) {
	if job.spec == nil || job.spec.Name != "pusher" {
		return false, "signed probe not configured"
	}
	env := job.projectEnv
	if env == nil {
		env = map[string]string{}
	}
	appKey := strings.TrimSpace(env["PUSHER_KEY"])
	appSecret := strings.TrimSpace(job.token)
	if appSecret == "" {
		appSecret = strings.TrimSpace(env["PUSHER_SECRET"])
	}
	appID := config.EffectiveProject(job.node, job.spec, env)
	if appID == "" {
		appID = strings.TrimSpace(job.node.Project)
	}
	if appKey == "" || appSecret == "" || appID == "" {
		return false, "need PUSHER_KEY, PUSHER_SECRET, and PUSHER_APP_ID in .env"
	}
	path := provider.PusherChannelsPath(appID)
	baseURL := provider.PusherAPIBaseURL(strings.TrimSpace(env["PUSHER_CLUSTER"]))
	code, err := provider.PusherSignedGET(ctx, client, baseURL, path, appKey, appSecret)
	return provider.PusherProbeOK(code, err)
}

func truncateProbeDetail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= probeDetailMaxLen {
		return s
	}
	return s[:probeDetailMaxLen-1] + "…"
}

// configuredViaAppEnv is true when the app runs off .env without a Perch credential.
func configuredViaAppEnv(spec *provider.Spec, env map[string]string, hasCred func(string) (bool, error)) bool {
	if spec == nil || env == nil {
		return false
	}
	switch spec.Name {
	case "inngest":
		if strings.TrimSpace(env["INNGEST_DEV"]) != "1" {
			return false
		}
	case "neon":
		db := strings.TrimSpace(env["DATABASE_URL"])
		if db == "" || !strings.Contains(db, "neon") {
			return false
		}
	default:
		return false
	}
	if spec.Credentials.Key == "" || hasCred == nil {
		return true
	}
	ok, err := hasCred(spec.Credentials.Key)
	if err != nil {
		return false
	}
	return !ok
}

func appEnvDetail(spec *provider.Spec) string {
	if spec == nil {
		return "configured from project .env"
	}
	switch spec.Name {
	case "inngest":
		return "INNGEST_DEV=1 (local dev server)"
	case "neon":
		return "DATABASE_URL points at Neon"
	default:
		return "configured from project .env"
	}
}

func shouldProbeAPI(spec *provider.Spec, opts CollectOptions, _ func(string) (bool, error)) (token string, ok bool) {
	if spec == nil || spec.Deployable || !opts.ProbeAPI {
		return "", false
	}
	if spec.StatusProbeMode() == "none" {
		return "", false
	}
	if spec.Credentials.Key == "" {
		return "", false
	}
	if opts.CredStore == nil {
		return "", false
	}
	v, has, err := opts.CredStore.Get(spec.Credentials.Key)
	if err != nil || !has || strings.TrimSpace(v) == "" {
		return "", false
	}
	if _, hasEndpoint := spec.API.Endpoints["status"]; !hasEndpoint {
		return "", false
	}
	return strings.TrimSpace(v), true
}
