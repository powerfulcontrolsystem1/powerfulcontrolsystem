package modules

import "testing"

func TestCatalogHasStableCoreBoundaries(t *testing.T) {
	t.Parallel()
	if err := Validate(Catalog()); err != nil {
		t.Fatalf("catalog: %v", err)
	}
	names := Names()
	if len(names) < 10 || names[0] == "" {
		t.Fatalf("unexpected catalog names: %#v", names)
	}
}

func TestCatalogRejectsUnknownDependency(t *testing.T) {
	t.Parallel()
	err := Validate([]Descriptor{{Name: "a", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"missing"}}})
	if err == nil {
		t.Fatal("expected unknown dependency error")
	}
}
