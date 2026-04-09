package kugire

// KugireDiff holds the result of comparing two kugire position lists.
type KugireDiff struct {
	Added   []KugirePos // positions present in after but not in before
	Removed []KugirePos // positions present in before but not in after
}

// DiffKugire returns the positions added and removed between before and after.
// Two positions are considered equal when both AfterSeg and Source match.
func DiffKugire(before, after []KugirePos) KugireDiff {
	inBefore := posSet(before)
	inAfter := posSet(after)

	var diff KugireDiff
	for _, p := range after {
		if !inBefore[p] {
			diff.Added = append(diff.Added, p)
		}
	}
	for _, p := range before {
		if !inAfter[p] {
			diff.Removed = append(diff.Removed, p)
		}
	}
	return diff
}

func posSet(positions []KugirePos) map[KugirePos]bool {
	s := make(map[KugirePos]bool, len(positions))
	for _, p := range positions {
		s[p] = true
	}
	return s
}
