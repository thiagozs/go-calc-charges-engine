package calc

import (
	"math"
	"time"

	"github.com/thiagozs/go-credit-card-engine/config"
	"github.com/thiagozs/go-credit-card-engine/domain"
)

type RotativeResult struct {
	Principal    domain.Money
	Interest     domain.Money
	IOF          domain.Money
	LateFee      domain.Money
	LateInterest domain.Money
	Charges      domain.Money
	Total        domain.Money
	Days         int
	ChargedDays  int
	ChargeCapped bool
}

func CalculateRotative(
	balance domain.RotativeBalance,
	calcDate time.Time,
	iofCfg config.IOFConfig,
	intCfg config.InterestConfig,
	lateFeeCfg config.LateFeeConfig,
	lateInterestCfg config.LateInterestConfig,
	rulesCfg config.RotativeRulesConfig,
) RotativeResult {
	days := int(calcDate.Sub(balance.StartDate).Hours() / 24)
	if days < 0 {
		days = 0
	}

	chargedDays := days
	if rulesCfg.MaxDays > 0 && chargedDays > rulesCfg.MaxDays {
		chargedDays = rulesCfg.MaxDays
	}

	interest := CalculateRotativeInterest(balance.Principal, chargedDays, intCfg)
	iof := CalculateIOF(balance.Principal, chargedDays, iofCfg)

	var lateFee domain.Money
	var lateInterest domain.Money
	if days > 0 {
		lateFee = CalculateLateFee(balance.Principal, lateFeeCfg)
		lateInterest = CalculateLateInterest(balance.Principal, chargedDays, lateInterestCfg)
	}

	charges := interest + iof + lateFee + lateInterest
	chargeCapped := false
	if rulesCfg.MaxChargeRate > 0 {
		maxCharges := domain.Money(math.Round(float64(balance.Principal) * rulesCfg.MaxChargeRate))
		if charges > maxCharges {
			chargeCapped = true
			interest = max(maxCharges-(iof+lateFee+lateInterest), 0)
			charges = interest + iof + lateFee + lateInterest
		}
	}

	total := balance.Principal + charges

	return RotativeResult{
		Principal:    balance.Principal,
		Interest:     interest,
		IOF:          iof,
		LateFee:      lateFee,
		LateInterest: lateInterest,
		Charges:      charges,
		Total:        total,
		Days:         days,
		ChargedDays:  chargedDays,
		ChargeCapped: chargeCapped,
	}
}
