import React, { useEffect, useState } from "react";

import "./Supplier.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { FaHome, FaSearch } from "react-icons/fa";
import { BsBuildingFillAdd } from "react-icons/bs";
import { BACKEND_BASE_URL } from "../../../App";

function fillTable(
  data: any,
  tableBody: Element | null,
  navigate: NavigateFunction
) {
  if (!tableBody) return;

  // Loop through each supplier and create a new row in the table
  const row = document.createElement("tr");
  row.className = "view-supplier-data-row";

  // Create and append cells for each column
  const idCell = document.createElement("td");
  idCell.textContent = data["id"].toString();
  idCell.className = "supplier-id-column";
  row.appendChild(idCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  nameCell.className = "supplier-name-column";
  row.appendChild(nameCell);

  const addressCell = document.createElement("td");
  addressCell.textContent = data["address"];
  addressCell.className = "supplier-address-column";
  row.appendChild(addressCell);

  const companyPhoneNumberCell = document.createElement("td");
  companyPhoneNumberCell.textContent = data["companyPhoneNumber"];
  companyPhoneNumberCell.className = "supplier-company-pn-column";
  row.appendChild(companyPhoneNumberCell);

  const contactPersonNameCell = document.createElement("td");
  contactPersonNameCell.textContent = data["contactPersonName"];
  contactPersonNameCell.className = "supplier-cp-name-column";
  row.appendChild(contactPersonNameCell);

  const contactPersonNumberCell = document.createElement("td");
  contactPersonNumberCell.textContent = data["contactPersonNumber"];
  contactPersonNumberCell.className = "supplier-cp-pn-column";
  row.appendChild(contactPersonNumberCell);

  // const termsCell = document.createElement("td");
  // termsCell.textContent = data["terms"];
  // termsCell.className = "supplier-terms-column";
  // row.appendChild(termsCell);

  // const vendorIsTaxableCell = document.createElement("td");
  // vendorIsTaxableCell.textContent = data["vendorIsTaxable"];
  // vendorIsTaxableCell.className = "supplier-vendor-is-taxable-column";
  // row.appendChild(vendorIsTaxableCell);

  // const lastModified = new Date(data["lastModified"]);
  // const lastModifiedCell = document.createElement("td");
  // lastModifiedCell.textContent = FormatDateTime(lastModified);
  // lastModifiedCell.className = "supplier-last-modified-column";
  // row.appendChild(lastModifiedCell);

  // const lastModifiedByUserNameCell = document.createElement("td");
  // lastModifiedByUserNameCell.textContent = data["lastModifiedByUserName"];
  // lastModifiedByUserNameCell.className = "supplier-last-modified-by-column";
  // row.appendChild(lastModifiedByUserNameCell);

  // const createdAt = new Date(data["createdAt"]);
  // const createdAtCell = document.createElement("td");
  // createdAtCell.textContent = FormatDateTime(createdAt);
  // createdAtCell.className = "supplier-created-at-column";
  // row.appendChild(createdAtCell);

  row.addEventListener("dblclick", () => {
    navigate("/supplier/detail", {
      state: {
        id: data["id"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewSupplierPage: React.FC = () => {
  const navigate = useNavigate();

  const [searchVal, setSearchVal] = useState("");
  const [searchParams, setSearchParams] = useState("none");

  const handleSearchValChange = (e: any) => {
    e.preventDefault();
    setSearchVal(e.target.value);
  };

  const handleSearchParamsChange = (e: any) => {
    // e.preventDefault();
    setSearchParams(e.target.value);
  };

  useEffect(() => {
    const token = sessionStorage.getItem("token");
    const getAllSupplierURL = `http://${BACKEND_BASE_URL}/supplier/all/all`;
    
    fetch(getAllSupplierURL, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Invalid credentials or network issue");
          }

          console.log(data);

          const tableBody = document.querySelector("#supplier-data-table tbody");
          if (!tableBody) {
            console.error("table body not found");
            return;
          }

          tableBody.innerHTML = "";

          for (let i = 0; i < data.length; i++) {
            fillTable(data[i], tableBody, navigate);
          }
        })
      )
      .catch((error) => {
        console.error("Error loading supplier data:", error);
        alert("Error loading supplier data");
      });
  });

  const testData = [
    {
      id: 1,
      name: "John Doe 1 dfashbdkolfahdlifandm,absd",
      address: "ASdlasij",
      companyPhoneNumber: "010-000-10200",
      contactPersonName: "John",
      contactPersonNumber: "000-0000-0000",
      terms: "Y",
      vendorIsTaxable: true,
      createdAt: "2024-08-01 12.10",
      lastModified: "2024-08-10 12.10",
      lastModifiedByUserName: "sakudhadsi",
    },
    {
      id: 2,
      name: "John Doe 2",
      address: "ASdlasij",
      companyPhoneNumber: "010-000-10200",
      contactPersonName: "John",
      contactPersonNumber: "000-0000-0000",
      terms: "Y",
      vendorIsTaxable: false,
      createdAt: "2024-08-01 12.10",
      lastModified: "2024-08-10 12.10",
      lastModifiedByUserName: "sadfasdf",
    },
  ];

  const search = () => {
    const token = sessionStorage.getItem("token");
    var getAllSupplierURL = "";

    // TEST DATA
    // const tableBody = document.querySelector("#supplier-data-table tbody");
    // if (!tableBody) {
    //   console.error("table body not found");
    //   return;
    // }
    // tableBody.innerHTML = "";
    // for (let i = 0; i < testData.length; i++) {
    //   fillTable(testData[i], tableBody, navigate);
    // }
    
    if (searchVal === "") {
      getAllSupplierURL = `http://${BACKEND_BASE_URL}/supplier/all/all`;
    } else {
      console.log(searchParams);
      if (searchParams === "none") {
        alert("search by cannot be none!");
        return;
      }

      getAllSupplierURL = `http://${BACKEND_BASE_URL}/supplier/${searchParams}/${searchVal}`;
    }

    fetch(getAllSupplierURL, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Invalid credentials or network issue");
          }

          const tableBody = document.querySelector(
            "#supplier-data-table tbody"
          );
          if (!tableBody) {
            console.error("table body not found");
            return;
          }

          tableBody.innerHTML = "";

          for (let i = 0; i < data.length; i++) {
            fillTable(data[i], tableBody, navigate);
          }
        })
      )
      .catch((error) => {
        console.error("Error loading supplier data:", error);
        alert("Error loading supplier data");
      });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const register = () => {
    navigate("/supplier/detail");
  };

  return (
    <div className="view-supplier-page">
      <h1>Supplier</h1>

      <div className="supplier-search-container">
        <input
          type="text"
          className="supplier-search-input"
          placeholder="Search"
          value={searchVal}
          onChange={handleSearchValChange}
        />
        <button onClick={search} className="supplier-search-button">
          <FaSearch size={30} id="supplier-search-icon" />
          Search
        </button>

        <button onClick={register} className="supplier-add-button">
          <BsBuildingFillAdd size={30} id="supplier-add-icon" />
          Add
        </button>

        <div className="supplier-search-radio-container">
          <label>Search By:</label>
          <div className="supplier-search-radio-grp">
            <div className="supplier-search-radio-item">
              <input
                type="radio"
                id="supplier-search-radio-none"
                checked={true}
                name="searchParams"
                value={"none"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="supplier-search-radio-none">None</label>
            </div>

            <div className="supplier-search-radio-item">
              <input
                type="radio"
                id="supplier-search-radio-name"
                name="searchParams"
                value={"name"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="supplier-search-radio-name">Supplier Name</label>
            </div>

            <div className="supplier-search-radio-item">
              <input
                type="radio"
                id="supplier-search-radio-cp-name"
                name="searchParams"
                value={"cp-name"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="supplier-search-radio-cp-name">
                Contact Person Name
              </label>
            </div>

            <div className="supplier-search-radio-item">
              <input
                type="radio"
                id="supplier-search-radio-id"
                name="searchParams"
                value={"id"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="supplier-search-radio-id">ID</label>
            </div>
          </div>
        </div>
      </div>

      <div className="supplier-table-container">
        <table id="supplier-data-table" border={1}>
          <thead>
            <tr>
              <th className="supplier-id-column">ID</th>
              <th className="supplier-name-column">Name</th>
              <th className="supplier-address-column">Address</th>
              <th className="supplier-company-pn-column">Company Phone Number</th>
              <th className="supplier-cp-name-column">Contact Person Name</th>
              <th className="supplier-cp-pn-column">Contact Person Number</th>
              {/* <th className="supplier-terms-column">Terms</th>
              <th className="supplier-vendor-is-taxable-column">Vendor is Taxable</th>
              <th className="supplier-last-modified-column">Last Modified</th>
              <th className="supplier-last-modified-by-column">Last Modified By</th>
              <th className="supplier-created-at-column">Created At</th> */}
            </tr>
          </thead>
          <tbody></tbody>
        </table>
      </div>

      <button onClick={returnToHome} className="view-supplier-home-button">
        <FaHome size={30} id="view-supplier-home-icon" />
        Back to Home
      </button>
    </div>
  );
};

export default ViewSupplierPage;
