package elastic

type esErrorResponse struct {
	Error esError `json:"error"`
}

type esError struct {
	RootCause []esError `json:"root_cause"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
}

func (e esError) Error() string {
	return e.Reason
}
