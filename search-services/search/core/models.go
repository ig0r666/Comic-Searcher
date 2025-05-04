package core

type DbComics struct {
	ID       int    `db:"comic_id"`
	URL      string `db:"image_url"`
	Keywords string `db:"keywords"`
}

type Comics struct {
	ID       int
	URL      string
	Keywords string
}
