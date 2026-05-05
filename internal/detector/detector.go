package detector

import "github.com/lhupalo/rbc-2026/internal/models"

type Detector struct {
	vectors [][]float64
	labels  []bool
	mcc     map[string]float64
	norm    normConstants
}

func (d *Detector) Score(req *models.FraudScoreRequest) (approved bool, fraudScore float64) {
	arr := vectorize(req, d.mcc, d.norm)
	fraudCount := knn(arr[:], d.vectors, d.labels)
	score := float64(fraudCount) / float64(k)
	return score < threshold, score
}
