package entities

import (
	"encoding/json"
	"fmt"
)

// Currency is the ISO 4217 code of a Money value.
type Currency string

const USD Currency = "USD"

var supportedCurrencies = map[Currency]int{
	USD: 2,
}

// Money is an immutable value object representing a monetary amount stored in
// whole-dollar units (the minor unit for this app). All arithmetic is integer
// arithmetic on those whole-dollar amounts — no floating-point rounding in the
// stored value. Use NewMoney at system boundaries and fromFloatUSD for
// internal projection engine results that come from float64 intermediate math.
type Money struct {
	minorUnits int64
	currency   Currency
}

// NewMoney constructs a Money value for the given currency. Returns an error if
// currency is not in supportedCurrencies.
func NewMoney(minorUnits int64, currency Currency) (Money, error) {
	if _, ok := supportedCurrencies[currency]; !ok {
		return Money{}, fmt.Errorf("%w: unsupported currency %q", ErrValidation, currency)
	}
	return Money{minorUnits: minorUnits, currency: currency}, nil
}

// fromFloatUSD constructs a USD Money by truncating a float64 to int64. Used
// only by the projection engine for intermediate float arithmetic results.
func fromFloatUSD(f float64) Money {
	return Money{minorUnits: int64(f), currency: USD}
}

func (m Money) effectiveCurrency() Currency {
	if m.currency == "" {
		return USD
	}
	return m.currency
}

func (m Money) MinorUnits() int64  { return m.minorUnits }
func (m Money) Currency() Currency { return m.effectiveCurrency() }

func (m Money) IsZero() bool     { return m.minorUnits == 0 }
func (m Money) IsNegative() bool { return m.minorUnits < 0 }
func (m Money) IsPositive() bool { return m.minorUnits > 0 }

func (m Money) Add(other Money) Money {
	if m.effectiveCurrency() != other.effectiveCurrency() {
		panic(fmt.Sprintf("money: currency mismatch %s + %s", m.effectiveCurrency(), other.effectiveCurrency()))
	}
	return Money{minorUnits: m.minorUnits + other.minorUnits, currency: m.effectiveCurrency()}
}

func (m Money) Sub(other Money) Money {
	if m.effectiveCurrency() != other.effectiveCurrency() {
		panic(fmt.Sprintf("money: currency mismatch %s - %s", m.effectiveCurrency(), other.effectiveCurrency()))
	}
	return Money{minorUnits: m.minorUnits - other.minorUnits, currency: m.effectiveCurrency()}
}

func (m Money) Neg() Money {
	return Money{minorUnits: -m.minorUnits, currency: m.effectiveCurrency()}
}

// Mul multiplies by a float64 factor and truncates to int64 (same truncation
// as the projection engine's existing float-to-Money conversions).
func (m Money) Mul(factor float64) Money {
	return Money{minorUnits: int64(float64(m.minorUnits) * factor), currency: m.effectiveCurrency()}
}

// Div performs integer division.
func (m Money) Div(divisor int64) Money {
	return Money{minorUnits: m.minorUnits / divisor, currency: m.effectiveCurrency()}
}

func (m Money) Less(other Money) bool {
	if m.effectiveCurrency() != other.effectiveCurrency() {
		panic(fmt.Sprintf("money: currency mismatch %s < %s", m.effectiveCurrency(), other.effectiveCurrency()))
	}
	return m.minorUnits < other.minorUnits
}

func (m Money) LessOrEqual(other Money) bool {
	if m.effectiveCurrency() != other.effectiveCurrency() {
		panic(fmt.Sprintf("money: currency mismatch %s <= %s", m.effectiveCurrency(), other.effectiveCurrency()))
	}
	return m.minorUnits <= other.minorUnits
}

func (m Money) Equal(other Money) bool {
	if m.effectiveCurrency() != other.effectiveCurrency() {
		panic(fmt.Sprintf("money: currency mismatch %s == %s", m.effectiveCurrency(), other.effectiveCurrency()))
	}
	return m.minorUnits == other.minorUnits
}

func (m Money) String() string {
	if m.minorUnits < 0 {
		return fmt.Sprintf("-$%d", -m.minorUnits)
	}
	return fmt.Sprintf("$%d", m.minorUnits)
}

type moneyJSON struct {
	MinorUnits int64    `json:"minor_units"`
	Currency   Currency `json:"currency"`
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(moneyJSON{MinorUnits: m.minorUnits, Currency: m.effectiveCurrency()})
}

// UnmarshalJSON decodes both the new struct format {"minor_units":N,"currency":"USD"}
// and the legacy plain-integer format for backward compatibility with plan
// data stored before this change.
func (m *Money) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '{' {
		var raw moneyJSON
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*m = Money{minorUnits: raw.MinorUnits, currency: raw.Currency}
		return nil
	}
	var units int64
	if err := json.Unmarshal(data, &units); err != nil {
		return err
	}
	*m = Money{minorUnits: units, currency: USD}
	return nil
}
