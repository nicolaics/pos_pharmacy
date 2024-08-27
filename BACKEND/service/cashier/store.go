package cashier

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewStore(db *sql.DB, redisClient *redis.Client) *Store {
	return &Store{db: db, redisClient: redisClient}
}

func (s *Store) GetCashierByName(name string) (*types.Cashier, error) {
	rows, err := s.db.Query("SELECT * FROM cashier WHERE name = ? ", name)

	if err != nil {
		return nil, err
	}

	cashier := new(types.Cashier)

	for rows.Next() {
		cashier, err = scanRowIntoCashier(rows)

		if err != nil {
			return nil, err
		}
	}

	if cashier.ID == 0 {
		return nil, fmt.Errorf("customer not found")
	}

	return cashier, nil
}

func (s *Store) GetCashierByID(id int) (*types.Cashier, error) {
	rows, err := s.db.Query("SELECT * FROM cashier WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	cashier := new(types.Cashier)

	for rows.Next() {
		cashier, err = scanRowIntoCashier(rows)

		if err != nil {
			return nil, err
		}
	}

	if cashier.ID == 0 {
		return nil, fmt.Errorf("cashier not found")
	}

	return cashier, nil
}

func (s *Store) CreateCashier(cashier types.Cashier) error {
	_, err := s.db.Exec("INSERT INTO cashier (name, password, admin) VALUES (?, ?, ?)",
		cashier.Name, cashier.Password, cashier.Admin)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteCashier(cashier *types.Cashier) error {
	_, err := s.db.Exec("DELETE FROM cashier WHERE id = ?", cashier.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllCashiers() ([]types.Cashier, error) {
	rows, err := s.db.Query("SELECT * FROM cashier")

	if err != nil {
		return nil, err
	}

	cashiers := make([]types.Cashier, 0)

	for rows.Next() {
		cashier, err := scanRowIntoCashier(rows)

		if err != nil {
			return nil, err
		}

		cashiers = append(cashiers, *cashier)
	}

	return cashiers, nil
}

func (s *Store) UpdateLastLoggedIn(cashier *types.Cashier) error {
	_, err := s.db.Exec("UPDATE cashier SET last_logged_in = ? WHERE id = ? ",
		time.Now(), cashier.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateAdmin(cashier *types.Cashier) error {
	_, err := s.db.Exec("UPDATE cashier SET admin = ? WHERE id = ? ",
		true, cashier.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveAuth(cashierId int, tokenDetails *types.TokenDetails) error {
	accessTokenExp := time.Unix(tokenDetails.AccessTokenExp, 0) //converting Unix to UTC(to Time object)
	refreshTokenExp := time.Unix(tokenDetails.RefreshTokenExp, 0)
	now := time.Now()

	err := s.redisClient.Set(context.Background(), tokenDetails.AccessUUID, strconv.Itoa(int(cashierId)), accessTokenExp.Sub(now)).Err()
	if err != nil {
		return err
	}

	err = s.redisClient.Set(context.Background(), tokenDetails.RefreshUUID, strconv.Itoa(int(cashierId)), refreshTokenExp.Sub(now)).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAuthentification(authDetails *types.AccessDetails) (int, error) {
	cashierIdStr, err := s.redisClient.Get(context.Background(), authDetails.AccessUUID).Result()
	if err != nil {
		return -1, err
	}

	cashierId, _ := strconv.Atoi(cashierIdStr)

	return cashierId, nil
}

func (s *Store) DeleteAuth(givenUuid string) (int, error) {
	deleted, err := s.redisClient.Del(context.Background(), givenUuid).Result()
	if err != nil {
		return -1, err
	}
	
	return int(deleted), nil
}

func (s *Store) FindCashierID(db *sql.DB, cashierName string) (int, error) {
	rows, err := db.Query("SELECT * FROM cashier WHERE name = ? ", cashierName)

	if err != nil {
		return -1, err
	}

	cashier := new(types.Cashier)

	for rows.Next() {
		cashier, err = scanRowIntoCashier(rows)

		if err != nil {
			return -1, err
		}
	}

	return cashier.ID, nil
}

func scanRowIntoCashier(rows *sql.Rows) (*types.Cashier, error) {
	cashier := new(types.Cashier)

	err := rows.Scan(
		&cashier.ID,
		&cashier.Name,
		&cashier.Password,
		&cashier.Admin,
		&cashier.CreatedAt,
		&cashier.LastLoggedIn,
	)

	if err != nil {
		return nil, err
	}

	return cashier, nil
}
