import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../App";

const ModifyCompanyPage: React.FC = () => {
  const navigate = useNavigate();

  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [businessNumber, setBusinessNumber] = useState("");
  const [pharmacist, setPharmacist] = useState("");
  const [pharmacistLicenseNumber, setPharmacistLicenseNumber] = useState("");
  const [lastModified, setLastModified] = useState("");
  const [lastModifiedByUserName, setLastModifiedByUserName] = useState("");

  var newCompany = false;

  useEffect(() => {
    const token = sessionStorage.getItem("token");
    const companyURL = `http://${BACKEND_BASE_URL}/company-profile`; // Set the URL or handle this logic
    fetch(companyURL, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Unable to modify company data");
          }

          console.log("company", data);

          if (data.id === 0) {
            newCompany = true;
            return;
          }

          setId(data.id);
          setName(data.name);
          setAddress(data.address);
          setBusinessNumber(data.businessNumber);
          setPharmacist(data.pharmacist);
          setPharmacistLicenseNumber(data.pharmacistLicenseNumber);
          setLastModified(data.lastModified);
          setLastModifiedByUserName(data.lastModifiedByUserName);
        })
      )
      .catch((error) => {
        console.error("Error load selected company:", error);
        alert("Error load selected company");
      });
  });

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handleAddressChange = (event: any) => {
    setAddress(event.target.value);
  };

  const handleBusinessNumberChange = (event: any) => {
    setBusinessNumber(event.target.value);
  };

  const handlePharmacistChange = (event: any) => {
    setPharmacist(event.target.value);
  };

  const handlePharmacistLicenseNumberChange = (event: any) => {
    setPharmacistLicenseNumber(event.target.value);
  };

  const handleLastModifiedChange = (event: any) => {
    setLastModified(event.target.value);
  };

  const handleLastModifiedByUserNameChange = (event: any) => {
    setLastModifiedByUserName(event.target.value);
  };

  const handleSendRequest = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/company`;

    if (newCompany === true) {
      fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          name: name,
          address: address,
          businessNumber: businessNumber,
          pharmacist: pharmacist,
          pharmacistLicenseNumber: pharmacistLicenseNumber,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to create company data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error create company:", error);
          alert("Error create company");
        });
    } else {
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
            businessNumber: businessNumber,
            pharmacist: pharmacist,
            pharmacistLicenseNumber: pharmacistLicenseNumber,
          },
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify company data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify company:", error);
          alert("Error modify company");
        });
    }

    // Reset the state
    navigate("/company");
  };

  const handleCancel = (e: any) => {
    navigate("/home");
  };

  return (
    <div className="modify-company-page">
      <h1>Modify Company Profile</h1>

      <div className="company-data-container">
        <div className="company-data-form-group">
          <label htmlFor="id">ID:</label>
          <input type="text" id="modify-company-id" value={id} readOnly />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-name">Name:</label>
          <input
            type="text"
            id="modify-company-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-address">Address:</label>
          <textarea
            id="modify-company-address"
            value={address}
            onChange={handleAddressChange}
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-businessNumber">
            Business Number:
          </label>
          <input
            type="text"
            id="modify-company-businessNumber"
            value={businessNumber}
            onChange={handleBusinessNumberChange}
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-pharmacist">Pharmacist:</label>
          <input
            type="text"
            id="modify-company-pharmacist"
            value={pharmacist}
            onChange={handlePharmacistChange}
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-pharmacistLicenseNumber">
            Pharmacist License Number:
          </label>
          <input
            type="text"
            id="modify-company-pharmacistLicenseNumber"
            value={pharmacistLicenseNumber}
            onChange={handlePharmacistLicenseNumberChange}
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-lastModified">Last Modified:</label>
          <input
            type="text"
            id="modify-company-lastModified"
            value={lastModified}
            onChange={handleLastModifiedChange}
            readOnly
          />
        </div>

        <div className="company-data-form-group">
          <label htmlFor="modify-company-lastModifiedByUserName">
            Last Modified By User ID:
          </label>
          <input
            type="text"
            id="modify-company-lastModifiedByUserName"
            value={lastModifiedByUserName}
            onChange={handleLastModifiedByUserNameChange}
            readOnly
          />
        </div>
      </div>

      <div className="modify-company-buttons">
        <div className="modify-company-btns-grp">
          <button
            type="button"
            className="modify-company-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="button"
            className="modify-company-ok-btn"
            onClick={handleSendRequest}
          >
            Modify
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifyCompanyPage;
