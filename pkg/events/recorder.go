package events

//Recorder records the events
type Recorder struct {
	Owner      string
	Repo       string
	PluginName string
	Version    string

	events []Event
}

//NewRecorder returns the new recorder
func NewRecorder() *Recorder {
	return &Recorder{
		events: []Event{},
	}
}

//Event is the event
type Event struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	RawPayload  string `json:"raw-payload"`
	Error       string `json:"error"`
}

//Record records the event
func (r *Recorder) Record(event Event) error {
	r.events = append(r.events, event)
	return nil
}
