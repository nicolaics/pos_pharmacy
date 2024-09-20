import React, { useState } from "react";

import "./ViewUser.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";
import { IoArrowUndoOutline } from "react-icons/io5";
import { BACKEND_BASE_URL } from "../../../App";

function fillTable(
  data: any,
  tableBody: Element | null,
  navigate: NavigateFunction
) {
  if (!tableBody) return;

  // Loop through each user and create a new row in the table
  const row = document.createElement("tr");
  row.className = "view-user-data-row";

  // Create and append cells for each column
  const idCell = document.createElement("td");
  idCell.textContent = data["id"].toString();
  idCell.className = "user-id-column";
  row.appendChild(idCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  nameCell.className = "user-name-column";
  row.appendChild(nameCell);

  const adminCell = document.createElement("td");
  if (data["admin"]) {
    adminCell.textContent = "Yes";
  } else {
    adminCell.textContent = "No";
  }
  adminCell.className = "user-admin-column";

  row.appendChild(adminCell);

  const phoneNumberCell = document.createElement("td");
  phoneNumberCell.textContent = data["phoneNumber"];
  phoneNumberCell.className = "user-phone-number-column";
  row.appendChild(phoneNumberCell);

  const lastLoggedInCell = document.createElement("td");

  const lastLoggedIn = new Date(data["lastLoggedIn"]);
  lastLoggedInCell.textContent = FormatDateTime(lastLoggedIn);
  lastLoggedInCell.className = "user-last-logged-in-column";
  row.appendChild(lastLoggedInCell);

  const createdAt = new Date(data["createdAt"]);
  const createdAtCell = document.createElement("td");
  createdAtCell.textContent = FormatDateTime(createdAt);
  createdAtCell.className = "user-created-at-column";
  row.appendChild(createdAtCell);

  row.addEventListener("dblclick", () => {
    navigate("/user/detail", {
      state: {
        reqType: "modify-admin",
        id: data["id"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewUserPage: React.FC = () => {
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
      id: 1,
      name: "John Doe 1",
      admin: true,
      phoneNumber: "010-4444-1111",
      lastLoggedIn: "2024-08-03 15:30",
      createdAt: "2024-08-01 12.10",
    },
    {
      id: 2,
      name: "John Doe 2",
      admin: true,
      phoneNumber: "010-4444-1111",
      lastLoggedIn: "2024-08-03 15:30",
      createdAt: "2024-08-01 12.10",
    },
  ];

  const search = () => {
    const token = sessionStorage.getItem("token");
    var getAllUserURL = "";
    
    // TEST DATA
    // const tableBody = document.querySelector("#user-data-table tbody");
    // if (!tableBody) {
    //   console.error("table body not found");
    //   return;
    // }
    // tableBody.innerHTML = "";
    // for (let i = 0; i < testData.length; i++) {
    //   fillTable(testData[i], tableBody, navigate);
    // }

    if (searchVal === "") {
      getAllUserURL = `http://${BACKEND_BASE_URL}/user/all/all`;
    }
    else {
      if (searchParams === "none") {
        alert("search by cannot be none!");
        return;
      }

      getAllUserURL = `http://${BACKEND_BASE_URL}/user/${searchParams}/${searchVal}`
    }

    console.log(getAllUserURL);

    fetch(getAllUserURL, {
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

          const tableBody = document.querySelector("#user-data-table tbody");
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
        console.error("Error loading user data:", error);
        alert("Error loading user data");
      });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const returnToUserLandingPage = () => {
    navigate("/user");
  }

  return (
    <div className="view-user-page">
      <h1>User</h1>

      <div className="user-search-container">
        <input
          type="text"
          className="user-search-input"
          placeholder="Search"
          value={searchVal}
          onChange={handleSearchValChange}
        />
        <button onClick={search} className="user-search-button">
          <MdPersonSearch size={30} id="user-search-icon" />
          Search
        </button>

        <div className="user-search-radio-container">
          <label>Search By:</label>
          <div className="user-search-radio-grp">
            <div className="user-search-radio-item">
              <input
                type="radio"
                id="user-search-radio-none"
                checked={true}
                name="searchParams"
                value={"none"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="user-search-radio-none">None</label>
            </div>

            <div className="user-search-radio-item">
              <input
                type="radio"
                id="user-search-radio-name"
                name="searchParams"
                value={"name"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="user-search-radio-name">Name</label>
            </div>

            <div className="user-search-radio-item">
              <input
                type="radio"
                id="user-search-radio-phone-number"
                name="searchParams"
                value={"phone-number"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="user-search-radio-phone-number">
                Phone Number
              </label>
            </div>

            <div className="user-search-radio-item">
              <input
                type="radio"
                id="user-search-radio-id"
                name="searchParams"
                value={"id"}
                onChange={handleSearchParamsChange}
              />
              <label htmlFor="user-search-radio-id">ID</label>
            </div>
          </div>
        </div>
      </div>
      <div className="user-table-container">
        <table id="user-data-table" border={1}>
          <thead>
            <tr>
              <th className="user-id-column">ID</th>
              <th className="user-name-column">Name</th>
              <th className="user-admin-column">Admin</th>
              <th className="user-phone-number-column">Phone Number</th>
              <th className="user-last-logged-in-column">Last Logged In</th>
              <th className="user-created-at-column">Created At</th>
            </tr>
          </thead>
          <tbody></tbody>
        </table>
      </div>

      <div className="user-return-btns-grp">
      <button onClick={returnToUserLandingPage} className="view-user-back-button">
          <IoArrowUndoOutline size={30} id="view-user-back-icon" />
          Cancel
        </button>

        <button onClick={returnToHome} className="view-user-home-button">
          <FaHome size={30} id="view-user-home-icon" />
          Back to Home
        </button>
      </div>
    </div>
  );
};

export default ViewUserPage;
