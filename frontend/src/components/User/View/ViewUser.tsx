import React, { useState } from "react";

import "./ViewUser.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";

// TODO: notify user if they have selected something (highlight or something in css)
function fillTable(
  data: any,
  tableBody: Element | null,
  navigate: NavigateFunction
) {
  if (!tableBody) return;

  // Loop through each user and create a new row in the table
  const row = document.createElement("tr");

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

  row.addEventListener("click", () => {
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

  const search = () => {
    const token = sessionStorage.getItem("token");
    const getAllUserURL = "http://localhost:19230/api/v1/user";

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

          const tableBody = document.querySelector("#dataTable tbody");
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

  return (
    <div className="view-user-page">
      <h1>User</h1>

      <div className="container">
        <span className="search">
          <input type="text" className="search-item" />
          <MdPersonSearch size={40} className="search-icon" onClick={search} />
        </span>
        <div className="table">
          <table id="dataTable" border={1}>
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
      </div>

      <div className="view-user-grid-item" onClick={returnToHome}>
          <FaHome size={50} />
          <h2>Back to Home</h2>
        </div>
    </div>
  );
};

export default ViewUserPage;
