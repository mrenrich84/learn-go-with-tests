package badbank

import r "github.com/mrenrich84/learn-go-with-tests/sum"

type Transaction struct {
	From string
	To   string
	Sum  float64
}

func NewTransaction(from, to Account, sum float64) Transaction {
	return Transaction{From: from.Name, To: to.Name, Sum: sum}
}

type Account struct {
	Name    string
	Balance float64
}

func NewBalanceFor(account Account, transactions []Transaction) Account {
	return r.Reduce(
		transactions,
		applyTransaction,
		account,
	)
}

func applyTransaction(a Account, transaction Transaction) Account {
	if transaction.From == a.Name {
		a.Balance -= transaction.Sum
	}
	if transaction.To == a.Name {
		a.Balance += transaction.Sum
	}
	return a
}