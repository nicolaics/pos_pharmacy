package utils

import (
	"fmt"

	"github.com/nicolaics/pos_pharmacy/types"
)

func CheckStock(medData *types.Medicine, unit *types.Unit, additionalQty float64) error {
	var tempStock float64

	if medData.FirstUnitID == unit.ID {
		tempStock = additionalQty
	} else if medData.SecondUnitID == unit.ID {
		tempStock = (additionalQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == unit.ID {
		tempStock = (additionalQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	if tempStock > medData.Qty {
		return fmt.Errorf("buy requested is higher than the available stock")
	}
	
	return nil
}

func AddStock(medStore types.MedicineStore, medData *types.Medicine, unit *types.Unit, additionalQty float64, user *types.User) error {
	var updatedQty float64

	if medData.FirstUnitID == unit.ID {
		updatedQty = (additionalQty + medData.Qty)
	} else if medData.SecondUnitID == unit.ID {
		updatedQty = (additionalQty * medData.SecondUnitToFirstUnitRatio) + medData.Qty
	} else if medData.ThirdUnitID == unit.ID {
		updatedQty = (additionalQty * medData.ThirdUnitToFirstUnitRatio) + medData.Qty
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	err := medStore.UpdateMedicineStock(medData.ID, updatedQty, user)
	if err != nil {
		return err
	}

	return nil
}

func SubtractStock(medStore types.MedicineStore, medData *types.Medicine, unit *types.Unit, subtractionQty float64, user *types.User) error {
	var updatedQty float64

	if medData.FirstUnitID == unit.ID {
		updatedQty = (medData.Qty - subtractionQty)
	} else if medData.SecondUnitID == unit.ID {
		updatedQty = medData.Qty - (subtractionQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == unit.ID {
		updatedQty = medData.Qty - (subtractionQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	err := medStore.UpdateMedicineStock(medData.ID, updatedQty, user)
	if err != nil {
		return err
	}

	return nil
}