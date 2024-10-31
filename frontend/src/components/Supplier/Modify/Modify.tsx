import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../../App";
import FormatDateTime from "../../../DateTimeFormatter";

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
  const [lastModified, setLastModified] = useState("");
  const [lastModifiedByUserName, setLastModifiedByUserName] = useState("");
  const [createdAt, setCreatedAt] = useState("");

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");
  const [showIdField, setShowIdField] = useState(false);

  var heading = "";
  if (state) {
    heading = "Modify";
  } else {
    heading = "Add";
  }

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state) {
      setOkBtnLabel("Modify");
      setShowIdField(true);

      const supplierURL = `http://${BACKEND_BASE_URL}/supplier/id/${state.id}`; // Set the URL or handle this logic
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

            console.log(data);

            const lastModifiedStr = FormatDateTime(new Date(data[0].lastModified));
            const createdAtStr = FormatDateTime(new Date(data[0].createdAt));

            setId(data[0].id);
            setName(data[0].name);
            setAddress(data[0].address);
            setCompanyPhoneNumber(data[0].companyPhoneNumber);
            setContactPersonName(data[0].contactPersonName);
            setTerms(data[0].terms);
            if (data[0].vendorIsTaxable === "true") {
              setVendorIsTaxable(true);
            } else {
              setVendorIsTaxable(false);
            }

            setLastModified(lastModifiedStr);
            setLastModifiedByUserName(data[0].lastModifiedByUserName);
            setCreatedAt(createdAtStr);
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
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: Number(id),
          newData: {
            name: name,
            address: address,
            companyPhoneNumber: companyPhoneNumber,
            contactPersonName: contactPersonName,
            contactPersonNumber: contactPersonNumber,
            terms: terms,
            vendorIsTaxable: vendorIsTaxable,
          },
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
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          name: name,
          address: address,
          companyPhoneNumber: companyPhoneNumber,
          contactPersonName: contactPersonName,
          contactPersonNumber: contactPersonNumber,
          terms: terms,
          vendorIsTaxable: vendorIsTaxable,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Invalid credentials or network issue");
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
            <label htmlFor="modify-supplier-id">ID:</label>
            <input type="text" id="modify-supplier-id" value={id} readOnly />
          </div>
        )}

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-name">Name:</label>
          <input
            type="text"
            id="modify-supplier-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-address">Address:</label>
          <textarea
            id="modify-supplier-address"
            value={address}
            onChange={handleAddressChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-company-pn">Company Phone Number:</label>
          <input
            type="text"
            id="modify-supplier-company-pn"
            value={companyPhoneNumber}
            onChange={handleCompanyPhoneNumberChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-cp-name">Contact Person Name:</label>
          <input
            type="text"
            id="modify-supplier-cp-name"
            value={contactPersonName}
            onChange={handleContactPersonNameChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-cp-pn">Contact Person Number:</label>
          <input
            type="text"
            id="modify-supplier-cp-pn"
            value={contactPersonNumber}
            onChange={handleContactPersonNumberChange}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-terms">Terms:</label>
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
            <label htmlFor="modify-supplier-radio-yes">Yes</label>

            <input
              type="radio"
              checked={vendorIsTaxable === false}
              id="modify-supplier-radio-no"
              name="vendorIsTaxable"
              value={"no"}
              onChange={handleVendorIsTaxableChange}
            />
            <label htmlFor="modify-supplier-radio-no">No</label>
          </div>
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-last-modified">Last Modified:</label>
          <input
            type="text"
            id="modify-supplier-last-modified"
            value={lastModified}
            readOnly={true}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-last-modified-by">Last Modified By:</label>
          <input
            type="text"
            id="modify-supplier-last-modified-by"
            value={lastModifiedByUserName}
            readOnly={true}
          />
        </div>

        <div className="supplier-data-form-group">
          <label htmlFor="modify-supplier-created-at">Created At:</label>
          <input
            type="text"
            id="modify-supplier-created-at"
            value={createdAt}
            readOnly={true}
          />
        </div>
        
      </div>

      <div className="modify-supplier-buttons">
        <div className="modify-supplier-btns-grp">
        <button
            type="button"
            className="modify-supplier-delete-btn"
            onClick={handleDelete}
          >
            Delete Supplier
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
            className="modify-supplier-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifySupplierPage;
