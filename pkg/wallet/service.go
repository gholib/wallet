package wallet

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gholib/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	return s.Pay(payment.AccountID, payment.Amount, payment.Category)
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Name:      name,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	return s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
}
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Print(err)
		}

	}()
	str := ""
	for _, account := range s.accounts {
		str += strconv.Itoa(int(account.ID)) + ";"
		str += string(account.Phone) + ";"
		str += strconv.Itoa(int(account.Balance)) + "|"
	}

	_, err = file.Write([]byte(str))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Service) ImportFromFile(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return err
	}

	data := string(byteData)

	splitSlice := strings.Split(data, "|")
	for _, split := range splitSlice {
		if split != "" {
			datas := strings.Split(split, ";")

			id, err := strconv.Atoi(datas[0])
			if err != nil {
				log.Println(err)
				return err
			}

			balance, err := strconv.Atoi(datas[2])
			if err != nil {
				log.Println(err)
				return err
			}

			newAccount := &types.Account{
				ID:      int64(id),
				Phone:   types.Phone(datas[1]),
				Balance: types.Money(balance),
			}

			s.accounts = append(s.accounts, newAccount)
		}
	}

	return nil

}

func (s *Service) Export(dir string) error {
	if len(s.accounts) > 0 {
		err := s.ExportAccounts(dir)
		if err != nil {
			return err
		}
	}

	if len(s.payments) > 0 {
		err := s.ExportPayments(dir)
		if err != nil {
			return err
		}
	}

	if len(s.favorites) > 0 {
		err := s.ExportFavorites(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ExportAccounts(dir string) error {
	str := ""
	for _, account := range s.accounts {
		str += strconv.Itoa(int(account.ID)) + ";"
		str += string(account.Phone) + ";"
		str += strconv.Itoa(int(account.Balance)) + ";"
		str += string('\n')
	}

	err := WriteToFile(dir+"/accounts.dump", str)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s *Service) ExportPayments(dir string) error {
	str := ""
	for _, payment := range s.payments {
		str += payment.ID + ";"
		str += strconv.Itoa(int(payment.AccountID)) + ";"
		str += strconv.Itoa(int(payment.Amount)) + ";"
		str += string(payment.Category) + ";"
		str += string(payment.Status) + ";"
		str += string('\n')
	}

	err := WriteToFile(dir+"/payments.dump", str)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s *Service) ExportFavorites(dir string) error {
	str := ""
	for _, favorite := range s.favorites {
		str += favorite.ID + ";"
		str += strconv.Itoa(int(favorite.AccountID)) + ";"
		str += strconv.Itoa(int(favorite.Amount)) + ";"
		str += string(favorite.Name) + ";"
		str += string(favorite.Category) + ";"
		str += string('\n')
	}

	err := WriteToFile(dir+"/favorites.dump", str)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s *Service) Import(dir string) error {
	err := s.ImportAccounts(dir + "/accounts.dump")
	log.Println(s.accounts)
	if err != nil {
		return err
	}

	err = s.ImportPayments(dir + "/payments.dump")
	log.Println(s.payments)
	if err != nil {
		return err
	}

	err = s.ImportFavorites(dir + "/favorites.dump")
	log.Println(s.favorites)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImportAccounts(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return err
	}
	datas := string(byteData)
	splits := strings.Split(datas, "\n")

	for _, split := range splits {
		if len(split) == 0 {
			break
		}

		data := strings.Split(split, ";")

		id, err := strconv.Atoi(data[0])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}

		phone := types.Phone(data[1])

		balance, err := strconv.Atoi(data[2])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}

		account, err := s.FindAccountByID(int64(id))
		if err != nil {
			acc, err := s.RegisterAccount(phone)
			if err != nil {
				log.Println("err from register account")
				return err
			}

			acc.Balance = types.Money(balance)
		} else {
			account.Phone = phone
			account.Balance = types.Money(balance)
		}
	}

	return nil
}

func (s *Service) ImportPayments(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return err
	}
	datas := string(byteData)
	splits := strings.Split(datas, "\n")

	for _, split := range splits {
		if len(split) == 0 {
			break
		}

		data := strings.Split(split, ";")
		id := data[0]

		accountID, err := strconv.Atoi(data[1])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}

		amount, err := strconv.Atoi(data[2])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}

		category := types.PaymentCategory(data[3])

		status := types.PaymentStatus(data[4])

		payment, err := s.FindPaymentByID(id)
		if err != nil {
			newPayment := &types.Payment{
				ID:        id,
				AccountID: int64(accountID),
				Amount:    types.Money(amount),
				Category:  types.PaymentCategory(category),
				Status:    types.PaymentStatus(status),
			}

			s.payments = append(s.payments, newPayment)
		} else {
			payment.AccountID = int64(accountID)
			payment.Amount = types.Money(amount)
			payment.Category = category
			payment.Status = status
		}
	}

	return nil
}

func (s *Service) ImportFavorites(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return err
	}
	datas := string(byteData)
	splits := strings.Split(datas, "\n")

	for _, split := range splits {
		if len(split) == 0 {
			break
		}

		data := strings.Split(split, ";")
		id := data[0]

		accountID, err := strconv.Atoi(data[1])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}

		amount, err := strconv.Atoi(data[2])
		if err != nil {
			log.Println("can't parse str to int")
			return err
		}
		name := data[3]
		category := types.PaymentCategory(data[4])

		favorite, err := s.FindFavoriteByID(id)
		if err != nil {
			newFavorite := &types.Favorite{
				ID:        id,
				AccountID: int64(accountID),
				Name:      name,
				Amount:    types.Money(amount),
				Category:  types.PaymentCategory(category),
			}

			s.favorites = append(s.favorites, newFavorite)
		} else {
			favorite.AccountID = int64(accountID)
			favorite.Name = name
			favorite.Amount = types.Money(amount)
			favorite.Category = category
		}
	}
	return nil
}

//WriteToFile
func WriteToFile(path string, data string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
			return
		}
	}()
	// он возвращает кол-во байтов

	_, err = file.WriteString(data)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
