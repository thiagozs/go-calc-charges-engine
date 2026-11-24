package domain

import "time"

type RotativeBalance struct {
	Principal Money
	StartDate time.Time

	// OriginalDebt is the "valor original da divida": the initial amount of the rotativo
	// operation, which is the base of the 100% cap on interest and charges
	// (Res. CMN 4.549, art. 2-A, IV and par. unico). Zero means "use Principal".
	OriginalDebt Money

	// ChargesAlreadyApplied is the total of interest, late fee and late interest
	// (never IOF) already charged since the start of the operation: the cap is cumulative (art. 2-A, par. unico, II; art. 2-C). Feed it with the sum of
	// RotativeResult.FinancialCharges of previous periods.
	ChargesAlreadyApplied Money
}
