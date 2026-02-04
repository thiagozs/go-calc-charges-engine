# go-calc-charges-engine

Motor simples para cálculo de encargos de cartão de crédito (Brasil), com IOF, juros rotativo, juros de mora, multa e regras de rotativo (30 dias e teto de 100%).

## Visao geral

- Foco em calculos determinísticos usando `domain.Money` (inteiro em centavos).
- APIs pequenas: funções no pacote `calc` e serviço em `service`.
- Exemplos prontos em `exemplos/` para validar o fluxo.

## Estrutura do projeto

- `calc`: funções de cálculo (IOF, juros, multa, rotativo e amortização).
- `config`: structs de taxas e regras.
- `domain`: tipos base (Money, Invoice, Transaction, RotativeBalance).
- `service`: serviço de alto nível para rotativo.
- `exemplos`: cenários executáveis.

## Tipos principais

```go
// domain.Money usa centavos como unidade
type Money int64

type Transaction struct {
	ID            string
	Amount        Money
	Date          time.Time
	International bool
}

type Invoice struct {
	ID          string
	ClosingDate time.Time
	DueDate     time.Time
	TotalAmount Money
	PaidAmount  Money
}

type RotativeBalance struct {
	Principal Money
	StartDate time.Time
}
```

## Configuracoes

```go
type IOFConfig struct {
	DailyRate      float64 // exemplo: 0.000082 (0,0082% a.d.)
	AdditionalRate float64 // exemplo: 0.0038 (0,38% fixo)
	MaxAnnualRate  float64 // exemplo: 0.0408 (4,08% a.a.)
}

type InterestConfig struct {
	MonthlyRate float64 // juros rotativo
}

type LateFeeConfig struct {
	Rate float64 // multa
}

type LateInterestConfig struct {
	MonthlyRate float64 // juros de mora
}

type RotativeRulesConfig struct {
	MaxDays       int     // regra dos 30 dias
	MaxChargeRate float64 // teto de 100% (1.0)
}

type InternationalIOFConfig struct {
	Rate float64 // IOF por transacao internacional
}
```

## Metodos disponiveis (publicos)

Pacote `calc`:

```go
func CalculateIOF(principal domain.Money, days int, cfg config.IOFConfig) domain.Money
func CalculateRotativeInterest(principal domain.Money, days int, cfg config.InterestConfig) domain.Money
func CalculateLateFee(principal domain.Money, cfg config.LateFeeConfig) domain.Money
func CalculateLateInterest(principal domain.Money, days int, cfg config.LateInterestConfig) domain.Money
func CalculateInternationalIOF(amount domain.Money, cfg config.InternationalIOFConfig) domain.Money
func CalculateRotative(
	balance domain.RotativeBalance,
	calcDate time.Time,
	iofCfg config.IOFConfig,
	intCfg config.InterestConfig,
	lateFeeCfg config.LateFeeConfig,
	lateInterestCfg config.LateInterestConfig,
	rulesCfg config.RotativeRulesConfig,
) RotativeResult
func ApplyPayment(
	total domain.Money,
	iof domain.Money,
	interest domain.Money,
	lateInterest domain.Money,
	lateFee domain.Money,
	principal domain.Money,
	payment domain.Money,
) AmortizationResult
```

Pacote `service`:

```go
type RotativeService struct {
	IOFConfig          config.IOFConfig
	InterestConfig     config.InterestConfig
	LateFeeConfig      config.LateFeeConfig
	LateInterestConfig config.LateInterestConfig
	RulesConfig        config.RotativeRulesConfig
}

func (s *RotativeService) Calculate(balance domain.RotativeBalance, at time.Time) calc.RotativeResult
```

## Como usar (exemplo rapido)

```go
iofCfg := config.IOFConfig{DailyRate: 0.000082, AdditionalRate: 0.0038, MaxAnnualRate: 0.0408}
intCfg := config.InterestConfig{MonthlyRate: 0.12}
lateFeeCfg := config.LateFeeConfig{Rate: 0.02}
lateInterestCfg := config.LateInterestConfig{MonthlyRate: 0.01}
rulesCfg := config.RotativeRulesConfig{MaxDays: 30, MaxChargeRate: 1.0}

balance := domain.RotativeBalance{
	Principal: 100_000, // R$ 1.000,00
	StartDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
}

result := calc.CalculateRotative(
	balance,
	time.Date(2024, 2, 25, 0, 0, 0, 0, time.UTC),
	iofCfg,
	intCfg,
	lateFeeCfg,
	lateInterestCfg,
	rulesCfg,
)

// Aplicar pagamento (ordem: IOF -> juros -> mora -> multa -> principal)
payment := domain.Money(20_000)
amort := calc.ApplyPayment(
	result.Total,
	result.IOF,
	result.Interest,
	result.LateInterest,
	result.LateFee,
	result.Principal,
	payment,
)
```

## Exemplos e logs (fluxo por acao)

### exemplos/on_time

```text
=== Linha do tempo ===
Ciclo: 2024-01-01 -> 2024-01-31
Vencimento: 2024-02-10
Pagamento: 2024-02-10 10:30

=== Transacoes no ciclo ===
- 2024-01-05 | R$ 350.00
- 2024-01-12 | R$ 125.00 (internacional)
- 2024-01-20 | R$ 89.00
- 2024-01-30 | R$ 52.00

=== Fatura no fechamento ===
- Principal: R$ 616.00
- IOF internacional: R$ 4.38
- Total no fechamento: R$ 620.38
- Fechamento: 2024-01-31 00:00
- Vencimento: 2024-02-10

=== Pagamento no vencimento (sem encargos) ===
- Pago: R$ 620.38
- Em aberto: R$ 0.00
```

### exemplos/partial_on_time

```text
=== Linha do tempo ===
Ciclo: 2024-01-01 -> 2024-01-31
Vencimento: 2024-02-10
Pagamento parcial: 2024-02-10 18:45
Pagamento final: 2024-02-25 10:00

=== Transacoes no ciclo ===
- 2024-01-05 | R$ 350.00
- 2024-01-12 | R$ 125.00 (internacional)
- 2024-01-20 | R$ 89.00
- 2024-01-30 | R$ 52.00

=== Fatura no fechamento ===
- Principal: R$ 616.00
- IOF internacional: R$ 4.38
- Total no fechamento: R$ 620.38
- Fechamento: 2024-01-31 00:00
- Vencimento: 2024-02-10

=== Pagamento parcial no vencimento ===
- Pago: R$ 400.00
- Em aberto: R$ 220.38

=== Encargos ate a quitacao ===
- Dias em atraso: 15 (cobrados 15)
- Saldo em atraso: R$ 220.38
- IOF: R$ 1.11
- Juros rotativo: R$ 13.22
- Juros de mora: R$ 1.10
- Multa: R$ 4.41
- Total encargos: R$ 19.84
- Teto aplicado: nao

=== Fatura final apos encargos ===
- Total: R$ 240.22
```

### exemplos/with_charges

```text
=== Linha do tempo ===
Ciclo: 2024-01-01 -> 2024-01-31
Vencimento: 2024-02-10
Pagamento: 2024-02-25

=== Transacoes no ciclo ===
- 2024-01-05 | R$ 350.00
- 2024-01-12 | R$ 125.00 (internacional)
- 2024-01-20 | R$ 89.00

=== Fatura no fechamento ===
- Principal: R$ 564.00
- IOF internacional: R$ 4.38
- Total no fechamento: R$ 568.38
- Fechamento: 2024-01-31
- Vencimento: 2024-02-10

=== Pagamento ===
- Valor pago: R$ 200.00

=== Encargos apos vencimento ===
- Dias em atraso: 15 (cobrados 15)
- Saldo em atraso: R$ 368.38
- IOF: R$ 1.85
- Juros rotativo: R$ 22.10
- Juros de mora: R$ 1.84
- Multa: R$ 7.37
- Total encargos: R$ 33.16
- Teto aplicado: nao

=== Fatura final ===
- Total: R$ 401.54
- Pago: R$ 200.00
- Em aberto: R$ 201.54
```

## Como executar os exemplos

```bash
go run ./exemplos/on_time
go run ./exemplos/partial_on_time
go run ./exemplos/with_charges
```

## License

Este projeto é distribuído sob a Licença MIT. Consulte o arquivo LICENSE para obter detalhes.

## Autor

2026, Thiago Zilli Sarmento ❤️