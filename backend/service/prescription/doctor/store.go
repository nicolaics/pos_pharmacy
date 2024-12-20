package doctor

import (
	"database/sql"
	"fmt"

	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetDoctorByName(name string) (*types.Doctor, error) {
	query := "SELECT * FROM doctor WHERE name = ? ORDER BY name ASC"
	rows, err := s.db.Query(query, name)
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

func (s *Store) GetDoctorsBySearchName(name string) ([]types.Doctor, error) {
	query := "SELECT COUNT(*) FROM doctor WHERE name = ?"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	doctors := make([]types.Doctor, 0)

	if count == 0 {
		query = "SELECT * FROM doctor WHERE name LIKE ? ORDER BY name ASC"
		searchVal := "%"

		for _, val := range name {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			doctor, err := scanRowIntoDoctor(rows)

			if err != nil {
				return nil, err
			}

			doctors = append(doctors, *doctor)
		}

		return doctors, nil
	}

	query = "SELECT * FROM doctor WHERE name = ? ORDER BY name ASC"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		doctor, err := scanRowIntoDoctor(rows)

		if err != nil {
			return nil, err
		}

		doctors = append(doctors, *doctor)
	}

	return doctors, nil
}

func (s *Store) GetDoctorByID(id int) (*types.Doctor, error) {
	query := "SELECT * FROM doctor WHERE id = ? ORDER BY name ASC"
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
		doctor.Name)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllDoctors() ([]types.Doctor, error) {
	rows, err := s.db.Query("SELECT * FROM doctor ORDER BY name ASC")
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

func (s *Store) DeleteDoctor(doctor *types.Doctor, user *types.User) error {
	data, err := s.GetDoctorByID(doctor.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "doctor", user.Name, data.ID, data)
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

func (s *Store) ModifyDoctor(id int, newName string, user *types.User) error {
	data, err := s.GetDoctorByID(id)
	if err != nil {
		return err
	}

	err = logger.WriteLog("modify", "doctor", user.Name, data.ID, map[string]interface{}{"previous_data": data})
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("UPDATE doctor SET name = ? WHERE id = ? ", newName, id)
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
