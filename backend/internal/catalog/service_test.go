package catalog

import "testing"

func TestCatalogService_BuscaEvolucao_ResultadoEsperado(t *testing.T) {
	svc := MustLoad()

	t.Run("Search_QueryVazia_DeveRetornarOrdenadoPorID", func(t *testing.T) {
		results := svc.Search("", 3)
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if results[0].ID != 1 || results[1].ID != 2 || results[2].ID != 3 {
			t.Fatalf("unexpected ids: %v", []int{results[0].ID, results[1].ID, results[2].ID})
		}
	})

	t.Run("EvolutionOptions_BaseBulbasaur_DevePermitirEvolve", func(t *testing.T) {
		opts, err := svc.EvolutionOptions(1, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(opts) == 0 {
			t.Fatal("expected non-empty evolution options")
		}
		found := false
		for _, o := range opts {
			if o.ID == 2 && o.Action == "Evolve" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("expected option id=2 action=Evolve")
		}
	})
}
