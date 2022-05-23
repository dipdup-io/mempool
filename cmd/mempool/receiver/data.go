package receiver

// Message -
type Message struct {
	Status   Status
	Protocol string
	Body     interface{}
}

// Status
type Status string

// Statuses
const (
	StatusApplied       Status = "applied"
	StatusBranchDelayed Status = "branch_delayed"
	StatusBranchRefused Status = "branch_refused"
	StatusRefused       Status = "refused"
	StatusOutdated      Status = "outdated"
	StatusUnprocessed   Status = "unprocessed"
)
