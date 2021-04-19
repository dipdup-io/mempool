package receiver

// Message -
type Message struct {
	Status Status
	Body   interface{}
}

// Status
type Status string

// Statuses
const (
	StatusApplied       Status = "applied"
	StatusBranchDelayed Status = "branch_delayed"
	StatusBranchRefused Status = "branch_refused"
	StatusRefused       Status = "refused"
	StatusUnprocessed   Status = "unprocessed"
)
