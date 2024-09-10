import React, { useState } from "react";

import "./ViewUser.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";
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
  row.appendChild(idCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  row.appendChild(nameCell);

  const adminCell = document.createElement("td");
  if (data["admin"]) {
    adminCell.textContent = "Yes";
  } else {
    adminCell.textContent = "No";
  }

  row.appendChild(adminCell);

  const phoneNumberCell = document.createElement("td");
  phoneNumberCell.textContent = data["phoneNumber"];
  row.appendChild(phoneNumberCell);

  const lastLoggenInCell = document.createElement("td");

  const lastLoggedIn = new Date(data["lastLoggedIn"]);
  lastLoggenInCell.textContent = FormatDateTime(lastLoggedIn);
  row.appendChild(lastLoggenInCell);

  const createdAt = new Date(data["createdAt"]);
  const createdAtCell = document.createElement("td");
  createdAtCell.textContent = FormatDateTime(createdAt);
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
    const getAllUserURL = `http://${BACKEND_BASE_URL}/user`;


    // TEST DATA
    const tableBody = document.querySelector("#user-data-table tbody");
    if (!tableBody) {
      console.error("table body not found");
      return;
    }
    tableBody.innerHTML = "";
    for (let i = 0; i < testData.length; i++) {
      fillTable(testData[i], tableBody, navigate);
    }

    // fetch(getAllUserURL, {
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

    //       const tableBody = document.querySelector("#user-data-table tbody");
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
    //     console.error("Error loading user data:", error);
    //     alert("Error loading user data");
    //   });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  return (
    <div className="view-user-page">
      <h1>User</h1>

      <div className="user-search-container">
        <input type="text" className="user-search-input" placeholder="Search" />
        <button onClick={search} className="user-search-button">
          <MdPersonSearch size={30} id="user-search-icon" />
          Search
        </button>
      </div>
      <div className="user-table-container">
        <table id="user-data-table" border={1}>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Admin</th>
              <th>Phone Number</th>
              <th>Last Logged In</th>
              <th>Created At</th>
            </tr>
          </thead>
          <tbody></tbody>
        </table>
      </div>

      <button onClick={returnToHome} className="view-user-home-button">
        <FaHome size={30} id="view-user-home-icon" />
        Back to Home
      </button>
    </div>
  );
};

export default ViewUserPage;
