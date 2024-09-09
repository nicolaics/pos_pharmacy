import React, { useState } from "react";

import "./User.css";
import { useNavigate } from "react-router-dom";
import FormatDateTime from "../../../DateTimeFormatter";
import AdminPasswordPopup from "../../AdminPasswordPopup/AdminPasswordPopup";

// TODO: notify user if they have selected something (highlight or something in css)
function fillTable(
  data: any,
  tableBody: Element | null,
  setName: React.Dispatch<React.SetStateAction<string>>,
  setID: React.Dispatch<React.SetStateAction<number>>
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
    setName(data["name"]);
    setID(data["id"]);
  });

  // Append the row to the table body
  tableBody.appendChild(row);
}

const UserPage: React.FC = () => {
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [id, setID] = useState(-1);
  const [adminPassword, setAdminPassword] = useState("");

  const [isAdminPopupOpen, setAdminPopup] = useState(false);

  const openPopup = (e: any) => {
    e.preventDefault(); // Prevent form submission
    setAdminPopup(true); // Open the popup
  };

  const closePopup = () => {
    setAdminPopup(false);
  };

  const handlePopupSubmit = (adminPasswordRcv: string) => {
    setAdminPassword(adminPasswordRcv); // Save the password received from the popup
    setAdminPopup(false); // Close the popup
  };

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
            fillTable(data[i], tableBody, setName, setID);
          }
        })
      )
      .catch((error) => {
        console.error("Error loading user data:", error);
        alert("Error loading user data");
      });
  };

  const add = () => {
    navigate("/user/add");
  };

  // TODO: TEST, not yet tested
  const remove = (e: any) => {
    const confirmText =
      "Are you sure you want to delete:\n" + "id: " + id + "\nname: " + name;
    const confirm = window.confirm(confirmText);

    if (confirm) {
      e.preventDefault();
      openPopup(e);

      const token = sessionStorage.getItem("token");
      const deleteURL = "http://localhost:19230/api/v1/user/delete";

      fetch(deleteURL, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: id,
          name: name,
          adminPassword: adminPassword,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              setAdminPassword("");
              throw new Error("Invalid credentials or network issue");
            }
          })
        )
        .catch((error) => {
          setAdminPassword("");
          console.error("Error deleting user:", error);
          alert("Error deleting user");
        });
    }

    setAdminPassword("");
  };

  return (
    <div className="user-page">
      <h1>User</h1>

      <div className="container">
        <span className="search">
          <input type="text" className="search-item" />
          <button className="search-item" onClick={search}>
            Search
          </button>
          <button className="search-item" onClick={add}>
            +
          </button>
          <button className="search-item" onClick={remove}>
            -
          </button>
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

      <AdminPasswordPopup
        isOpen={isAdminPopupOpen}
        onClose={closePopup}
        onSubmit={handlePopupSubmit}
      />
    </div>
  );
};

export default UserPage;
