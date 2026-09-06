package enums

type SelectionReason string

const (
	SelectionReasonWinner    SelectionReason = "winner"
	SelectionReasonMinimum   SelectionReason = "minimum"
	SelectionReasonDisplaced SelectionReason = "displaced"
	SelectionReasonBudget    SelectionReason = "budget"
)

type SelectionCycleStatus string

const (
	SelectionCycleRunning   SelectionCycleStatus = "running"
	SelectionCycleCompleted SelectionCycleStatus = "completed"
	SelectionCycleFailed    SelectionCycleStatus = "failed"
)
