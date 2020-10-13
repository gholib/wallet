package main

import (
	"log"

	"github.com/gholib/wallet/pkg/wallet"
)

func main() {
	s := &wallet.Service{}

	_, err := s.RegisterAccount("+992880806776")
	if err != nil {
		log.Println(err)
		return
	}

	err = s.ExportToFile("../data/accounts.txt")
	if err != nil {
		log.Println(err)
		return
	}
}
