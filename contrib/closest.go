package flags

// TODO: OPTIMIZE: less memory could be used
func dist(s, t string) int {
	if s == "" {
		return len(t)
	}

	if t == "" {
		return len(s)
	}

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
				dists[i+1][j+1] = min(
					dists[i][j],
					dists[i+1][j],
					dists[i][j+1],
				) + 1
			}
		}
	}

	return dists[len(s)][len(t)]
}

func getClosest(cmd string, choices []string) (closest_ string, distMin_ int) {
	if len(choices) == 0 {
		return "", 0
	}

	closest, distMin := "", -1
	for _, c := range choices {
		if distCur := dist(cmd, c); distMin == -1 || distCur < distMin {
			closest, distMin = c, distCur
		}
	}
	return closest, distMin
}
