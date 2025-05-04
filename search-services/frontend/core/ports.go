package core

type API interface {
	Search(string) (SearchResponse, error)
	Update(string) error
	Drop(string) error
	GetStatus() (Status, error)
	GetStats() (Stats, error)
	Login(string, string) (string, error)
}
