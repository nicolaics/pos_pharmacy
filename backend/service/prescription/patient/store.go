package patient

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

func (s *Store) GetPatientByName(name string, age int) (*types.Patient, error) {
	query := "SELECT * FROM patient WHERE name = ? AND age = ? ORDER BY name ASC"
	rows, err := s.db.Query(query, name, age)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patient := new(types.Patient)

	for rows.Next() {
		patient, err = scanRowIntoPatient(rows)

		if err != nil {
			return nil, err
		}
	}

	if patient.ID == 0 {
		return nil, fmt.Errorf("patient not found")
	}

	return patient, nil
}

func (s *Store) GetPatientsBySearchName(name string) ([]types.Patient, error) {
	query := "SELECT COUNT(*) FROM patient WHERE name = ?"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	patients := make([]types.Patient, 0)

	if count == 0 {
		query = "SELECT * FROM patient WHERE name LIKE ? ORDER BY name ASC"
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
			patient, err := scanRowIntoPatient(rows)

			if err != nil {
				return nil, err
			}

			patients = append(patients, *patient)
		}

		return patients, nil
	}

	query = "SELECT * FROM patient WHERE name = ? ORDER BY name ASC"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		patient, err := scanRowIntoPatient(rows)

		if err != nil {
			return nil, err
		}

		patients = append(patients, *patient)
	}

	return patients, nil
}

func (s *Store) GetPatientByID(id int) (*types.Patient, error) {
	query := "SELECT * FROM patient WHERE id = ? ORDER BY name ASC"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patient := new(types.Patient)

	for rows.Next() {
		patient, err = scanRowIntoPatient(rows)

		if err != nil {
			return nil, err
		}
	}

	if patient.ID == 0 {
		return nil, fmt.Errorf("patient not found")
	}

	return patient, nil
}

func (s *Store) CreatePatient(patient types.Patient) error {
	_, err := s.db.Exec("INSERT INTO patient (name, age) VALUES (?, ?)",
		patient.Name, patient.Age)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllPatients() ([]types.Patient, error) {
	rows, err := s.db.Query("SELECT * FROM patient ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patients := make([]types.Patient, 0)

	for rows.Next() {
		patient, err := scanRowIntoPatient(rows)

		if err != nil {
			return nil, err
		}

		patients = append(patients, *patient)
	}

	return patients, nil
}

func (s *Store) DeletePatient(patient *types.Patient, user *types.User) error {
	query := "DELETE FROM patient WHERE id = ?"
	_, err := s.db.Exec(query, patient.ID)
	if err != nil {
		return err
	}

	data, err := s.GetPatientByID(patient.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "patient", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) ModifyPatient(id int, patient types.RegisterPatientPayload, user *types.User) error {
	data, err := s.GetPatientByID(id)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "patient", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("UPDATE patient SET name = ?, age = ? WHERE id = ? ", patient.Name, patient.Age, id)

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPatient(rows *sql.Rows) (*types.Patient, error) {
	patient := new(types.Patient)

	err := rows.Scan(
		&patient.ID,
		&patient.Name,
		&patient.Age,
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	patient.CreatedAt = patient.CreatedAt.Local()

	return patient, nil
}
