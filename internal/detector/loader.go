package detector

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
)

type normConstants struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio     float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKm                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

func Load(dataDir string) (*Detector, error) {
	norm, err := loadNorm(dataDir + "/normalization.json")
	if err != nil {
		return nil, fmt.Errorf("normalization: %w", err)
	}

	mcc, err := loadMCC(dataDir + "/mcc_risk.json")
	if err != nil {
		return nil, fmt.Errorf("mcc_risk: %w", err)
	}

	vectors, labels, err := loadReferences(dataDir + "/references.json.gz")
	if err != nil {
		return nil, fmt.Errorf("references: %w", err)
	}

	return &Detector{
		vectors: vectors,
		labels:  labels,
		mcc:     mcc,
		norm:    norm,
	}, nil
}

func loadNorm(path string) (normConstants, error) {
	f, err := os.Open(path)
	if err != nil {
		return normConstants{}, err
	}
	defer f.Close()

	var n normConstants
	if err := json.NewDecoder(f).Decode(&n); err != nil {
		return normConstants{}, err
	}
	return n, nil
}

func loadMCC(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]float64
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func loadReferences(path string) ([][]float64, []bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, err
	}
	defer gz.Close()

	dec := json.NewDecoder(gz)

	if _, err := dec.Token(); err != nil { // consume '['
		return nil, nil, err
	}

	var vectors [][]float64
	var labels []bool

	var ref struct {
		Vector []float64 `json:"vector"`
		Label  string    `json:"label"`
	}

	for dec.More() {
		if err := dec.Decode(&ref); err != nil {
			return nil, nil, err
		}
		vec := make([]float64, len(ref.Vector))
		copy(vec, ref.Vector)
		vectors = append(vectors, vec)
		labels = append(labels, ref.Label == "fraud")
	}

	return vectors, labels, nil
}
