import React, { useEffect, useState } from "react";

import "./Patient.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import { MdPersonSearch } from "react-icons/md";
import { FaHome } from "react-icons/fa";
import { BsPersonFillAdd } from "react-icons/bs";
import { BACKEND_BASE_URL } from "../../../App";

function fillTable(
  data: any,
  tableBody: Element | null,
  navigate: NavigateFunction
) {
  if (!tableBody) return;

  // Loop through each patient and create a new row in the table
  const row = document.createElement("tr");
  row.className = "view-patient-data-row";

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

  row.addEventListener("dblclick", () => {
    navigate("/patient/detail", {
      state: {
        id: data["id"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewPatientPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchVal, setSearchVal] = useState("");

  const handleSearchValChange = (e: any) => {
    e.preventDefault();
    setSearchVal(e.target.value);
  };

  useEffect(() => {
    const token = sessionStorage.getItem("token");
    const getAllPatientURL = `http://${BACKEND_BASE_URL}/patient/all`;
    
    fetch(getAllPatientURL, {
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

          const tableBody = document.querySelector("#patient-data-table tbody");
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
        console.error("Error loading patient data:", error);
        alert("Error loading patient data");
      });
  });

  const testData = [
    {
      id: 1,
      name: "John Doe 1",
      createdAt: "2024-08-01 12.10",
    },
    {
      id: 2,
      name: "John Doe 2",
      createdAt: "2024-08-01 12.10",
    },
  ];

  const search = () => {
    const token = sessionStorage.getItem("token");
    var getPatientURL = "";

    if (searchVal === "") {
      getPatientURL = `http://${BACKEND_BASE_URL}/patient/all`;
    }
    else {
      getPatientURL = `http://${BACKEND_BASE_URL}/patient/${searchVal}`;
    }

    // TEST DATA
    // const tableBody = document.querySelector("#patient-data-table tbody");
    // if (!tableBody) {
    //   console.error("table body not found");
    //   return;
    // }
    // tableBody.innerHTML = "";
    // for (let i = 0; i < testData.length; i++) {
    //   fillTable(testData[i], tableBody, navigate);
    // }

    fetch(getPatientURL, {
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

          const tableBody = document.querySelector("#patient-data-table tbody");
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
        console.error("Error loading patient data:", error);
        alert("Error loading patient data");
      });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const register = () => {
    navigate("/patient/detail");
  };

  return (
    <div className="view-patient-page">
      <h1>Patient</h1>
      
      <div className="patient-container">
        <input
          type="text"
          className="patient-search-input"
          placeholder="Search"
          value={searchVal}
          onChange={handleSearchValChange}
        />
        <button onClick={search} className="patient-search-button">
          <MdPersonSearch size={30} id="patient-search-icon" />
          Search
        </button>

        <button onClick={register} className="patient-add-button">
          <BsPersonFillAdd size={30} id="patient-add-icon" />
          Add
        </button>
      </div>

      <div className="patient-table">
        <table id="patient-data-table" border={1}>
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

      <button onClick={returnToHome} className="view-patient-home-button">
        <FaHome size={30} id="view-patient-home-icon" />
        Back to Home
      </button>
    </div>
  );
};

export default ViewPatientPage;
