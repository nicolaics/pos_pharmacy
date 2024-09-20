import React, { useEffect, useState } from "react";

import "./Doctor.css";
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

  // Loop through each doctor and create a new row in the table
  const row = document.createElement("tr");
  row.className = "view-doctor-data-row";

  // Create and append cells for each column
  const idCell = document.createElement("td");
  idCell.textContent = data["id"].toString();
  idCell.className = "doctor-id-column";
  row.appendChild(idCell);

  const nameCell = document.createElement("td");
  nameCell.textContent = data["name"];
  nameCell.className = "doctor-name-column";
  row.appendChild(nameCell);

  const createdAt = new Date(data["createdAt"]);
  const createdAtCell = document.createElement("td");
  createdAtCell.textContent = FormatDateTime(createdAt);
  createdAtCell.className = "doctor-created-at-column";
  row.appendChild(createdAtCell);

  row.addEventListener("dblclick", () => {
    navigate("/doctor/detail", {
      state: {
        id: data["id"],
        name: data["name"],
      },
    });
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const ViewDoctorPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchVal, setSearchVal] = useState("");

  const handleSearchValChange = (event: any) => {
    event.preventDefault();
    setSearchVal(event.target.value);
  };

  /*
  useEffect(() => {
    const token = sessionStorage.getItem("token");
    const getAllDoctorURL = `http://${BACKEND_BASE_URL}/doctor/all`;
    
    fetch(getAllDoctorURL, {
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

          const tableBody = document.querySelector("#doctor-data-table tbody");
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
        console.error("Error loading doctor data:", error);
        alert("Error loading doctor data");
      });
  });
  */

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
    var getDoctorURL = "";
    
    // TEST DATA
    const tableBody = document.querySelector("#doctor-data-table tbody");
    if (!tableBody) {
      console.error("table body not found");
      return;
    }
    tableBody.innerHTML = "";
    for (let i = 0; i < testData.length; i++) {
      fillTable(testData[i], tableBody, navigate);
    }

    // if (searchVal === "") {
    //   getDoctorURL = `http://${BACKEND_BASE_URL}/doctor/all`;
    // }
    // else {
    //   getDoctorURL = `http://${BACKEND_BASE_URL}/doctor/${searchVal}`;
    // }

    // fetch(getDoctorURL, {
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

    //       const tableBody = document.querySelector("#doctor-data-table tbody");
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
    //     console.error("Error loading doctor data:", error);
    //     alert("Error loading doctor data");
    //   });
  };

  const returnToHome = () => {
    navigate("/home");
  };

  const register = () => {
    navigate("/doctor/detail");
  };

  return (
    <div className="view-doctor-page">
      <h1>Doctor</h1>
      
      <div className="doctor-container">
        <input
          type="text"
          className="doctor-search-input"
          placeholder="Search"
          value={searchVal}
          onChange={handleSearchValChange}
        />
        <button onClick={search} className="doctor-search-button">
          <MdPersonSearch size={30} id="doctor-search-icon" />
          Search
        </button>

        <button onClick={register} className="doctor-add-button">
          <BsPersonFillAdd size={30} id="doctor-add-icon" />
          Add
        </button>
      </div>

      <div className="doctor-table-container">
        <table id="doctor-data-table" border={1}>
          <thead>
            <tr>
              <th className="doctor-id-column">ID</th>
              <th className="doctor-name-column">Name</th>
              <th className="doctor-created-at-column">Created At</th>
            </tr>
          </thead>
          <tbody></tbody>
        </table>
      </div>

      <button onClick={returnToHome} className="view-doctor-home-button">
        <FaHome size={30} id="view-doctor-home-icon" />
        Back to Home
      </button>
    </div>
  );
};

export default ViewDoctorPage;
