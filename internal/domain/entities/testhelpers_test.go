package entities

func mustUSD(n int64) Money {
	m, err := NewMoney(n, USD)
	if err != nil {
		panic(err)
	}
	return m
}
