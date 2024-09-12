import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../../App";

const ModifySupplierPage: React.FC = () => {
  const navigate = useNavigate();
  const state = useLocation().state;

  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [companyPhoneNumber, setCompanyPhoneNumber] = useState("");
  const [contactPersonName, setContactPersonName] = useState("");
  const [contactPersonNumber, setContactPersonNumber] = useState("");
  const [terms, setTerms] = useState("");
  const [vendorIsTaxable, setVendorIsTaxable] = useState(true);

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");
  const [showIdField, setShowIdField] = useState(false);

  var heading = "";
  if (state) {
    heading = "Modify";
  } else {
    heading = "Add";
  }

  /*
  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state) {
      setOkBtnLabel("Modify");
      setShowIdField(true);

      // TODO: add query for searching suppliers
      const supplierURL = `http://${BACKEND_BASE_URL}/supplier?id=`; // Set the URL or handle this logic
      fetch(supplierURL, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify supplier data");
            }

            console.log(data["data"]);

            setId(data["data"].id);
            setName(data["data"].name);
          })
        )
        .catch((error) => {
          console.error("Error load selected supplier:", error);
          alert("Error load selected supplier");
        });
    } else {
      setOkBtnLabel("Add");
      setShowIdField(false);
    }
  }, [state]); // Dependency array ensures this effect only runs when reqType changes
  */

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handleAddressChange = (event: any) => {
    setAddress(event.target.value);
  };

  const handleCompanyPhoneNumberChange = (event: any) => {
    setCompanyPhoneNumber(event.target.value);
  };

  const handleContactPersonNameChange = (event: any) => {
    setContactPersonName(event.target.value);
  };

  const handleContactPersonNumberChange = (event: any) => {
    setContactPersonNumber(event.target.value);
  };

  const handleTermsChange = (event: any) => {
    setTerms(event.target.value);
  };

  const handleVendorIsTaxableChange = (event: any) => {
    if (event.target.value == "yes") {
      setVendorIsTaxable(true);
    } else {
      setVendorIsTaxable(false);
    }
  };

  const handleSendRequest = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");

    if (state) {
      const url = `http://${BACKEND_BASE_URL}/supplier`;

      fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          name: name,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify supplier data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify supplier:", error);
          alert("Error modify supplier");
        });
    } else {
      const url = `http://${BACKEND_BASE_URL}/supplier`;

      fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: Number(id),
          newData: {
            name: name,
          },
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Invalid c redentials or network issue");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error adding new supplier:", error);
          alert("Error adding new supplier");
        });
    }

    // Reset the state
    navigate("/supplier");
  };

  const handleCancel = (e: any) => {
    navigate("/supplier");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();

    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/supplier`;

    fetch(url, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        id: Number(id),
        name: name,
      }),
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Unable to delete supplier data");
          }

          console.log(data);
          navigate("/supplier");
        })
      )
      .catch((error) => {
        console.error("Error delete supplier:", error);
        alert("Error delete supplier");
      });
  };

  return (
    <div className="modify-supplier-page">
      <h1>{heading} Supplier</h1>

      <div className="supplier-data-container">
        {showIdField && (
        <div className="supplier-data-form-group">
          <label htmlFor="id">ID:</label>
          <input type="text" id="modify-supplier-id" value={id} readOnly />
        </div>
        )}

        <div className="supplier-data-form-group">
          <label htmlFor="name">Name:</label>
          <input
            type="text"
            id="modify-supplier-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="address">Address:</label>
          <textarea
            id="modify-supplier-address"
            value={address}
            onChange={handleAddressChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="companyPhoneNumber">Company Phone Number:</label>
          <input
            type="text"
            id="modify-supplier-companyPhoneNumber"
            value={companyPhoneNumber}
            onChange={handleCompanyPhoneNumberChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="contactPersonName">Contact Person Name:</label>
          <input
            type="text"
            id="modify-supplier-contactPersonName"
            value={contactPersonName}
            onChange={handleContactPersonNameChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="contactPersonNumber">Contact Person Number:</label>
          <input
            type="text"
            id="modify-supplier-contactPersonNumber"
            value={contactPersonNumber}
            onChange={handleContactPersonNumberChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="terms">Terms:</label>
          <input
            type="text"
            id="modify-supplier-terms"
            value={terms}
            onChange={handleTermsChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label>Vendor is Taxable:</label>
          <div className="supplier-data-radio-grp">
            <input
              type="radio"
              id="modify-supplier-radio-yes"
              checked={vendorIsTaxable === true}
              name="vendorIsTaxable"
              value={"yes"}
              onChange={handleVendorIsTaxableChange}
            />
            <label htmlFor="radio-yes">Yes</label>
            <input
              type="radio"
              checked={vendorIsTaxable === false}
              id="modify-supplier-radio-no"
              name="vendorIsTaxable"
              value={"no"}
              onChange={handleVendorIsTaxableChange}
            />
            <label htmlFor="radio-no">No</label>
          </div>
        </div>
      </div>

      <div className="modify-supplier-buttons">
        <div className="modify-supplier-btns-grp">
          <button
            type="button"
            className="modify-supplier-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>

          <button
            type="button"
            className="modify-supplier-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="button"
            className="modify-supplier-delete-btn"
            onClick={handleDelete}
          >
            Delete Supplier
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifySupplierPage;
