# go-calc-charges-engine

Motor simples para cálculo de encargos de cartão de crédito (Brasil), com IOF, juros rotativo, juros de mora, multa, regras de rotativo (30 dias e teto de 100%) e parcelamento (Tabela Price / sem juros).

## Visao geral

- Foco em calculos determinísticos usando `domain.Money` (inteiro em centavos).
- Aritmética inteira pura: taxas representadas como `domain.Rate` (inteiro por milhão), sem uso de float64.
- APIs pequenas: funções no pacote `calc` e serviços em `service`.
- Exemplos prontos em `exemplos/` para validar o fluxo.

## Estrutura do projeto

- `calc`: funções de cálculo (IOF, juros, multa, rotativo, parcelamento e amortização).
- `config`: structs de taxas e regras.
- `domain`: tipos base (Money, Rate, Invoice, Transaction, RotativeBalance, InstallmentPlan).
- `service`: serviços de alto nível para rotativo e parcelamento.
- `exemplos`: cenários executáveis.

## Tipos principais

```go
// domain.Money usa centavos como unidade
type Money int64

// domain.Rate representa taxa com 6 casas decimais (denominador 1.000.000)
// Exemplo: 0.000082 = Rate(82), 0.12 = Rate(120_000), 1.0 = Rate(1_000_000)
type Rate int64
const RateDenominator int64 = 1_000_000

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

type InstallmentPlan struct {
	TotalAmount   Money
	TotalIOF      Money
	TotalInterest Money
	TotalWithIOF  Money
	Installments  []Installment
}

type Installment struct {
	Number    int
	DueDate   time.Time
	Principal Money
	Interest  Money
	IOF       Money
	Amount    Money
}
```

## Configuracoes

Todas as taxas usam `domain.Rate` (inteiro por milhão). Para converter: `taxa_decimal * 1_000_000`.

```go
type IOFConfig struct {
	DailyRate      domain.Rate // exemplo: 82 (0,0082% a.d.)
	AdditionalRate domain.Rate // exemplo: 3_800 (0,38% fixo)
	MaxAnnualRate  domain.Rate // exemplo: 40_800 (4,08% a.a.)
}

type InterestConfig struct {
	MonthlyRate domain.Rate // juros rotativo, ex: 120_000 (12%)
}

type LateFeeConfig struct {
	Rate domain.Rate // multa, ex: 20_000 (2%)
}

type LateInterestConfig struct {
	MonthlyRate domain.Rate // juros de mora, ex: 10_000 (1%)
}

type RotativeRulesConfig struct {
	MaxDays       int         // regra dos 30 dias
	MaxChargeRate domain.Rate // teto de 100%: 1_000_000
}

type InternationalIOFConfig struct {
	Rate domain.Rate // IOF por transacao internacional, ex: 35_000 (3,5%)
}

type InstallmentConfig struct {
	MonthlyRate domain.Rate // juros do parcelamento, 0 = sem juros
}
```

## Configuracao via variaveis de ambiente

A lib suporta carregar configuracoes via variaveis de ambiente com valores padrao.
Os valores sao inteiros (Rate por milhão).

Principais variaveis:

- `IOF_DAILY_RATE` (default 82)
- `IOF_ADDITIONAL_RATE` (default 3800)
- `IOF_MAX_ANNUAL_RATE` (default 40800)
- `ROTATIVE_MONTHLY_RATE` (default 120000)
- `LATE_FEE_RATE` (default 20000)
- `LATE_INTEREST_MONTHLY_RATE` (default 10000)
- `ROTATIVE_MAX_DAYS` (default 30)
- `ROTATIVE_MAX_CHARGE_RATE` (default 1000000)
- `INTERNATIONAL_IOF_RATE` (default 35000)
- `INSTALLMENT_MONTHLY_RATE` (default 0)

Exemplo de uso:

```go
cfg, err := config.LoadFromEnv()
if err != nil {
	log.Fatal(err)
}

svc := service.NewRotativeService(cfg)
instSvc := service.NewInstallmentService(cfg)
```

Ou direto:

```go
svc, err := service.NewRotativeServiceFromEnv()
if err != nil {
	log.Fatal(err)
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
func CalculateInstallmentPlan(
	totalAmount domain.Money,
	numInstallments int,
	purchaseDate time.Time,
	firstDueDate time.Time,
	iofCfg config.IOFConfig,
	instCfg config.InstallmentConfig,
) domain.InstallmentPlan
```

Pacote `service`:

```go
type RotativeService struct { ... }
func (s *RotativeService) Calculate(balance domain.RotativeBalance, at time.Time) calc.RotativeResult

type InstallmentService struct { ... }
func (s *InstallmentService) Calculate(amount domain.Money, numInstallments int, purchaseDate time.Time, firstDueDate time.Time) domain.InstallmentPlan
```

## Validacao e audit trail

- **Validacao de inputs:** a engine nao valida inputs negativos. A validacao (principal >= 0, dias >= 0, datas validas) e responsabilidade do caller (ledger).
- **Audit trail:** as structs de resultado (`RotativeResult`, `AmortizationResult`, `InstallmentPlan`) contem o detalhamento completo dos calculos. O caller (ledger) deve persistir essas structs como entradas no historico para fins de auditoria.

## Como usar (exemplo rapido)

```go
iofCfg := config.IOFConfig{DailyRate: 82, AdditionalRate: 3_800, MaxAnnualRate: 40_800}
intCfg := config.InterestConfig{MonthlyRate: 120_000}
lateFeeCfg := config.LateFeeConfig{Rate: 20_000}
lateInterestCfg := config.LateInterestConfig{MonthlyRate: 10_000}
rulesCfg := config.RotativeRulesConfig{MaxDays: 30, MaxChargeRate: 1_000_000}

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

### Parcelamento

```go
iofCfg := config.IOFConfig{DailyRate: 82, AdditionalRate: 3_800, MaxAnnualRate: 40_800}

// Sem juros: 10x de R$ 1.000,00
plan := calc.CalculateInstallmentPlan(
	100_000, 10,
	time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // data da compra
	time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), // vencimento da 1a parcela
	iofCfg,
	config.InstallmentConfig{MonthlyRate: 0}, // 0 = sem juros
)

// Com juros: 12x a 1.99% a.m.
planJuros := calc.CalculateInstallmentPlan(
	100_000, 12,
	time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
	iofCfg,
	config.InstallmentConfig{MonthlyRate: 19_900}, // 1.99% = 19_900
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

### exemplos/installments (parcelamento)

```text
=== Parcelamento sem juros - 10x de R$ 1.000,00 ===

  Parcela  1 | Venc: 2024-02-10 | Principal: R$ 100.00 | IOF: R$ 0.59 | Total: R$ 100.59
  Parcela  2 | Venc: 2024-03-10 | Principal: R$ 100.00 | IOF: R$ 0.83 | Total: R$ 100.83
  Parcela  3 | Venc: 2024-04-10 | Principal: R$ 100.00 | IOF: R$ 1.09 | Total: R$ 101.09
  Parcela  4 | Venc: 2024-05-10 | Principal: R$ 100.00 | IOF: R$ 1.33 | Total: R$ 101.33
  Parcela  5 | Venc: 2024-06-10 | Principal: R$ 100.00 | IOF: R$ 1.59 | Total: R$ 101.59
  Parcela  6 | Venc: 2024-07-10 | Principal: R$ 100.00 | IOF: R$ 1.83 | Total: R$ 101.83
  Parcela  7 | Venc: 2024-08-10 | Principal: R$ 100.00 | IOF: R$ 2.09 | Total: R$ 102.09
  Parcela  8 | Venc: 2024-09-10 | Principal: R$ 100.00 | IOF: R$ 2.34 | Total: R$ 102.34
  Parcela  9 | Venc: 2024-10-10 | Principal: R$ 100.00 | IOF: R$ 2.59 | Total: R$ 102.59
  Parcela 10 | Venc: 2024-11-10 | Principal: R$ 100.00 | IOF: R$ 2.84 | Total: R$ 102.84

  Total compra: R$ 1000.00
  Total IOF: R$ 17.12
  Total com IOF: R$ 1017.12

=== Parcelamento com juros - 12x de R$ 1.000,00 a 1.99% a.m. ===

  Parcela  1 | Venc: 2024-02-10 | Principal: R$ 74.60 | Juros: R$ 19.90 | IOF: R$ 0.44 | Total: R$ 94.94
  Parcela  2 | Venc: 2024-03-10 | Principal: R$ 76.08 | Juros: R$ 18.42 | IOF: R$ 0.63 | Total: R$ 95.13
  Parcela  3 | Venc: 2024-04-10 | Principal: R$ 77.60 | Juros: R$ 16.90 | IOF: R$ 0.84 | Total: R$ 95.34
  Parcela  4 | Venc: 2024-05-10 | Principal: R$ 79.14 | Juros: R$ 15.36 | IOF: R$ 1.05 | Total: R$ 95.55
  Parcela  5 | Venc: 2024-06-10 | Principal: R$ 80.72 | Juros: R$ 13.78 | IOF: R$ 1.28 | Total: R$ 95.78
  Parcela  6 | Venc: 2024-07-10 | Principal: R$ 82.32 | Juros: R$ 12.18 | IOF: R$ 1.50 | Total: R$ 96.00
  Parcela  7 | Venc: 2024-08-10 | Principal: R$ 83.96 | Juros: R$ 10.54 | IOF: R$ 1.75 | Total: R$ 96.25
  Parcela  8 | Venc: 2024-09-10 | Principal: R$ 85.63 | Juros: R$ 8.87 | IOF: R$ 2.01 | Total: R$ 96.51
  Parcela  9 | Venc: 2024-10-10 | Principal: R$ 87.34 | Juros: R$ 7.16 | IOF: R$ 2.26 | Total: R$ 96.76
  Parcela 10 | Venc: 2024-11-10 | Principal: R$ 89.08 | Juros: R$ 5.42 | IOF: R$ 2.53 | Total: R$ 97.03
  Parcela 11 | Venc: 2024-12-10 | Principal: R$ 90.85 | Juros: R$ 3.65 | IOF: R$ 2.81 | Total: R$ 97.31
  Parcela 12 | Venc: 2025-01-10 | Principal: R$ 92.68 | Juros: R$ 1.82 | IOF: R$ 3.09 | Total: R$ 97.59

  Total compra: R$ 1000.00
  Total juros: R$ 134.00
  Total IOF: R$ 20.19
  Total com juros + IOF: R$ 1154.19
```

## Como executar os exemplos

```bash
go run ./exemplos/on_time
go run ./exemplos/partial_on_time
go run ./exemplos/with_charges
go run ./exemplos/installments (parcelamento)
```

## License

Este projeto é distribuído sob a Licença MIT. Consulte o arquivo [LICENSE](LICENSE) para obter detalhes.

## Autor

2026, Thiago Zilli Sarmento :heart:
