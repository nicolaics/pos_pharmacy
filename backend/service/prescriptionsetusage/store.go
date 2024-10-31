package prescriptionsetusage

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

func (s *Store) GetPrescriptionSetUsageByName(prescriptionSetUsageName string) (*types.PrescriptionSetUsage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE name = ? ", strings.ToLower(prescriptionSetUsageName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptionSetUsage := new(types.PrescriptionSetUsage)

	for rows.Next() {
		prescriptionSetUsage, err = scanRowIntoPrescriptionSetUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	if prescriptionSetUsage.ID == 0 {
		return nil, fmt.Errorf("prescription set usage not found")
	}

	return prescriptionSetUsage, nil
}

func (s *Store) GetPrescriptionSetUsageByID(id int) (*types.PrescriptionSetUsage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptionSetUsage := new(types.PrescriptionSetUsage)

	for rows.Next() {
		prescriptionSetUsage, err = scanRowIntoPrescriptionSetUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	return prescriptionSetUsage, nil
}


func (s *Store) CreatePrescriptionSetUsage(prescriptionSetUsageName string) error {
	_, err := s.db.Exec("INSERT INTO prescription_set_usage (name) VALUES (?)", strings.ToLower(prescriptionSetUsageName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPrescriptionSetUsage(rows *sql.Rows) (*types.PrescriptionSetUsage, error) {
	prescriptionSetUsage := new(types.PrescriptionSetUsage)

	err := rows.Scan(
		&prescriptionSetUsage.ID,
		&prescriptionSetUsage.Name,
		&prescriptionSetUsage.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	prescriptionSetUsage.CreatedAt = prescriptionSetUsage.CreatedAt.Local()

	return prescriptionSetUsage, nil
}
