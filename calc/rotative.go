package calc

import (
	"time"

	"github.com/thiagozs/go-calc-charges-engine/config"
	"github.com/thiagozs/go-calc-charges-engine/domain"
)

// RotativeResult contains the breakdown of rotative credit charges.
// The caller (ledger) should persist this struct for audit trail purposes.
type RotativeResult struct {
	Principal    domain.Money
	Interest     domain.Money
	IOF          domain.Money
	LateFee      domain.Money
	LateInterest domain.Money
	// FinancialCharges is Interest + LateFee + LateInterest after the cap: the amount
	// that counts toward the 100% limit (IOF is a tax and does not).
	FinancialCharges domain.Money
	// Charges is FinancialCharges + IOF.
	Charges      domain.Money
	Total        domain.Money
	Days         int
	ChargedDays  int
	ChargeCapped bool
}

// CalculateRotative computes all charges for a rotative credit balance.
//
// Returns ErrInvalidAmount, ErrInvalidRate, ErrInvalidDate or ErrOverflow when the
// input is invalid or too large to compute exactly in int64.
func CalculateRotative(
	balance domain.RotativeBalance,
	calcDate time.Time,
	iofCfg config.IOFConfig,
	intCfg config.InterestConfig,
	lateFeeCfg config.LateFeeConfig,
	lateInterestCfg config.LateInterestConfig,
	rulesCfg config.RotativeRulesConfig,
) (RotativeResult, error) {
	if err := validateRotativeInput(balance, calcDate, iofCfg, intCfg, lateFeeCfg, lateInterestCfg, rulesCfg); err != nil {
		return RotativeResult{}, err
	}

	days := daysBetween(balance.StartDate, calcDate)

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

	// Cap (Lei 14.690/2023 art. 28; Res. CMN 4.549 art. 2-A): juros + encargos financeiros
	// (juros remuneratorios, multa, juros de mora, tarifas) cobrados desde o inicio da operacao
	// nao podem exceder o valor original da divida. IOF nao entra na conta.
	// When the limit is hit, the excess is removed from interest first, then late interest,
	// then late fee (the regulation does not prescribe an order).
	chargeCapped := false
	if rulesCfg.MaxChargeRate > 0 {
		base := balance.OriginalDebt
		if base == 0 {
			base = balance.Principal
		}
		limit := max(mulRate(base, rulesCfg.MaxChargeRate)-balance.ChargesAlreadyApplied, 0)
		if excess := interest + lateInterest + lateFee - limit; excess > 0 {
			chargeCapped = true
			for _, part := range []*domain.Money{&interest, &lateInterest, &lateFee} {
				cut := min(excess, *part)
				*part -= cut
				excess -= cut
			}
		}
	}

	financial := interest + lateInterest + lateFee
	charges := financial + iof
	total := balance.Principal + charges

	return RotativeResult{
		Principal:        balance.Principal,
		Interest:         interest,
		IOF:              iof,
		LateFee:          lateFee,
		LateInterest:     lateInterest,
		FinancialCharges: financial,
		Charges:          charges,
		Total:            total,
		Days:             days,
		ChargedDays:      chargedDays,
		ChargeCapped:     chargeCapped,
	}, nil
}

// validateRotativeInput checks inputs and guarantees every int64 product computed
// by CalculateRotative (and the helpers it calls) fits, so nothing overflows silently.
func validateRotativeInput(
	balance domain.RotativeBalance,
	calcDate time.Time,
	iofCfg config.IOFConfig,
	intCfg config.InterestConfig,
	lateFeeCfg config.LateFeeConfig,
	lateInterestCfg config.LateInterestConfig,
	rulesCfg config.RotativeRulesConfig,
) error {
	if balance.Principal < 0 || balance.OriginalDebt < 0 || balance.ChargesAlreadyApplied < 0 {
		return ErrInvalidAmount
	}
	if balance.StartDate.IsZero() || calcDate.IsZero() {
		return ErrInvalidDate
	}
	if iofCfg.DailyRate < 0 || iofCfg.AdditionalRate < 0 || iofCfg.MaxDays < 0 ||
		intCfg.MonthlyRate < 0 || lateFeeCfg.Rate < 0 || lateInterestCfg.MonthlyRate < 0 ||
		rulesCfg.MaxDays < 0 || rulesCfg.MaxChargeRate < 0 {
		return ErrInvalidRate
	}

	days := int64(daysBetween(balance.StartDate, calcDate))
	if rulesCfg.MaxDays > 0 && days > int64(rulesCfg.MaxDays) {
		days = int64(rulesCfg.MaxDays)
	}
	p := int64(balance.Principal)
	if !mulFits(p, int64(intCfg.MonthlyRate), days) ||
		!mulFits(p, int64(iofCfg.DailyRate), days) ||
		!mulFits(p, int64(iofCfg.AdditionalRate)) ||
		!mulFits(p, int64(lateFeeCfg.Rate)) ||
		!mulFits(p, int64(lateInterestCfg.MonthlyRate), days) ||
		!mulFits(capBase(balance), int64(rulesCfg.MaxChargeRate)) {
		return ErrOverflow
	}
	return nil
}

// capBase returns the base of the 100% cap: OriginalDebt, or Principal when unset.
func capBase(b domain.RotativeBalance) int64 {
	if b.OriginalDebt != 0 {
		return int64(b.OriginalDebt)
	}
	return int64(b.Principal)
}
