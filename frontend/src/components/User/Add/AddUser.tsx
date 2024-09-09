import React, { useState } from "react";

import "./AddUser.css";
import { useNavigate } from "react-router-dom";
import AdminPasswordPopup from "../../AdminPasswordPopup/AdminPasswordPopup";

const AddUserPage: React.FC = () => {
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [admin, setAdmin] = useState(false);
  const [adminPassword, setAdminPassword] = useState("");

  const [isAdminPopupOpen, setAdminPopup] = useState(false);
  const [showAddBtn, setshowAddBtn] = useState(false);

  const openPopup = (e: any) => {
    e.preventDefault(); // Prevent form submission
    setAdminPopup(true); // Open the popup
    setshowAddBtn(true); // Show the final submit button after the popup is closed
  };

  const closePopup = () => {
    setAdminPopup(false);
  };

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handlePasswordChange = (event: any) => {
    setPassword(event.target.value);
  };

  const handlePhoneNumberChange = (event: any) => {
    setPhoneNumber(event.target.value);
  };

  const handleAdminChange = (event: any) => {
    if (event.target.value == "yes") {
      setAdmin(true);
    } else {
      setAdmin(false);
    }
  };

  const handlePopupSubmit = (adminPasswordRcv: string) => {
    setAdminPassword(adminPasswordRcv); // Save the password received from the popup
    setAdminPopup(false); // Close the popup
  };

  const handleSubmit = (e: any) => {
    e.preventDefault(); // Prevent form submission and page reload
    openPopup(e); // Open the popup
  };

  const handleFinalSubmit = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");
    const url = "http://localhost:19230/api/v1/user/register";

    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        name: name,
        password: password,
        phoneNumber: phoneNumber,
        admin: admin,
        adminPassword: adminPassword,
      }),
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Invalid credentials or network issue");
          }

          console.log(data);
        })
      )
      .catch((error) => {
        console.error("Error adding new user:", error);
        alert("Error adding new user");
      });

    // Reset the state
    setshowAddBtn(false);
    navigate("/user");
  };

  return (
    <div className="add-user-page">
      <h1>Add User</h1>

      <form className="add-user-form" onSubmit={handleSubmit}>
        <div className="name">
          <h3>Name: </h3>
          <input type="text" value={name} onChange={handleNameChange} />
        </div>
        <div className="password">
          <h3>Password: </h3>
          <input
            type="password"
            value={password}
            onChange={handlePasswordChange}
          />
        </div>
        <div className="phone-number">
          <h3>Phone Number: </h3>
          <input
            type="text"
            value={phoneNumber}
            onChange={handlePhoneNumberChange}
          />
        </div>
        <div className="admin">
          <h3>Admin: </h3>
          <input
            type="radio"
            id="radio-yes"
            name="admin"
            value={"yes"}
            onChange={handleAdminChange}
          />
          <label htmlFor="radio-yes">Yes</label>
          <input
            type="radio"
            id="radio-no"
            name="admin"
            value={"no"}
            onChange={handleAdminChange}
          />
          <label htmlFor="radio-no">No</label>
        </div>
        <div className="open-popup-button">
          {/* Initially visible submit button */}
          <button type="submit">Ok</button>
        </div>
        <div className="add-button">
          {/* Final submit button, shown only after popup is closed */}
          {showAddBtn && (
            <button type="button" onClick={handleFinalSubmit}>
              Add
            </button>
          )}
        </div>
      </form>

      <AdminPasswordPopup
        isOpen={isAdminPopupOpen}
        onClose={closePopup}
        onSubmit={handlePopupSubmit}
      />
    </div>
  );
};

export default AddUserPage;
