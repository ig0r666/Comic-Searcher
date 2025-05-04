package core

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusIdle    ServiceStatus = "idle"
)

type DBStats struct {
	WordsTotal    int `db:"words_total"`
	WordsUnique   int `db:"words_unique"`
	ComicsFetched int `db:"comics_fetched"`
}

type ServiceStats struct {
	DBStats
	ComicsTotal int
}

type Comics struct {
	ID    int
	URL   string
	Words []string
}

type JsonXKCDInfo struct {
	ID         int    `json:"num"`
	URL        string `json:"img"`
	Title      string `json:"title"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	SafeTitle  string `json:"safe_title"`
}

type XKCDInfo struct {
	ID         int
	URL        string
	Title      string
	Alt        string
	Transcript string
	SafeTitle  string
}
