package detector

import (
	"math"
	"time"

	"github.com/lhupalo/rbc-2026/internal/models"
)

const vecSize = 14

func vectorize(req *models.FraudScoreRequest, mcc map[string]float64, norm normConstants) [vecSize]float64 {
	var v [vecSize]float64

	// 0 - amount
	v[0] = clamp(req.Transaction.Amount / norm.MaxAmount)

	// 1 - installments
	v[1] = clamp(float64(req.Transaction.Installments) / norm.MaxInstallments)

	// 2 - amount_vs_avg
	if req.Customer.AvgAmount > 0 {
		v[2] = clamp((req.Transaction.Amount / req.Customer.AvgAmount) / norm.AmountVsAvgRatio)
	}

	// 3 - hour_of_day (UTC)
	// 4 - day_of_week (Mon=0, Sun=6)
	if t, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt); err == nil {
		t = t.UTC()
		v[3] = float64(t.Hour()) / 23.0
		// Go: Sun=0, Mon=1 ... Sat=6 → spec: Mon=0 ... Sun=6
		v[4] = float64((int(t.Weekday())+6)%7) / 6.0
	}

	// 5 - minutes_since_last_tx  |  6 - km_from_last_tx
	if req.LastTransaction != nil {
		if txTime, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt); err == nil {
			if lastTime, err := time.Parse(time.RFC3339, req.LastTransaction.Timestamp); err == nil {
				minutes := txTime.Sub(lastTime).Minutes()
				v[5] = clamp(minutes / norm.MaxMinutes)
			} else {
				v[5] = -1
			}
		} else {
			v[5] = -1
		}
		v[6] = clamp(req.LastTransaction.KmFromCurrent / norm.MaxKm)
	} else {
		v[5] = -1
		v[6] = -1
	}

	// 7 - km_from_home
	v[7] = clamp(req.Terminal.KmFromHome / norm.MaxKm)

	// 8 - tx_count_24h
	v[8] = clamp(float64(req.Customer.TxCount24h) / norm.MaxTxCount24h)

	// 9 - is_online
	if req.Terminal.IsOnline {
		v[9] = 1
	}

	// 10 - card_present
	if req.Terminal.CardPresent {
		v[10] = 1
	}

	// 11 - unknown_merchant
	v[11] = 1
	for _, m := range req.Customer.KnownMerchants {
		if m == req.Merchant.ID {
			v[11] = 0
			break
		}
	}

	// 12 - mcc_risk
	if risk, ok := mcc[req.Merchant.MCC]; ok {
		v[12] = risk
	} else {
		v[12] = 0.5
	}

	// 13 - merchant_avg_amount
	v[13] = clamp(req.Merchant.AvgAmount / norm.MaxMerchantAvgAmount)

	return v
}

func clamp(x float64) float64 {
	return math.Max(0, math.Min(1, x))
}
