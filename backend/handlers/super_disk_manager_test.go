package handlers

import "testing"

func TestSuperDiskManagerValidCategories(t *testing.T) {
	if !superDiskManagerValidCategories([]string{"images", "anonymous_volumes"}) {
		t.Fatal("expected allowlisted categories to be valid")
	}
	if superDiskManagerValidCategories([]string{"images", "images"}) {
		t.Fatal("duplicate category must be rejected")
	}
	if superDiskManagerValidCategories([]string{"images; rm -rf /"}) {
		t.Fatal("arbitrary command-like input must be rejected")
	}
}

func TestSuperDiskManagerSortedCategories(t *testing.T) {
	got := superDiskManagerSortedCategories([]string{" Images ", "containers"})
	if len(got) != 2 || got[0] != "containers" || got[1] != "images" {
		t.Fatalf("unexpected sorted categories: %#v", got)
	}
}
