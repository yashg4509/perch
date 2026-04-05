package detect

// EdgePair is a directed edge between logical node names in perch.yaml.
type EdgePair struct {
	From string
	To   string
}

// Inference collects inferred edges and cases that need a user prompt.
type Inference struct {
	Edges       []EdgePair
	NeedsPrompt []string
}

// InferEdges applies deterministic rules (spec: supabase + dep). npmPackages are dependency keys from package.json.
func InferEdges(nodes map[string]string, npmPackages []string) Inference {
	var inf Inference
	if hasNPM(npmPackages, "@supabase/supabase-js") {
		if !hasProvider(nodes, "supabase") {
			inf.NeedsPrompt = append(inf.NeedsPrompt, "package.json lists @supabase/supabase-js but no supabase provider was detected")
			return inf
		}
		backend := firstBackendNode(nodes)
		if backend == "" {
			inf.NeedsPrompt = append(inf.NeedsPrompt, "@supabase/supabase-js present but no backend host (render/railway/fly) detected for automatic edge")
			return inf
		}
		sup := firstNodeWithProvider(nodes, "supabase")
		inf.Edges = append(inf.Edges, EdgePair{From: backend, To: sup})
	}
	return inf
}

func hasNPM(pkgs []string, want string) bool {
	for _, p := range pkgs {
		if p == want {
			return true
		}
	}
	return false
}

func hasProvider(nodes map[string]string, provider string) bool {
	for _, p := range nodes {
		if p == provider {
			return true
		}
	}
	return false
}

func firstBackendNode(nodes map[string]string) string {
	for name, p := range nodes {
		switch p {
		case "render", "railway", "fly":
			return name
		}
	}
	return ""
}

func firstNodeWithProvider(nodes map[string]string, provider string) string {
	for name, p := range nodes {
		if p == provider {
			return name
		}
	}
	return ""
}
