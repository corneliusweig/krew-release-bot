package events

//Publisher defines the publisher interface
type Publisher interface {
	Publish([]Event) error
}

//InMemoryPublisher is the in memory publisher
type InMemoryPublisher struct{}

//Publish events using InMemoryPublisher
func (im *InMemoryPublisher) Publish(events []Event) error {
	return nil
}

//GithubRepoPublisher is the github publisher
type GithubRepoPublisher struct{}

//Publish publish events using GithubRepoPublisher
func (gh *GithubRepoPublisher) Publish(events []Event) error {
	return nil
}
