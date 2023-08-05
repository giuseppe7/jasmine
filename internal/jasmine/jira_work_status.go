package jasmine

type JiraWorkStatus struct {
	name        string
	count       int
	elapsedTime float64
	successful  bool
	err         error
}
