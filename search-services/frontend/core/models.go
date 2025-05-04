package core

type SearchResponse struct {
	Comics []Comic `json:"comics"`
}

type Comic struct {
	ID       int    `json:"id"`
	ImageURL string `json:"url"`
}

type Stats struct {
	ComicsTotal   int `json:"comics_total"`
	ComicsFetched int `json:"comics_fetched"`
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
}

type Status struct {
	Status string `json:"status"`
}
