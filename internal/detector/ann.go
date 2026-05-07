package detector

import (
	"math"
	"sync"
)

// IVFIndex é um índice Inverted File para busca aproximada de vizinhos mais próximos.
//
// Tanto centroides quanto vetores do dataset são quantizados em int8 (escala 127).
// Distâncias são computadas em int32 (quadrado euclideano, sem sqrt), o que é
// suficiente para comparações de ranking: max sum = 14 × 254² ≈ 903k << 2³¹.
type IVFIndex struct {
	centroids []int8    // flat [nClusters × dim], quantizado em [-127, 127]
	vectors   []int8    // flat [nVectors × dim], quantizado em [-127, 127]
	labels    []bool
	invLists  [][]int32
	dim       int
	nClusters int
	nprobe    int
	k         int
	threshold float64

	candidatePool sync.Pool
}

type centDist struct {
	i    int
	dist int32
}

func newIVFIndex(centroids []int8, vectors []int8, labels []bool, invLists [][]int32,
	dim, nClusters, nprobe, k int, threshold float64) *IVFIndex {

	maxCandidates := 0
	for i := 0; i < nClusters && i < nprobe*4; i++ {
		maxCandidates += len(invLists[i])
	}
	avgPerCluster := maxCandidates / max(nprobe*4, 1)
	cap := avgPerCluster*nprobe*2 + 256

	idx := &IVFIndex{
		centroids: centroids,
		vectors:   vectors,
		labels:    labels,
		invLists:  invLists,
		dim:       dim,
		nClusters: nClusters,
		nprobe:    nprobe,
		k:         k,
		threshold: threshold,
	}
	idx.candidatePool = sync.Pool{
		New: func() interface{} {
			s := make([]int32, 0, cap)
			return &s
		},
	}
	return idx
}

func (idx *IVFIndex) Search(query []float64) (approved bool, score float64) {
	dim := idx.dim

	// Pré-quantiza query uma única vez para int8 — evita N×dim conversões float64→int8
	// dentro dos loops de distância.
	var qBuf [vecSize]int8
	qQuery := qBuf[:dim]
	for i, v := range query {
		qQuery[i] = quantizeFloat64(v)
	}

	// Seleção linear dos nprobe centroides mais próximos.
	// Usa distância ao quadrado (int32) — sqrt é monotônico e desnecessário para ranking.
	nprobe := idx.nprobe
	nearest := make([]centDist, 0, nprobe)
	worstIdx := 0
	var worstDist int32

	for i := 0; i < idx.nClusters; i++ {
		d := squaredDistInt(qQuery, idx.centroids[i*dim:(i+1)*dim])
		if len(nearest) < nprobe {
			nearest = append(nearest, centDist{i, d})
			if d > worstDist {
				worstDist = d
				worstIdx = len(nearest) - 1
			}
		} else if d < worstDist {
			nearest[worstIdx] = centDist{i, d}
			worstDist = nearest[0].dist
			worstIdx = 0
			for j := 1; j < nprobe; j++ {
				if nearest[j].dist > worstDist {
					worstDist = nearest[j].dist
					worstIdx = j
				}
			}
		}
	}

	// Coleta candidatos das posting lists via sync.Pool (evita alocação por request).
	cp := idx.candidatePool.Get().(*[]int32)
	candidates := (*cp)[:0]
	for _, c := range nearest {
		candidates = append(candidates, idx.invLists[c.i]...)
	}

	if len(candidates) == 0 {
		*cp = candidates
		idx.candidatePool.Put(cp)
		return true, 0
	}

	// KNN local entre candidatos — distância ao quadrado int32, sem sqrt.
	type neighbor struct {
		dist    int32
		isFraud bool
	}
	best := make([]neighbor, 0, idx.k)
	for _, ci := range candidates {
		d := squaredDistInt(qQuery, idx.vectors[int(ci)*dim:(int(ci)+1)*dim])
		if len(best) < idx.k {
			best = append(best, neighbor{d, idx.labels[ci]})
			continue
		}
		wi := 0
		for j := 1; j < len(best); j++ {
			if best[j].dist > best[wi].dist {
				wi = j
			}
		}
		if d < best[wi].dist {
			best[wi] = neighbor{d, idx.labels[ci]}
		}
	}

	*cp = candidates
	idx.candidatePool.Put(cp)

	frauds := 0
	for _, n := range best {
		if n.isFraud {
			frauds++
		}
	}
	score = float64(frauds) / float64(len(best))
	return score < idx.threshold, score
}

// squaredDistInt computa a soma dos quadrados das diferenças entre dois vetores int8.
// Overflow seguro: max = vecSize × 254² = 14 × 64516 ≈ 903k << math.MaxInt32.
func squaredDistInt(a, b []int8) int32 {
	var sum int32
	for i := range a {
		d := int32(a[i]) - int32(b[i])
		sum += d * d
	}
	return sum
}

// quantizeFloat64 mapeia um valor float64 em [-1, 1] para int8 em [-127, 127].
func quantizeFloat64(v float64) int8 {
	scaled := math.Round(v * 127.0)
	if scaled > 127 {
		return 127
	}
	if scaled < -127 {
		return -127
	}
	return int8(scaled)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
