package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nicolaics/pos_pharmacy/service/auth"
	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db          *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db,}
}

func (s *Store) GetUserByName(name string) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM user WHERE name = ? ", name)

	if err != nil {
		return nil, err
	}

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("customer not found")
	}

	return user, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM user WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) CreateUser(user types.User) error {
	_, err := s.db.Exec("INSERT INTO user (name, password, admin, phone_number) VALUES (?, ?, ?, ?)",
		user.Name, user.Password, user.Admin, user.PhoneNumber)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteUser(user *types.User) error {
	_, err := s.db.Exec("DELETE FROM user WHERE id = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllUsers() ([]types.User, error) {
	rows, err := s.db.Query("SELECT * FROM user")

	if err != nil {
		return nil, err
	}

	users := make([]types.User, 0)

	for rows.Next() {
		user, err := scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func (s *Store) UpdateLastLoggedIn(id int) error {
	_, err := s.db.Exec("UPDATE user SET last_logged_in = ? WHERE id = ? ",
		time.Now(), id)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyUser(id int, user types.User) error {
	columns := "name = ?, password = ?, admin = ?, phone_number = ?"

	_, err := s.db.Exec(fmt.Sprintf("UPDATE user SET %s WHERE id = ? ", columns),
		user.Name, user.Password, user.Admin, user.PhoneNumber, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveToken(userId int, tokenDetails *types.TokenDetails) error {
	accessTokenExp := time.Unix(tokenDetails.TokenExp, 0) //converting Unix to UTC(to Time object)
	refreshTokenExp := time.Unix(tokenDetails.RefreshTokenExp, 0)
	now := time.Now()

	err := s.redisClient.Set(context.Background(), tokenDetails.AccessUUID, strconv.Itoa(int(userId)), accessTokenExp.Sub(now)).Err()
	if err != nil {
		return err
	}

	err = s.redisClient.Set(context.Background(), tokenDetails.RefreshUUID, strconv.Itoa(int(userId)), refreshTokenExp.Sub(now)).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserIDFromRedis(accessDetails *types.AccessDetails, refreshDetails *types.RefreshDetails) (int, error) {
	var userIdStr string
	var err error

	if refreshDetails == nil {
		userIdStr, err = s.redisClient.Get(context.Background(), accessDetails.AccessUUID).Result()
	} else {
		userIdStr, err = s.redisClient.Get(context.Background(), refreshDetails.RefreshUUID).Result()
	}
	if err != nil {
		return -1, err
	}

	userId, _ := strconv.Atoi(userIdStr)

	return userId, nil
}

func (s *Store) DeleteToken(givenUuid string) (int, error) {
	deleted, err := s.redisClient.Del(context.Background(), givenUuid).Result()
	if err != nil {
		return -1, err
	}

	return int(deleted), nil
}

func (s *Store) FindUserID(db *sql.DB, userName string) (int, error) {
	rows, err := db.Query("SELECT * FROM user WHERE name = ? ", userName)

	if err != nil {
		return -1, err
	}

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return -1, err
		}
	}

	return user.ID, nil
}

// TODO: change into from sql, not redis
// TODO: using SELECT COUNT(*) WHERE name = ? AND
// TODO: token = ? AND TIMESTAMPDIFF(HOUR, created_at, NOW()) <= ?
// TODO: check using the count, if count 0, means not valid
// TODO: delete the token
func (s *Store) ValidateUserToken(w http.ResponseWriter, r *http.Request, needAdmin bool) (*types.User, error) {
	accessDetails, err := auth.ExtractTokenFromClient(r)
	if err != nil {
		return nil, err
	}

	userID, err := s.GetUserIDFromRedis(accessDetails, nil)
	if err != nil {
		return nil, fmt.Errorf("renew your token")
	}

	// check if user exist
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// if the account must be admin
	if needAdmin {
		if !user.Admin {
			return nil, fmt.Errorf("unauthorized! not admin")
		}
	}

	return user, nil
}


func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.Admin,
		&user.PhoneNumber,
		&user.LastLoggedIn,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.LastLoggedIn = user.LastLoggedIn.Local()
	user.CreatedAt = user.CreatedAt.Local()

	return user, nil
}
