import React, { useState } from "react";

import "./Customer.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";
import { BsPersonFillAdd } from "react-icons/bs";

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

  const createdAt = new Date(data["createdAt"]);
  const createdAtCell = document.createElement("td");
  createdAtCell.textContent = FormatDateTime(createdAt);
  row.appendChild(createdAtCell);

  row.addEventListener("click", () => {
    navigate("/customer/detail", {
      state: {
        id: data["id"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewCustomerPage: React.FC = () => {
  const navigate = useNavigate();

  const search = () => {
    const token = sessionStorage.getItem("token");
    const getAllCustomerURL = "http://localhost:19230/api/v1/customer";

    fetch(getAllCustomerURL, {
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

  const register = () => {
    navigate("/customer/detail");
  }

  return (
    <div className="view-customer-page">
      <h1>User</h1>

      <div className="customer-container">
        <span className="customer-search">
          <input type="text" className="customer-search-item" />
          <MdPersonSearch size={40} className="customer-search-btn" onClick={search} />
          <BsPersonFillAdd size={40} className="customer-add-btn" onClick={register} />
        </span>
        <div className="customer-table">
          <table id="customer-data-table" border={1}>
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody></tbody>
          </table>
        </div>
      </div>

      <div className="customer-grid-item" onClick={returnToHome}>
          <FaHome size={50} />
          <h2>Back to Home</h2>
        </div>
    </div>
  );
};

export default ViewCustomerPage;
