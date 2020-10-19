package main

import (
	"fmt"
	"log"

	"github.com/gholib/wallet/pkg/types"
	"github.com/gholib/wallet/pkg/wallet"
)

type testService struct {
	*wallet.Service
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992880806776",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{{
		amount:   1000_00,
		category: "auto",
	}},
}

func main() {
	s := newTestService()

	_, _, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		log.Print(err)
		return
	}

	_, _, err = s.addAcoount(testAccount{
		phone:   "+992935444994",
		balance: 10_000_00,
		payments: []struct {
			amount   types.Money
			category types.PaymentCategory
		}{{
			amount:   1000_00,
			category: "auto",
		}, {
			amount:   1020_00,
			category: "auto",
		}},
	})

	if err != nil {
		log.Print(err)
		return
	}

	payments, err := s.ExportAccountHistory(2)
	// log.Print(payments)
	// log.Print(payments)
	if err != nil {
		log.Print(err)
		return
	}

	err = s.HistoryToFiles(payments, "data", 1)
	if err != nil {
		log.Printf("HistoryToFiles() Error can't export to file, error = %v", err)
		return
	}
}

func newTestService() *testService {
	return &testService{Service: &wallet.Service{}}
}

func (s *testService) addAcoount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("cant register account %v = ", err)
	}

	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("cant deposit account %v = ", err)
	}

	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("cant make payment %v = ", err)
		}
	}

	return account, payments, nil
}
