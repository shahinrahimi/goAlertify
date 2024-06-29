package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	GetUser(id string) (*User, error)
	GetUserByUserId(userId int64) (*User, error)
	GetUsers() ([]User, error)
	CreateUser(user User) error
	UpdateUser(id string, user User) error

	GetAlert(id string) (*Alert, error)
	GetAlerts() ([]Alert, error)
	GetAlertsByUserId(userId int64) ([]Alert, error)
	GetAlertByNumber(userId int64, number int32) (*Alert, error)
	GetAlertsByUserIdAndSymbol(userId int64, symbol string) ([]Alert, error)
	CreateAlert(alert *Alert) error
	UpdateAlert(alert *Alert) error
	DeleteAlert(id string) error

	DeleteUserAndAlerts(userId int64) error
}

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore() (*SqliteStore, error) {
	if err := CreateDirecryIfNotExist("database"); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", "./database/mydb.db")
	if err != nil {
		return nil, err
	}
	log.Println("DB Connected!")

	return &SqliteStore{
		db,
	}, nil
}

func (s *SqliteStore) Init() error {

	_, err := s.db.Exec(GetCreateUsersTable())
	if err != nil {
		return err
	}

	_, err = s.db.Exec(GetCreateAlertsTable())
	if err != nil {
		return err
	}
	return nil
}

// users crud
func (s *SqliteStore) GetUser(id string) (*User, error) {
	row := s.db.QueryRow(`SELECT id, user_id, username, firstname, lastname ,created_at FROM users WHERE id = ?`, id)
	var user User
	if err := row.Scan(&user.Id, &user.UserId, &user.Username, &user.Firstname, &user.Lastname, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
func (s *SqliteStore) GetUserByUserId(userId int64) (*User, error) {
	row := s.db.QueryRow(`SELECT id, user_id, username, firstname, lastname, password, created_at FROM users WHERE user_id = ?`, userId)
	var user User
	if err := row.Scan(&user.Id, &user.UserId, &user.Username, &user.Firstname, &user.Lastname, &user.Password, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
func (s *SqliteStore) GetUsers() ([]User, error) {
	rows, err := s.db.Query(`SELECT id, user_id, username, firstname, lastname ,created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.UserId, &u.Username, &u.Firstname, &u.Lastname, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
func (s *SqliteStore) CreateUser(user User) error {
	_, err := s.db.Exec(`INSERT INTO users (id, user_id, username, firstname, lastname ,password, created_at) VALUES (?,?,?,?,?,?,?)`, user.Id, user.UserId, user.Username, user.Firstname, user.Lastname, user.Password, user.CreatedAt)
	return err
}
func (s *SqliteStore) UpdateUser(id string, user User) error {
	_, err := s.db.Exec(`UPDATE users SET user_id = ?, username = ?, firstname = ?, lastname = ?, password = ?, cretated_at = ? WHERE id = ?`, user.UserId, user.Username, user.Firstname, user.Lastname, user.Password, user.CreatedAt, id)
	return err
}
func (s *SqliteStore) DeleteUser(id string) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// alert CRUD
func (s *SqliteStore) GetAlert(id string) (*Alert, error) {
	var alert Alert
	err := s.db.QueryRow(`SELECT id, user_id, number, symbol, description, target_price, start_price, active, created_at FROM alerts WHERE id = ?;`, id).
		Scan(&alert.Id, &alert.UserId, &alert.Number, &alert.Symbol, &alert.Description, &alert.TargetPrice, &alert.StartPrice, &alert.Active, &alert.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &alert, nil
}
func (s *SqliteStore) GetAlerts() ([]Alert, error) {
	rows, err := s.db.Query(`SELECT id, user_id, number, symbol, description, target_price, start_price, active, created_at FROM alerts;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.Id, &alert.UserId, &alert.Number, &alert.Symbol, &alert.Description, &alert.TargetPrice, &alert.StartPrice, &alert.Active, &alert.CreatedAt)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}
func (s *SqliteStore) GetAlertsByUserId(userId int64) ([]Alert, error) {
	rows, err := s.db.Query("SELECT id, user_id, number, symbol, description, target_price, start_price, active, created_at FROM alerts WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.Id, &alert.UserId, &alert.Number, &alert.Symbol, &alert.Description, &alert.TargetPrice, &alert.StartPrice, &alert.Active, &alert.CreatedAt)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}
func (s *SqliteStore) GetAlertByNumber(userId int64, number int32) (*Alert, error) {
	var alert Alert
	if err := s.db.QueryRow(`SELECT id, user_id, number, symbol, description, target_price, start_price, active, created_at FROM alerts WHERE user_id = ? AND number = ?;`, userId, number).Scan(&alert.Id, &alert.UserId, &alert.Number, &alert.Symbol, &alert.Description, &alert.TargetPrice, &alert.StartPrice, &alert.Active, &alert.CreatedAt); err != nil {
		return nil, err
	}
	return &alert, nil
}
func (s *SqliteStore) GetAlertsByUserIdAndSymbol(userId int64, symbol string) ([]Alert, error) {
	rows, err := s.db.Query("SELECT id, user_id, number, symbol, description, target_price, start_price, active, created_at FROM alerts WHERE user_id = ? AND symbol = ? COLLATE NOCASE", userId, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.Id, &alert.UserId, &alert.Number, &alert.Symbol, &alert.Description, &alert.TargetPrice, &alert.StartPrice, &alert.Active, &alert.CreatedAt)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}
func (s *SqliteStore) CreateAlert(alert *Alert) error {
	var maxNumber int32
	err := s.db.QueryRow("SELECT IFNULL(MAX(number), 0) FROM alerts WHERE user_id = ?", alert.UserId).Scan(&maxNumber)
	if err != nil {
		return err
	}
	alert.Number = maxNumber + 1

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO alerts (id, user_id, number, description, symbol, target_price, start_price, active, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(alert.Id, alert.UserId, alert.Number, alert.Description, alert.Symbol, alert.TargetPrice, alert.StartPrice, alert.Active, alert.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
func (s *SqliteStore) UpdateAlert(alert *Alert) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`UPDATE alerts SET description=?, symbol=?, target_price=?, start_price=?, active=? WHERE id=?;`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(alert.Description, alert.Symbol, alert.TargetPrice, alert.StartPrice, alert.Active, alert.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
func (s *SqliteStore) DeleteAlert(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM alerts WHERE id=?;`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
func (s *SqliteStore) DeleteUserAndAlerts(userId int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM alerts WHERE user_id = ?`, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM users WHERE user_id = ?`, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
