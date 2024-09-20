import React, { useState } from "react";

import "./Inventory.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";
import { IoArrowUndoOutline } from "react-icons/io5";
import { BACKEND_BASE_URL } from "../../../App";
import { BsPersonFillAdd } from "react-icons/bs";

// id-ID: use dot as seperator (Indonesia)
// en-US: use comma as seperator (US)
const NUMBER_LOCALE = "id-ID";

function fillTable(
  data: any,
  tableBody: Element | null,
  navigate: NavigateFunction
) {
  if (!tableBody) return;

  // Loop through each inventory and create a new row in the table
  const row = document.createElement("tr");
  row.className = "view-inventory-data-row";

  // Create and append cells for each column
  const idCell = document.createElement("td");
  idCell.textContent = data["id"].toString();
  idCell.className = "inventory-id-column";
  row.appendChild(idCell);

  const barcodeCell = document.createElement("td");
  barcodeCell.textContent = data["barcode"].toString();
  barcodeCell.className = "inventory-barcode-column";
  row.appendChild(barcodeCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  nameCell.className = "inventory-name-column";
  row.appendChild(nameCell);

  const stockCell = document.createElement("td");
  stockCell.textContent = data["qty"].toLocaleString(NUMBER_LOCALE);
  stockCell.className = "inventory-stock-column";
  row.appendChild(stockCell);

  const unitCell = document.createElement("td");
  unitCell.textContent = data["firstUnitName"];
  unitCell.className = "inventory-unit-column";
  row.appendChild(unitCell);

  const priceCell = document.createElement("td");
  priceCell.textContent =
    "Rp. " + data["firstPrice"].toLocaleString(NUMBER_LOCALE);
  priceCell.className = "inventory-price-column";
  row.appendChild(priceCell);

  row.addEventListener("dblclick", () => {
    navigate("/inventory/detail", {
      state: {
        reqType: "modify",
        id: data["id"],
        barcode: data["barcode"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewInventoryPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchVal, setSearchVal] = useState("");
  const [searchParams, setSearchParams] = useState("");

  const handleSearchValChange = (e: any) => {
    e.preventDefault();
    setSearchVal(e.target.value);
  };

  const handleSearchParamsChange = (e: any) => {
    setSearchParams(e.target.value);
  };

  const testData = [
    {
      id: 3000,
      barcode: "9999991818288282",
      name: "John Doe 1",
      qty: 100,
      firstUnitName: "TAB",
      firstPrice: 10000.219,
    },
    {
      id: 2,
      barcode: "183819381021",
      name: "John Doe 2",
      qty: 29.1,
      firstUnitName: "BTL",
      firstPrice: 19282.1,
    },
  ];

  const search = () => {
    const token = sessionStorage.getItem("token");
    var getAllInventoryURL = "";

    // TEST DATA
    const tableBody = document.querySelector("#inventory-data-table tbody");
    if (!tableBody) {
      console.error("table body not found");
      return;
    }
    tableBody.innerHTML = "";
    for (let i = 0; i < testData.length; i++) {
      fillTable(testData[i], tableBody, navigate);
    }

    // if (searchVal === "") {
    //   getAllInventoryURL = `http://${BACKEND_BASE_URL}/medicine/all/all`;
    // }
    // else {
    //   if (searchParams === "none") {
    //     alert("search by cannot be none!");
    //     return;
    //   }

    //   getAllInventoryURL = `http://${BACKEND_BASE_URL}/medicine/${searchParams}/${searchVal}`
    // }

    // console.log(getAllInventoryURL);

    // fetch(getAllInventoryURL, {
    //   method: "GET",
    //   headers: {
    //     "Content-Type": "application/json",
    //     Authorization: "Bearer " + token,
    //   },
    // })
    //   .then((response) =>
    //     response.json().then((data) => {
    //       if (!response.ok) {
    //         throw new Error("Invalid credentials or network issue");
    //       }

    //       const tableBody = document.querySelector("#inventory-data-table tbody");
    //       if (!tableBody) {
    //         console.error("table body not found");
    //         return;
    //       }

    //       tableBody.innerHTML = "";

    //       for (let i = 0; i < data.length; i++) {
    //         fillTable(data[i], tableBody, navigate);
    //       }
    //     })
    //   )
    //   .catch((error) => {
    //     console.error("Error loading inventory data:", error);
    //     alert("Error loading inventory data");
    //   });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const register = () => {
    navigate("/inventory/detail", {
      state: {
        reqType: "add",
      },
    });
  };

  return (
    <div className="view-inventory-page">
      <h1>Inventory</h1>

      <div className="inventory-search-container">
        <div className="inventory-search-input-container">
          <input
            type="text"
            className="inventory-search-input"
            placeholder="Search"
            value={searchVal}
            onChange={handleSearchValChange}
          />
          <button onClick={search} className="inventory-search-button">
            <MdPersonSearch size={30} id="inventory-search-icon" />
            Search
          </button>

          <button onClick={register} className="inventory-add-button">
            <BsPersonFillAdd size={30} id="inventory-add-icon" />
            Add
          </button>
        </div>

        <div className="inventory-search-radio-container">
          <label>Search By:</label>
          <div className="inventory-search-radio-grp">
            <div className="inventory-search-radio-item">
              <input
                type="radio"
                id="inventory-search-radio-none"
                checked={!searchParams === true}
                name="searchParams"
                value={"none"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="inventory-search-radio-none">None</label>
            </div>

            <div className="inventory-search-radio-item">
              <input
                type="radio"
                id="inventory-search-radio-barcode"
                name="searchParams"
                value={"barcode"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="inventory-search-radio-barcode">Barcode</label>
            </div>

            <div className="inventory-search-radio-item">
              <input
                type="radio"
                id="inventory-search-radio-name"
                name="searchParams"
                value={"name"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="inventory-search-radio-name">Name</label>
            </div>

            <div className="inventory-search-radio-item">
              <input
                type="radio"
                id="inventory-search-radio-id"
                name="searchParams"
                value={"id"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="inventory-search-radio-id">ID</label>
            </div>

            <div className="inventory-search-radio-item">
              <input
                type="radio"
                id="inventory-search-radio-description"
                name="searchParams"
                value={"description"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="inventory-search-radio-description">
                Description
              </label>
            </div>
          </div>
        </div>
      </div>

      <div className="inventory-table-container">
        <table id="inventory-data-table" border={1}>
          <thead>
            <tr>
              <th className="inventory-id-column">ID</th>
              <th className="inventory-barcode-column">Barcode</th>
              <th className="inventory-name-column">Name</th>
              <th className="inventory-stock-column">Stock</th>
              <th className="inventory-unit-column">Unit</th>
              <th className="inventory-price-column">Price</th>
            </tr>
          </thead>
          <tbody></tbody>
        </table>
      </div>

      <div className="inventory-return-btns-grp">
        <button onClick={returnToHome} className="view-inventory-home-button">
          <FaHome size={30} id="view-inventory-home-icon" />
          Back to Home
        </button>
      </div>
    </div>
  );
};

export default ViewInventoryPage;
