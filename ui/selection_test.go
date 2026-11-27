package nui_test

import (
	"math"
	"slices"
	"testing"

	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSelectionStateOrderAndRevisions(t *testing.T) {
	type key int
	var s nui.SelectionState[key]
	if s.Len() != 0 || s.Revision() != 0 || s.Has(0) || s.Clear() || s.Set(nil) {
		t.Fatal("zero selection is not an unchanged empty set")
	}
	if !s.Set([]key{2, 0, 2, 1, 0}) || !slices.Equal(s.Values(), []key{2, 0, 1}) || s.Revision() != 1 {
		t.Fatal("ordered deduplication lost a typed zero or its first position")
	}
	if s.Add(0) || s.Remove(4) || s.Set([]key{2, 0, 1, 2}) || s.Revision() != 1 {
		t.Fatal("no-op mutation advanced the revision")
	}
	if !s.Remove(0) || !s.Add(0) || !slices.Equal(s.Values(), []key{2, 1, 0}) || s.Revision() != 3 {
		t.Fatal("remove/add did not preserve ordered membership")
	}
	if !s.Set([]key{0, 1, 2}) || s.Revision() != 4 {
		t.Fatal("ordered replacement was not a change")
	}
	if !s.Clear() || s.Clear() || s.Revision() != 5 || s.Len() != 0 {
		t.Fatal("clear revision or contents")
	}
	var strings nui.SelectionState[string]
	if !strings.Add("") || !strings.Has("") {
		t.Fatal("selection state reserved a valid typed zero")
	}
}

func TestSelectionStateOwnsSnapshots(t *testing.T) {
	values := []string{"a", "b"}
	var s nui.SelectionState[string]
	s.Set(values)
	values[0] = "source changed"
	snapshot := s.Values()
	snapshot[1] = "snapshot changed"
	if !slices.Equal(s.Values(), []string{"a", "b"}) || s.Has("source changed") || s.Has("snapshot changed") {
		t.Fatal("a writable alias escaped selection ownership")
	}
	s.Remove("a")
	if snapshot[0] != "a" {
		t.Fatal("a later mutation rewrote an earlier snapshot")
	}
}

func TestSelectionStateRejectsNonreflexiveKeysAtomically(t *testing.T) {
	var s nui.SelectionState[float64]
	s.Set([]float64{1, 2})
	for _, mutate := range []func(){
		func() { s.Set([]float64{3, math.NaN(), 4}) },
		func() { s.Add(math.NaN()) },
		func() { s.Remove(math.NaN()) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("non-reflexive key was accepted")
				}
			}()
			mutate()
		}()
		if !slices.Equal(s.Values(), []float64{1, 2}) || s.Revision() != 1 {
			t.Fatal("invalid mutation damaged the previous snapshot")
		}
	}
}
