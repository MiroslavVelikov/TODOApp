package utils

const (
	UnknownPriority = "Undefined"
	MediumPriority  = "Medium"
)

const (
	Undefined   = "Undefined"
	NotAssigned = "Not Assigned"
	Assigned    = "Assigned"
	InProgress  = "In Progress"
	InReview    = "In Review"
	Completed   = "Completed"
)

func NextStatus(status string) string {
	switch status {
	case NotAssigned:
		return Assigned
	case Assigned:
		return InProgress
	case InReview:
		return Completed
	case InProgress:
		return InReview
	case Completed:
		return Completed
	default:
		return Undefined
	}
}
