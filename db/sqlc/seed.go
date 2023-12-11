package postgresdb

import (
	"context"
	"fmt"

	"github.com/mdmn07C5/bank/util"
)

func CreateSeedUser(username, password, fullname, email string) (CreateUserParams, error) {
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return CreateUserParams{}, err
	}
	user := CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		FullName:       fullname,
		Email:          email,
	}
	return user, nil
}

func CreateSeedAccounts(username string, currencies []string) []CreateAccountParams {
	createAccountParams := make([]CreateAccountParams, 3, len(currencies))
	for i := range currencies {
		arg := CreateAccountParams{
			Owner:    username,
			Balance:  100,
			Currency: currencies[i],
		}
		createAccountParams[i] = arg
	}
	return createAccountParams
}

func Seed(store Store) error {
	peepoUserParams, err := CreateSeedUser("peepo", "password123", "Apu Apustaja", "apu@apustaja.com")
	if err != nil {
		return fmt.Errorf("something went wrong while making %s", "peepo")
	}
	gondolaUserParams, err := CreateSeedUser("gondola", "password123", "Spurdo Spärde", "spurdo@sparde.com")
	if err != nil {
		return fmt.Errorf("something went wrong while making %s", "gondola")
	}

	peepo, err := store.CreateUser(context.Background(), peepoUserParams)

	if err != nil {
		return fmt.Errorf("something went wrong while making %s", peepoUserParams.Username)
	}

	gondola, err := store.CreateUser(context.Background(), gondolaUserParams)
	if err != nil {
		return fmt.Errorf("something went wrong while making %s", gondola.Username)
	}

	peepoAccountsCurrencies := []string{"USD", "GBP", "EUR"}
	gondolaAccountCurrencies := []string{"USD", "EUR", "CAD"}

	peepoAccounts := CreateSeedAccounts(peepo.Username, peepoAccountsCurrencies)
	gondolaAccounts := CreateSeedAccounts(gondola.Username, gondolaAccountCurrencies)

	for i := range peepoAccounts {
		_, err := store.CreateAccount(context.Background(), peepoAccounts[i])
		if err != nil {
			return fmt.Errorf("something went wrong while making accounts for %s", peepo.Username)
		}
		_, err = store.CreateAccount(context.Background(), gondolaAccounts[i])
		if err != nil {
			return fmt.Errorf("something went wrong while making accounts for %s", gondola.Username)
		}
	}

	return nil
}
