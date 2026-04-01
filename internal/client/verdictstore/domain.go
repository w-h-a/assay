package verdictstore

// Result represents the outcome of a verification run.
type Result string

const (
	Pass  Result = "pass"
	Fail  Result = "fail"
	Error Result = "error"
)

// Verdict is the structured outcome of verifying a spec against an implementation.
type Verdict struct {
	FormatVersion  string   `json:"format_version"`
	SpecHash       string   `json:"spec_hash"`
	BindingHash    string   `json:"binding_hash"`
	Backend        string   `json:"backend"`
	AssayVersion   string   `json:"assay_version"`
	BackendVersion string   `json:"backend_version"`
	Result         Result   `json:"result"`
	Evidence       Evidence `json:"evidence"`
	Timestamp      string   `json:"timestamp"`
}

// Evidence captures the details of a verification run.
type Evidence struct {
	Checks         int    `json:"checks"`
	Counterexample string `json:"counterexample,omitempty"`
	DurationMs     int64  `json:"duration_ms"`
}
