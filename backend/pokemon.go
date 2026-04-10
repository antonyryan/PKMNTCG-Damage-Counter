package main

// Pokemon is the minimal response shape used by the autocomplete API.
type Pokemon struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
