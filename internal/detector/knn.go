package detector

import "math"

const (
	k         = 5
	threshold = 0.6
)

type neighbor struct {
	dist    float64
	isFraud bool
}

func knn(query []float64, vectors [][]float64, labels []bool) int {
	best := make([]neighbor, 0, k)

	for i, vec := range vectors {
		d := euclidean(query, vec)

		if len(best) < k {
			best = append(best, neighbor{d, labels[i]})
			continue
		}

		worstIdx := worstIndex(best)
		if d < best[worstIdx].dist {
			best[worstIdx] = neighbor{d, labels[i]}
		}
	}

	frauds := 0
	for _, n := range best {
		if n.isFraud {
			frauds++
		}
	}
	return frauds
}

func euclidean(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func worstIndex(ns []neighbor) int {
	idx := 0
	for i := 1; i < len(ns); i++ {
		if ns[i].dist > ns[idx].dist {
			idx = i
		}
	}
	return idx
}
