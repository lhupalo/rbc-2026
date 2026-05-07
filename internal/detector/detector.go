package detector

import "github.com/lhupalo/rbc-2026/internal/models"

type Detector struct {
	index *IVFIndex
	mcc   map[string]float64
	norm  normConstants
}

func (d *Detector) Score(req *models.FraudScoreRequest) (approved bool, fraudScore float64) {
	arr := vectorize(req, d.mcc, d.norm)
	return d.index.Search(arr[:])
}
