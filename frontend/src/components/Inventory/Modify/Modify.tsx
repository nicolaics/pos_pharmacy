import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import AdminPasswordPopup from "../../AdminPasswordPopup/AdminPasswordPopup";
import { BACKEND_BASE_URL } from "../../../App";
import FormatDateTime from "../../../DateTimeFormatter";

// use window.location.href if the files have been moved to the server

const ModifyInventoryPage: React.FC = () => {
  const navigate = useNavigate();
  const state = useLocation().state;

  const [id, setId] = useState("");
  const [barcode, setBarcode] = useState("");
  const [name, setName] = useState("");
  const [stock, setStock] = useState(0.0);

  const [firstUnit, setFirstUnit] = useState("");
  const [firstUnitDiscount, setFirstUnitDiscount] = useState(0.0);
  const [firstUnitPrice, setFirstUnitPrice] = useState(0.0);

  const [secondUnit, setSecondUnit] = useState("");
  const [secondUnitDiscount, setSecondUnitDiscount] = useState(0.0);
  const [secondUnitPrice, setSecondUnitPrice] = useState(0.0);

  const [thirdUnit, setThirdUnit] = useState("");
  const [thirdUnitDiscount, setThirdUnitDiscount] = useState(0.0);
  const [thirdUnitPrice, setThirdUnitPrice] = useState(0.0);

  const [description, setDescription] = useState("");
  const [createdAt, setCreatedAt] = useState("");
  const [lastModified, setLastModified] = useState("");
  const [lastModifiedBy, setLastModifiedBy] = useState("");

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");
  const [showIdField, setShowIdField] = useState(false);
  const [showDeleteButton, setShowDeleteButton] = useState(false);

  const [reqType, setReqType] = useState(state.reqType);

  var heading = "";
  if (reqType == "add") {
    heading = "Add";
  } else {
    heading = "Modify";
  }

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state.reqType === "modify") {
      setOkBtnLabel("Modify");
      setShowIdField(true);

      const currentInventoryURL = `http://${BACKEND_BASE_URL}/medicine/detail`;
      fetch(currentInventoryURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: state.id,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to get medicine data");
            }

            console.log(data);
            
            // TODO: get discount from decimal number to percentage, process here
            setId(data.id);
            setBarcode(data.barcode);
            setName(data.name);
            setStock(data.qty);
            setFirstUnit(data.firstUnitName);
            setFirstUnitDiscount(data.firstDiscount);
            setFirstUnitPrice(data.firstPrice);
            setSecondUnit(data.secondUnitName);
            setSecondUnitDiscount(data.secondDiscount);
            setSecondUnitPrice(data.secondPrice);
            setThirdUnit(data.thirdUnitName);
            setThirdUnitDiscount(data.thirdDiscount);
            setThirdUnitPrice(data.thirdPrice);
            setDescription(data.description);

            const createdAtStr = FormatDateTime(new Date(data.createdAt));
            setCreatedAt(createdAtStr);

            const lastModifiedStr = FormatDateTime(new Date(data.lastModified));
            setLastModified(lastModifiedStr);

            setLastModifiedBy(data.lastModifiedByUserName);
          })
        )
        .catch((error) => {
          console.error("Error load current inventory:", error);
          alert("Error load current inventory");
        });
    } else if (state.reqType === "add") {
      setOkBtnLabel("Add");
      setShowDeleteButton(false);
      setShowIdField(false);
    }
  }, [state.reqType]); // Dependency array ensures this effect only runs when reqType changes

  const handleBarcodeChange = (event: any) => {
    setBarcode(event.target.value);
  };

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handleStockChange = (event: any) => {
    setStock(event.target.value);
  };

  const handleFirstUnitChange = (event: any) => {
    setFirstUnit(event.target.value);
  };

  const handleFirstUnitDiscountChange = (event: any) => {
    setFirstUnitDiscount(event.target.value);
  };

  const handleFirstUnitPriceChange = (event: any) => {
    setFirstUnitPrice(event.target.value);
  };

  const handleSecondUnitChange = (event: any) => {
    setSecondUnit(event.target.value);
  };

  const handleSecondUnitDiscountChange = (event: any) => {
    setSecondUnitDiscount(event.target.value);
  };

  const handleSecondUnitPriceChange = (event: any) => {
    setSecondUnitPrice(event.target.value);
  };

  const handleThirdUnitChange = (event: any) => {
    setThirdUnit(event.target.value);
  };

  const handleThirdUnitDiscountChange = (event: any) => {
    setThirdUnitDiscount(event.target.value);
  };

  const handleThirdUnitPriceChange = (event: any) => {
    setThirdUnitPrice(event.target.value);
  };

  const handleDescriptionChange = (event: any) => {
    setDescription(event.target.value);
  };

  const handleSendRequest = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");

    const url = `http://${BACKEND_BASE_URL}/medicine`;

    // TODO: send discount with decimal number, process here
    const firstTotalPrice = firstUnitPrice - firstUnitDiscount;
    const secondTotalPrice = secondUnitPrice - secondUnitDiscount;
    const thirdTotalPrice = thirdUnitPrice - thirdUnitDiscount;

    console.log(stock);

    if (state.reqType === "add") {
      fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          barcode: barcode,
          name: name,
          qty: Number(stock),
          firstUnit: firstUnit,
          firstSubtotal: Number(firstUnitPrice),
          firstDiscount: Number(firstUnitDiscount),
          firstPrice: Number(firstTotalPrice),
          secondSubtotal: Number(secondUnitPrice),
          secondDiscount: Number(secondUnitDiscount),
          secondPrice: Number(secondTotalPrice),
          thirdSubtotal: Number(thirdUnitPrice),
          thirdDiscount: Number(thirdUnitDiscount),
          thirdPrice: Number(thirdTotalPrice),
          description: description,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Invalid credentials or network issue");
            }

            alert(data);
          })
        )
        .catch((error) => {
          console.error("Error adding new inventory:", error);
          alert("Error adding new inventory");
        });
    }

    // Reset the state
    navigate("/inventory");
  };

  const handleCancel = (e: any) => {
    navigate("/inventory");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();
  };

  return (
    <div className="modify-inventory-page">
      <h1>{heading} Inventory</h1>

      <div className="inventory-data-container">
        {showIdField && (
          <div className="inventory-data-form-group">
            <label htmlFor="modify-inventory-id">ID:</label>
            <input type="text" id="modify-inventory-id" value={id} readOnly />
          </div>
        )}

        <div className="inventory-data-form-group">
          <label htmlFor="modify-inventory-barcode">Barcode:</label>
          <input
            type="text"
            id="modify-inventory-barcode"
            value={barcode}
            onChange={handleBarcodeChange}
          />
        </div>

        <div className="inventory-data-form-group">
          <label htmlFor="modify-inventory-name">Name:</label>
          <input
            type="text"
            id="modify-inventory-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>

        <div className="inventory-data-form-group">
          <label htmlFor="modify-inventory-stock">Stock:</label>
          <input
            type="text"
            id="modify-inventory-stock"
            value={stock}
            onChange={handleStockChange}
          />
        </div>

        <div className="inventory-data-form-group">
          <div className="unit-form-group">
            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-one-name">Unit 1:</label>
              <input
                type="text"
                id="modify-inventory-unit-one-name"
                value={firstUnit}
                onChange={handleFirstUnitChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-one-discount">
                Discount:
              </label>
              <input
                type="text"
                id="modify-inventory-unit-one-discount"
                value={firstUnitDiscount}
                onChange={handleFirstUnitDiscountChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-one-price">Price:</label>
              <input
                type="text"
                id="modify-inventory-unit-one-price"
                value={firstUnitPrice}
                onChange={handleFirstUnitPriceChange}
              />
            </div>
          </div>

          <div className="unit-form-group">
            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-two-name">Unit 2:</label>
              <input
                type="text"
                id="modify-inventory-unit-two-name"
                value={secondUnit}
                onChange={handleSecondUnitChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-two-discount">
                Discount:
              </label>
              <input
                type="text"
                id="modify-inventory-unit-two-discount"
                value={secondUnitDiscount}
                onChange={handleSecondUnitDiscountChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-two-price">Price:</label>
              <input
                type="text"
                id="modify-inventory-unit-two-price"
                value={secondUnitPrice}
                onChange={handleSecondUnitPriceChange}
              />
            </div>
          </div>

          <div className="unit-form-group">
            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-three-name">Unit 3:</label>
              <input
                type="text"
                id="modify-inventory-unit-three-name"
                value={thirdUnit}
                onChange={handleThirdUnitChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-three-discount">
                Discount:
              </label>
              <input
                type="text"
                id="modify-inventory-unit-three-discount"
                value={thirdUnitDiscount}
                onChange={handleThirdUnitDiscountChange}
              />
            </div>

            <div className="unit-grid-item">
              <label htmlFor="modify-inventory-unit-three-price">Price:</label>
              <input
                type="text"
                id="modify-inventory-unit-three-price"
                value={thirdUnitPrice}
                onChange={handleThirdUnitPriceChange}
              />
            </div>
          </div>
        </div>

        <div className="inventory-data-form-group">
          <label htmlFor="modify-inventory-description">Description:</label>
          <textarea
            id="modify-inventory-description"
            value={description}
            onChange={handleDescriptionChange}
          />
        </div>

        {showIdField && (
          <>
            <div className="inventory-data-form-group">
              <label htmlFor="modify-inventory-created-at">Created At:</label>
              <input
                type="text"
                id="modify-inventory-created-at"
                value={createdAt}
                readOnly={true}
              />
            </div>

            <div className="inventory-data-form-group">
              <label htmlFor="modify-inventory-last-modified">
                Last Modified:
              </label>
              <input
                type="text"
                id="modify-inventory-last-modified"
                value={lastModified}
                readOnly={true}
              />
            </div>

            <div className="inventory-data-form-group">
              <label htmlFor="modify-inventory-last-modified-by">
                Last Modified By:
              </label>
              <input
                type="text"
                id="modify-inventory-last-modified-by"
                value={lastModifiedBy}
                readOnly={true}
              />
            </div>
          </>
        )}
      </div>

      <div className="modify-inventory-buttons">
        <div className="modify-inventory-btns-grp">
          {showDeleteButton && (
            <button
              type="button"
              className="modify-inventory-delete-btn"
              onClick={handleDelete}
            >
              Delete Inventory
            </button>
          )}

          <button
            type="button"
            className="modify-inventory-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="submit"
            className="modify-inventory-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifyInventoryPage;
