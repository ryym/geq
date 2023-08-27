package geq

func selectionIndex(from Selection, sels []Selection, target Selection) int {
	f := -1
	t := -1
	for i, sel := range sels {
		if sel == from {
			f = i
		}
		if sel == target {
			t = i
		}
	}
	if f < 0 || t < f {
		return -1
	}
	return t - f
}
