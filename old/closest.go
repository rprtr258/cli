package cli

// dist returns the Levenshtein distance between two strings.
// that is 0 for s == t, 1 for s and t differing by one letter, etc.
func dist(s, t string) int {
	if s == "" {
		return len(t)
	}

	if t == "" {
		return len(s)
	}

	// TODO: memory layout can be optimized: store contaguous array instead of list of lists
	// TODO: memory consumption can be optimized: store just last row
	dists := make([][]int, len(s)+1)
	for i := range dists {
		dists[i] = make([]int, len(t)+1)
		dists[i][0] = i
	}

	for j := range t {
		dists[0][j] = j
	}

	for i, sc := range s {
		for j, tc := range t {
			if sc == tc {
				dists[i+1][j+1] = dists[i][j]
			} else {
				dists[i+1][j+1] = dists[i][j] + 1
				if dists[i+1][j] < dists[i+1][j+1] {
					dists[i+1][j+1] = dists[i+1][j] + 1
				}
				if dists[i][j+1] < dists[i+1][j+1] {
					dists[i+1][j+1] = dists[i][j+1] + 1
				}
			}
		}
	}

	return dists[len(s)][len(t)]
}

func closestChoice(cmd string, choices []string) (string, int) {
	if len(choices) == 0 {
		return "", 0
	}

	minChoice, minDist := "", -1
	for _, c := range choices {
		if l := dist(cmd, c); minDist == -1 || l < minDist {
			minChoice, minDist = c, l
		}
	}
	return minChoice, minDist
}
