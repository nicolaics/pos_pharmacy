package doctor

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetDoctorByName(name string) (*types.Doctor, error) {
	query := "SELECT * FROM doctor WHERE name = ?"
	rows, err := s.db.Query(query, strings.ToUpper(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctor := new(types.Doctor)

	for rows.Next() {
		doctor, err = scanRowIntoDoctor(rows)

		if err != nil {
			return nil, err
		}
	}

	if doctor.ID == 0 {
		return nil, fmt.Errorf("doctor not found")
	}

	return doctor, nil
}

func (s *Store) GetDoctorByID(id int) (*types.Doctor, error) {
	query := "SELECT * FROM doctor WHERE id = ?"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctor := new(types.Doctor)

	for rows.Next() {
		doctor, err = scanRowIntoDoctor(rows)

		if err != nil {
			return nil, err
		}
	}

	if doctor.ID == 0 {
		return nil, fmt.Errorf("doctor not found")
	}

	return doctor, nil
}

func (s *Store) CreateDoctor(doctor types.Doctor) error {
	_, err := s.db.Exec("INSERT INTO doctor (name) VALUES (?)",
						strings.ToUpper(doctor.Name))

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllDoctors() ([]types.Doctor, error) {
	rows, err := s.db.Query("SELECT * FROM doctor")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctors := make([]types.Doctor, 0)

	for rows.Next() {
		doctor, err := scanRowIntoDoctor(rows)

		if err != nil {
			return nil, err
		}

		doctors = append(doctors, *doctor)
	}

	return doctors, nil
}

func (s *Store) DeleteDoctor(doctor *types.Doctor, uid int) error {
	data, err := s.GetDoctorByID(doctor.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "doctor", uid, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := "DELETE FROM doctor WHERE id = ?"
	_, err = s.db.Exec(query, doctor.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyDoctor(id int, newName string, uid int) error {
	data, err := s.GetDoctorByID(id)
	if err != nil {
		return err
	}

	err = logger.WriteLog("modify", "doctor", uid, data.ID, map[string]interface{}{"previous_data": data})
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("UPDATE doctor SET name = ? WHERE id = ? ", strings.ToUpper(newName), id)

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoDoctor(rows *sql.Rows) (*types.Doctor, error) {
	doctor := new(types.Doctor)

	err := rows.Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	doctor.CreatedAt = doctor.CreatedAt.Local()

	return doctor, nil
}
