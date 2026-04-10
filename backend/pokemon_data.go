package main

// pokemonMock is the temporary in-memory catalog used by the search endpoint.
// It is intentionally small and easy to replace with an official dataset later.
var pokemonMock = []Pokemon{
	{ID: 1, Name: "Pikachu"},
	{ID: 2, Name: "Charizard"},
	{ID: 3, Name: "Mewtwo"},
	{ID: 4, Name: "Arceus"},
	{ID: 5, Name: "Gengar"},
	{ID: 6, Name: "Lucario"},
	{ID: 7, Name: "Greninja"},
	{ID: 8, Name: "Eevee"},
	{ID: 9, Name: "Snorlax"},
	{ID: 10, Name: "Gardevoir"},
	{ID: 11, Name: "Rayquaza"},
	{ID: 12, Name: "Dragonite"},
	{ID: 13, Name: "Blastoise"},
	{ID: 14, Name: "Venusaur"},
	{ID: 15, Name: "Garchomp"},
}
