# Guia de Consulta: Encargos e IOF no Cartão de Crédito (Brasil)

Resumo das regras implementadas pela engine, com a fonte de cada uma. Quando uma
regra não foi confirmada em texto oficial, isso está indicado na seção
[Pendências de verificação](#pendências-de-verificação).

**Fontes oficiais consultadas**

- [Decreto 6.306/2007 (Regulamento do IOF), texto consolidado](https://www.planalto.gov.br/ccivil_03/_ato2007-2010/2007/decreto/d6306compilado.htm) — Receita Federal / Presidência.
- [Resolução CMN 4.549/2017, consolidada com a Res. CMN 5.112/2023](https://normativos.bcb.gov.br/Lists/Normativos/Attachments/50330/Res_4549_v2_L.pdf) — Banco Central.
- [Lei 14.690/2023, art. 28](https://www.planalto.gov.br/ccivil_03/_ato2023-2026/2023/lei/l14690.htm) — teto de 100%.

Todas as taxas usam ponto fixo de 6 casas (`domain.Rate`): `0,0082% = 82`, `0,38% = 3800`.

---

## 1. IOF

### A. IOF-crédito (financiamento de fatura: rotativo e parcelamento) — pessoa física

Fonte: Decreto 6.306/2007, art. 7º.

| Componente         | Valor                  | Regra                                                                    |
| ------------------ | ---------------------- | ------------------------------------------------------------------------ |
| Alíquota diária    | **0,0082% ao dia**     | Sobre o principal, multiplicada pelo número de dias (art. 7º, I e VII)   |
| Limite de dias     | **365 dias**           | O IOF não excede o principal × alíquota diária × 365, mais o adicional (art. 7º, § 1º) |
| Alíquota adicional | **0,38%**              | Uma única vez, independentemente do prazo (art. 7º, § 15)                |

```
IOF = Principal × 0,000082 × min(dias, 365) + Principal × 0,0038
```

Teto efetivo: `0,0082% × 365 + 0,38% = 3,373%` do principal.

> **Atenção:** versões anteriores deste documento citavam 0,01118% ao dia e
> "máximo de 4,08% ao ano". Esses valores vêm do § 22 do art. 7º, que fixou uma
> alíquota **reduzida e temporária** para fatos geradores entre 20/09/2021 e
> 31/12/2021 (0,01118% × 365 ≈ 4,08%). Não estão mais em vigor.

No parcelamento, o IOF é calculado por parcela, sobre o principal de cada uma e os
dias entre a compra e o vencimento daquela parcela.

### B. Compras internacionais (transação individual)

- **Alíquota:** 3,5%, por transação.
- **Fórmula:** `Valor da Compra × 0,035`
- Ver [Pendências de verificação](#pendências-de-verificação).

---

## 2. Encargos do crédito rotativo

Os encargos incidem sobre o saldo devedor (valor da fatura menos pagamentos).

### A. Juros remuneratórios

Definidos pela instituição (campo "Taxa de Juros" da fatura).

```
Juros = Saldo × (Taxa Mensal / 30) × Dias
```

### B. Multa por atraso

Alíquota usual de 2% (Código de Defesa do Consumidor, art. 52, § 1º).

```
Multa = Valor em Atraso × 0,02
```

### C. Juros de mora

Alíquota usual de 1% ao mês, pro rata die.

```
Mora = Valor em Atraso × (0,01 / 30) × Dias
```

---

## 3. Regras do Banco Central

Fonte: Resolução CMN 4.549/2017 (com a Res. CMN 5.112/2023) e Lei 14.690/2023, art. 28.

1. **Prazo no rotativo (art. 1º):** o saldo não pago no vencimento só pode ser
   financiado no rotativo **até o vencimento da fatura seguinte** (≈ 30 dias). Depois
   disso, o saldo pode ser financiado em linha de **parcelamento**, em condições mais
   vantajosas ao cliente (art. 2º). O parcelamento é uma faculdade da instituição, não
   uma obrigação.
2. **Teto de 100% (art. 2º-A):** o total cobrado a título de **juros e encargos
   financeiros** não pode exceder o **valor original da dívida**.
   - *Juros* = juros remuneratórios (inciso II).
   - *Encargos financeiros* = multa e juros de mora por atraso, mais tarifas e
     comissões da operação (inciso III).
   - **IOF não entra** nessa soma: é tributo, não está nas definições.
   - *Valor original da dívida* = montante inicial da operação de rotativo (inciso IV e
     parágrafo único, I).
   - A contagem é **acumulada desde o início da operação** de rotativo (parágrafo
     único, II), e na renegociação desconta-se o que já foi pago (art. 2º-C).
   - Vale para operações posteriores ao prazo de 90 dias da Lei 14.690/2023 (art. 2º-D).
3. **Ordem de abatimento de pagamento (convenção da engine):** IOF → juros → juros de
   mora → multa → principal. Não vem do normativo do BC.
4. **CET (Custo Efetivo Total):** deve ser informado ao cliente (juros + tributos + tarifas).

### Como a engine aplica o teto

```
limite = max(ValorOriginal × 100% − EncargosJáCobrados, 0)
se juros + mora + multa > limite → corta o excesso de: juros, depois mora, depois multa
IOF = calculado à parte, nunca é cortado pelo teto
```

A ordem de corte é uma escolha da engine; o normativo não define qual componente
reduzir. `RotativeBalance.OriginalDebt` e `RotativeBalance.ChargesAlreadyApplied`
alimentam o teto acumulado (`ChargesAlreadyApplied` = soma de
`RotativeResult.FinancialCharges` dos períodos anteriores).

---

## Tabela de referência (implementada na engine)

| Categoria         | Encargo                     | Valor            | Periodicidade        | Fonte                          |
| ----------------- | --------------------------- | ---------------- | -------------------- | ------------------------------ |
| **IOF crédito**   | Diário (pessoa física)      | **0,0082%**      | Ao dia, até 365 dias | Decreto 6.306, art. 7º         |
| **IOF crédito**   | Adicional                   | **0,38%**        | Uma única vez        | Decreto 6.306, art. 7º, § 15   |
| **IOF exterior**  | Compras internacionais      | **3,5%**         | Por transação        | ver pendências                 |
| **Penalidades**   | Multa por atraso            | **2,0%**         | Uma única vez        | CDC art. 52, § 1º (a confirmar)|
| **Penalidades**   | Juros de mora               | **1,0%**         | Ao mês (pro rata)    | a confirmar                    |
| **Financiamento** | Juros do rotativo           | Variável         | Definida pela instituição | —                         |
| **Teto**          | Juros + multa + mora        | **100%**         | Acumulado            | Res. CMN 4.549, art. 2º-A      |

---

## Pendências de verificação

- **Multa de 2% e juros de mora de 1% a.m.:** o texto da Res. CMN 4.549 apenas
  classifica multa e mora como "encargos financeiros"; **não fixa os percentuais**. Os
  valores usados são os usuais de mercado e do CDC, ainda não confirmados em fonte
  oficial por esta documentação.
- **IOF internacional (3,5%):** confirmado apenas em fontes secundárias, que indicam que
  o cronograma de redução (2,38% / 1,38% / 0%) foi revogado em 2025 (Decreto 12.499/2025).
  Conferir no texto oficial antes de depender do valor.
- **Decreto Legislativo 176/2025 e ADC 96:** o texto consolidado do Decreto 6.306 faz
  referência a ambos; não foi verificado se alteram o IOF-crédito de pessoa física.
- **Pessoa jurídica:** não implementada pela engine. O consolidado traz 0,0082% ao dia
  (redação do Decreto 12.499/2025); valores anteriores, como 0,0041%, estão desatualizados.
