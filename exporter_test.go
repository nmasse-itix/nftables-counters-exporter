package exporter

import (
	"testing"
)

func TestCleanupName(t *testing.T) {
	input := []string{"ABC-DEF", "Awesome-Counter-1", "NomÉtrangeAvecDesCaract#r{$Bizares"}
	expected := []string{"abc_def", "awesome_counter_1", "nom_trangeavecdescaract_r__bizares"}

	for i := range input {
		output := cleanupName(input[i])
		if output != expected[i] {
			t.Errorf("cleanupName(%s): got %s, expected %s", input[i], output, expected[i])
		}
	}
}
