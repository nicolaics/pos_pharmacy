package patient

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetPatientByName(name string) (*types.Patient, error) {
	query := "SELECT * FROM patient WHERE name = ?"
	rows, err := s.db.Query(query, strings.ToUpper(name))
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

func (s *Store) GetPatientByID(id int) (*types.Patient, error) {
	query := "SELECT * FROM patient WHERE id = ?"
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
	_, err := s.db.Exec("INSERT INTO patient (name) VALUES (?)",
						strings.ToUpper(patient.Name))

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllPatients() ([]types.Patient, error) {
	rows, err := s.db.Query("SELECT * FROM patient")
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

func (s *Store) DeletePatient(patient *types.Patient) error {
	query := "DELETE FROM patient WHERE id = ?"
	_, err := s.db.Exec(query, patient.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPatient(id int, newName string) error {
	_, err := s.db.Exec("UPDATE patient SET name = ? WHERE id = ? ", strings.ToUpper(newName), id)

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
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	patient.CreatedAt = patient.CreatedAt.Local()

	return patient, nil
}
