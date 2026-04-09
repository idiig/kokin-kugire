package kugire

import (
	"reflect"
	"testing"
)

func TestDiffKugire(t *testing.T) {
	tests := []struct {
		name    string
		before  []KugirePos
		after   []KugirePos
		added   []KugirePos
		removed []KugirePos
	}{
		{
			name:    "no change",
			before:  []KugirePos{{1, "morph"}, {3, "kaneko"}},
			after:   []KugirePos{{1, "morph"}, {3, "kaneko"}},
			added:   nil,
			removed: nil,
		},
		{
			name:    "one added",
			before:  []KugirePos{{1, "morph"}},
			after:   []KugirePos{{1, "morph"}, {3, "kaneko"}},
			added:   []KugirePos{{3, "kaneko"}},
			removed: nil,
		},
		{
			name:    "one removed",
			before:  []KugirePos{{1, "morph"}, {3, "kaneko"}},
			after:   []KugirePos{{1, "morph"}},
			added:   nil,
			removed: []KugirePos{{3, "kaneko"}},
		},
		{
			name:    "added and removed",
			before:  []KugirePos{{1, "morph"}, {2, "kaneko"}},
			after:   []KugirePos{{1, "morph"}, {3, "kaneko"}},
			added:   []KugirePos{{3, "kaneko"}},
			removed: []KugirePos{{2, "kaneko"}},
		},
		{
			name:    "both empty",
			before:  nil,
			after:   nil,
			added:   nil,
			removed: nil,
		},
		{
			name:    "before empty",
			before:  nil,
			after:   []KugirePos{{2, "morph"}},
			added:   []KugirePos{{2, "morph"}},
			removed: nil,
		},
		{
			name:    "after empty",
			before:  []KugirePos{{2, "morph"}},
			after:   nil,
			added:   nil,
			removed: []KugirePos{{2, "morph"}},
		},
		{
			name:   "same seg different source",
			before: []KugirePos{{1, "morph"}},
			after:  []KugirePos{{1, "kaneko"}},
			added:  []KugirePos{{1, "kaneko"}},
			removed: []KugirePos{{1, "morph"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := DiffKugire(tt.before, tt.after)
			if !reflect.DeepEqual(diff.Added, tt.added) {
				t.Errorf("Added:\n  got  %v\n  want %v", diff.Added, tt.added)
			}
			if !reflect.DeepEqual(diff.Removed, tt.removed) {
				t.Errorf("Removed:\n  got  %v\n  want %v", diff.Removed, tt.removed)
			}
		})
	}
}
