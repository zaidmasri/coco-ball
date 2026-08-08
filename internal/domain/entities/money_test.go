package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

func usd(n int64) entities.Money {
	m, err := entities.NewMoney(n, entities.USD)
	if err != nil {
		panic(err)
	}
	return m
}

func TestMoney_Arithmetic(t *testing.T) {
	a := usd(1000) // $1,000
	b := usd(250)  // $250

	if got := a.Add(b).MinorUnits(); got != 1250 {
		t.Errorf("Add: want 1250 got %d", got)
	}
	if got := a.Sub(b).MinorUnits(); got != 750 {
		t.Errorf("Sub: want 750 got %d", got)
	}
	if got := a.Neg().MinorUnits(); got != -1000 {
		t.Errorf("Neg: want -1000 got %d", got)
	}
	if got := a.Mul(2.0).MinorUnits(); got != 2000 {
		t.Errorf("Mul(2.0): want 2000 got %d", got)
	}
	if got := a.Div(4).MinorUnits(); got != 250 {
		t.Errorf("Div(4): want 250 got %d", got)
	}
}

func TestMoney_Predicates(t *testing.T) {
	if !usd(0).IsZero() {
		t.Error("zero should be IsZero")
	}
	if !usd(-1).IsNegative() {
		t.Error("-1 should be IsNegative")
	}
	if !usd(1).IsPositive() {
		t.Error("1 should be IsPositive")
	}
}

func TestMoney_Comparisons(t *testing.T) {
	a := usd(500)
	b := usd(300)

	if !b.Less(a) {
		t.Error("300 should be Less than 500")
	}
	if !b.LessOrEqual(a) {
		t.Error("300 should be LessOrEqual 500")
	}
	if !b.LessOrEqual(b) {
		t.Error("300 should be LessOrEqual 300")
	}
	if !a.Equal(usd(500)) {
		t.Error("500 should Equal 500")
	}
}

func TestMoney_ZeroValueAdd(t *testing.T) {
	// var total Money (zero value) should accumulate correctly.
	var total entities.Money
	total = total.Add(usd(100))
	total = total.Add(usd(200))
	if total.MinorUnits() != 300 {
		t.Errorf("zero-value accumulation: want 300 got %d", total.MinorUnits())
	}
}

func TestMoney_String(t *testing.T) {
	cases := []struct {
		amount int64
		want   string
	}{
		{0, "$0"},
		{100, "$100"},
		{1234, "$1234"},
		{-50, "-$50"},
		{-1234, "-$1234"},
	}
	for _, tc := range cases {
		got := usd(tc.amount).String()
		if got != tc.want {
			t.Errorf("amount=%d: want %q got %q", tc.amount, tc.want, got)
		}
	}
}

func TestMoney_JSON(t *testing.T) {
	m := usd(4999)

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	want := `{"minor_units":4999,"currency":"USD"}`
	if string(data) != want {
		t.Errorf("want JSON %s got %s", want, data)
	}

	var got entities.Money
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON (struct): %v", err)
	}
	if got.MinorUnits() != 4999 {
		t.Errorf("want 4999 got %d", got.MinorUnits())
	}
}

func TestMoney_JSON_LegacyPlainInt(t *testing.T) {
	// Backward compatibility: plain JSON integer (old format).
	var got entities.Money
	if err := json.Unmarshal([]byte(`4999`), &got); err != nil {
		t.Fatalf("UnmarshalJSON (plain int): %v", err)
	}
	if got.MinorUnits() != 4999 {
		t.Errorf("want 4999 got %d", got.MinorUnits())
	}
	if got.Currency() != entities.USD {
		t.Errorf("want USD got %s", got.Currency())
	}
}

func TestNewMoney_UnsupportedCurrency(t *testing.T) {
	_, err := entities.NewMoney(100, "XYZ")
	if err == nil {
		t.Error("expected error for unsupported currency")
	}
}
