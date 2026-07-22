package stackstatus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stacklogs"
	"golang.org/x/sync/errgroup"
)

const (
	defaultProbeConcurrency = 8
	probeDetailMaxLen       = 80
	deployableProbeTimeout  = 5 * time.Second
)

// SourceAPI means a live vendor status endpoint was called (non-deployable providers).
const SourceAPI = "api"

// SourceProbe means a live deployable host health check was called.
const SourceProbe = "probe"

// SourceAppEnv means the node is satisfied by project .env (e.g. INNGEST_DEV, DATABASE_URL).
const SourceAppEnv = "app_env"

// errDeployableProbeUnimplemented is returned when a deployable provider has no probe yet.
var errDeployableProbeUnimplemented = errors.New("deployable probe not implemented")

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

// probeProvider runs a 5s live health check for deployable hosts (vercel, render, supabase).
func probeProvider(ctx context.Context, spec *provider.Spec, n config.Node) (healthy bool, detail string, err error) {
	if spec == nil {
		return false, "", errDeployableProbeUnimplemented
	}
	switch spec.Name {
	case "vercel", "render", "supabase":
	default:
		return false, "", errDeployableProbeUnimplemented
	}

	ctx, cancel := context.WithTimeout(ctx, deployableProbeTimeout)
	defer cancel()

	token, ok := deployableProbeToken(spec)
	if !ok {
		return false, "missing credential (run perch logs setup)", nil
	}

	client := provider.HTTPClientForAPI()
	switch spec.Name {
	case "vercel":
		return probeVercelDeployable(ctx, client, spec, n, token)
	case "render":
		return probeRenderDeployable(ctx, client, spec, n, token)
	case "supabase":
		return probeSupabaseDeployable(ctx, client, spec, n, token)
	default:
		return false, "", errDeployableProbeUnimplemented
	}
}

// deployableTokenFn is swapped in tests so machine CLI auth does not leak into unit tests.
var deployableTokenFn = defaultDeployableProbeToken

func deployableProbeToken(spec *provider.Spec) (string, bool) {
	return deployableTokenFn(spec)
}

func defaultDeployableProbeToken(spec *provider.Spec) (string, bool) {
	if tok, ok := stacklogs.AuthFileToken(spec); ok {
		return tok, true
	}
	key := strings.TrimSpace(spec.Credentials.Key)
	if key != "" {
		store := credentials.NewStore()
		if v, ok, err := store.Get(key); err == nil && ok {
			if tok := strings.TrimSpace(v); tok != "" {
				return tok, true
			}
		}
	}
	if ev := strings.TrimSpace(spec.Credentials.EnvVar); ev != "" {
		if tok := strings.TrimSpace(os.Getenv(ev)); tok != "" {
			return tok, true
		}
	}
	return "", false
}

// SetDeployableTokenFnForTest replaces deployable token discovery. Returns the previous fn.
func SetDeployableTokenFnForTest(fn func(*provider.Spec) (string, bool)) func(*provider.Spec) (string, bool) {
	prev := deployableTokenFn
	if fn == nil {
		deployableTokenFn = defaultDeployableProbeToken
	} else {
		deployableTokenFn = fn
	}
	return prev
}

func probeVercelDeployable(ctx context.Context, client *http.Client, spec *provider.Spec, n config.Node, token string) (bool, string, error) {
	project := strings.TrimSpace(n.Project)
	if project == "" {
		return false, "node needs a project field for Vercel probe", nil
	}
	var body struct {
		Name string `json:"name"`
	}
	err := provider.DoGETJSON(ctx, client, spec, "status", map[string]string{
		"token": token, "project": project, "service": "",
	}, &body)
	if detail, handled := deployableProbeError(err); handled {
		return false, detail, nil
	}
	if err != nil {
		return false, truncateProbeDetail(err.Error()), nil
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = project
	}
	return true, fmt.Sprintf("project %s active", name), nil
}

func probeRenderDeployable(ctx context.Context, client *http.Client, spec *provider.Spec, n config.Node, token string) (bool, string, error) {
	service := strings.TrimSpace(n.Service)
	if service == "" {
		return false, "node needs a service field for Render probe", nil
	}
	var body struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Suspended string `json:"suspended"`
	}
	err := provider.DoGETJSON(ctx, client, spec, "status", map[string]string{
		"token": token, "project": "", "service": service,
	}, &body)
	if detail, handled := deployableProbeError(err); handled {
		return false, detail, nil
	}
	if err != nil {
		return false, truncateProbeDetail(err.Error()), nil
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = service
	}
	status := strings.ToLower(strings.TrimSpace(body.Status))
	suspended := strings.ToLower(strings.TrimSpace(body.Suspended))
	// Render API uses suspended=not_suspended; some status fields use "live".
	live := status == "live" || suspended == "not_suspended" || (status == "" && suspended == "")
	if !live {
		detail := fmt.Sprintf("service %s not live", name)
		if suspended != "" {
			detail = fmt.Sprintf("service %s suspended=%s", name, body.Suspended)
		} else if status != "" {
			detail = fmt.Sprintf("service %s status=%s", name, body.Status)
		}
		return false, detail, nil
	}
	return true, fmt.Sprintf("service %s live", name), nil
}

func probeSupabaseDeployable(ctx context.Context, client *http.Client, spec *provider.Spec, n config.Node, token string) (bool, string, error) {
	ref := strings.TrimSpace(n.Project)
	if ref == "" {
		return false, "node needs a project field for Supabase probe", nil
	}
	var body struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	err := provider.DoGETJSON(ctx, client, spec, "status", map[string]string{
		"token": token, "project": ref, "service": "",
	}, &body)
	if detail, handled := deployableProbeError(err); handled {
		return false, detail, nil
	}
	if err != nil {
		return false, truncateProbeDetail(err.Error()), nil
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = ref
	}
	if strings.TrimSpace(body.Status) != "ACTIVE_HEALTHY" {
		return false, fmt.Sprintf("project %s status=%s", name, strings.TrimSpace(body.Status)), nil
	}
	return true, fmt.Sprintf("project %s healthy", name), nil
}

func deployableProbeError(err error) (detail string, handled bool) {
	if err == nil {
		return "", false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "probe timed out", true
	}
	msg := err.Error()
	if strings.Contains(msg, "401") || strings.Contains(msg, "403") ||
		strings.Contains(msg, "Unauthorized") || strings.Contains(msg, "Forbidden") {
		return "credential invalid or expired", true
	}
	if strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "Client.Timeout") {
		return "probe timed out", true
	}
	return "", false
}
