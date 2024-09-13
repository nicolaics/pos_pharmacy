import React, { useState } from "react";

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
  row.appendChild(idCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  row.appendChild(nameCell);

  const addressCell = document.createElement("td");
  addressCell.textContent = data["address"];
  row.appendChild(addressCell);

  const companyPhoneNumberCell = document.createElement("td");
  companyPhoneNumberCell.textContent = data["companyPhoneNumber"];
  row.appendChild(companyPhoneNumberCell);

  const contactPersonNameCell = document.createElement("td");
  contactPersonNameCell.textContent = data["contactPersonName"];
  row.appendChild(contactPersonNameCell);

  const contactPersonNumberCell = document.createElement("td");
  contactPersonNumberCell.textContent = data["contactPersonNumber"];
  row.appendChild(contactPersonNumberCell);

  const termsCell = document.createElement("td");
  termsCell.textContent = data["terms"];
  row.appendChild(termsCell);

  const vendorIsTaxableCell = document.createElement("td");
  vendorIsTaxableCell.textContent = data["vendorIsTaxable"];
  row.appendChild(vendorIsTaxableCell);

  const lastModified = new Date(data["lastModified"]);
  const lastModifiedCell = document.createElement("td");
  lastModifiedCell.textContent = FormatDateTime(lastModified);
  row.appendChild(lastModifiedCell);

  const lastModifiedByUserIdCell = document.createElement("td");
  lastModifiedByUserIdCell.textContent = data["lastModifiedByUserId"];
  row.appendChild(lastModifiedByUserIdCell);

  const createdAt = new Date(data["createdAt"]);
  const createdAtCell = document.createElement("td");
  createdAtCell.textContent = FormatDateTime(createdAt);
  row.appendChild(createdAtCell);

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
  const [searchParams, setSearchParams] = useState();

  const handleSearchValChange = (e: any) => {
    e.preventDefault();
    setSearchVal(e.target.value);
  };

  const handleSearchParamsChange = (e: any) => {
    e.preventDefault();
    setSearchParams(e.target.value);
  };

  const testData = [
    {
      id: 1,
      name: "John Doe 1",
      address: "ASdlasij",
      companyPhoneNumber: "010-000-10200",
      contactPersonName: "John",
      contactPersonNumber: "000-0000-0000",
      terms: "Y",
      vendorIsTaxable: true,
      createdAt: "2024-08-01 12.10",
      lastModified: "2024-08-10 12.10",
      lastModifiedByUserId: 1,
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
      lastModifiedByUserId: 1,
    },
  ];

  const search = () => {
    const token = sessionStorage.getItem("token");
    var getAllSupplierURL = "";
    
    if (searchVal === "") {
      getAllSupplierURL = `http://${BACKEND_BASE_URL}/supplier/all/all`;
    }
    else {
      getAllSupplierURL = `http://${BACKEND_BASE_URL}/supplier/${searchParams}/${searchVal}`;
    }

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
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const register = () => {
    navigate("/supplier/detail");
  };

  // TODO: add search options for the parameters
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
      </div>

      <div className="supplier-table-container">
        <table id="supplier-data-table" border={1}>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Address</th>
              <th>Company Phone Number</th>
              <th>Contact Person Name</th>
              <th>Contact Person Number</th>
              <th>Terms</th>
              <th>Vendor is Taxable</th>
              <th>Last Modified</th>
              <th>Last Modified By ID</th>
              <th>Created At</th>
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
