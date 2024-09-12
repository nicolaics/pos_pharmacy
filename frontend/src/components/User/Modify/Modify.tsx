import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import AdminPasswordPopup from "../../AdminPasswordPopup/AdminPasswordPopup";
import { BACKEND_BASE_URL } from "../../../App";

// use window.location.href if the files have been moved to the server

const ModifyUserPage: React.FC = () => {
  const navigate = useNavigate();
  const state = useLocation().state;

  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [admin, setAdmin] = useState(false);
  const [adminPassword, setAdminPassword] = useState("");

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");
  const [showAdminOption, setShowAdminOption] = useState(false);
  const [showDeleteButton, setShowDeleteButton] = useState(false);
  const [isReadOnly, setIsReadOnly] = useState(false);

  const [isAdminPopupOpen, setAdminPopup] = useState(false);
  const [showAddBtn, setshowAddBtn] = useState(false);
  const [showIdField, setShowIdField] = useState(false);

  const reqType = state.reqType;

  var heading = "";
  if (reqType == "add") {
    heading = "Add";
  } else {
    heading = "Modify";
  }

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state.reqType === "modify") {
      setOkBtnLabel("Modify");
      setShowAdminOption(false);
      setIsReadOnly(false);
      setShowDeleteButton(false);
      setShowIdField(true);

      const currentUserURL = `http://${BACKEND_BASE_URL}/user/current`; // Set the URL or handle this logic
      fetch(currentUserURL, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify user data");
            }

            console.log(data["data"]);
            setId(data["data"].id);
            setName(data["data"].name);
            setPassword(data["data"].password);
            setPhoneNumber(data["data"].phoneNumber);
            setAdmin(data["data"].admin);
          })
        )
        .catch((error) => {
          console.error("Error load current user:", error);
          alert("Error load current user");
        });
    } else if (state.reqType === "add") {
      setOkBtnLabel("Add");
      setShowAdminOption(true);
      setShowDeleteButton(false);
      setShowIdField(false);
    } else if (state.reqType === "modify-admin") {
      setIsReadOnly(true);
      setShowAdminOption(true);
      setShowDeleteButton(true);
      setShowIdField(true);

      setId(state.id);
      setName(state.name);
    }
  }, [state.reqType]); // Dependency array ensures this effect only runs when reqType changes

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

  const handleGetAdminPassword = (adminPasswordRcv: string) => {
    setAdminPassword(adminPasswordRcv); // Save the password received from the popup
    setAdminPopup(false); // Close the popup
  };

  const handleRequestAdminPassword = (e: any) => {
    e.preventDefault(); // Prevent form submission and page reload
    openPopup(e); // Open the popup
  };

  const handleSendRequest = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");

    if (reqType === "modify") {
      const url = `http://${BACKEND_BASE_URL}/user/modify`;

      fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: Number(id),
          newData: {
            name: name,
            password: password,
            phoneNumber: phoneNumber,
            admin: admin,
            adminPassword: adminPassword,
          },
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify user data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify user:", error);
          alert("Error modify user");
        });
    } else if (reqType === "add") {
      const url = `http://${BACKEND_BASE_URL}/user/register`;

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
    } else if (reqType === "modify-admin") {
      const url = `http://${BACKEND_BASE_URL}/user/admin`;

      fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: Number(id),
          admin: admin,
          adminPassword: adminPassword,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify user data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify user:", error);
          alert("Error modify user");
        });
    }

    // Reset the state
    setshowAddBtn(false);
    setAdminPassword("");
    navigate("/user");
  };

  const handleCancel = (e: any) => {
    navigate("/user");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();
    handleRequestAdminPassword(e);

    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/user/delete`;

    fetch(url, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        id: Number(id),
        name: name,
        adminPassword: adminPassword,
      }),
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Unable to delete user data");
          }

          console.log(data);
          setAdminPassword("");
          navigate("/user");
        })
      )
      .catch((error) => {
        console.error("Error delete user:", error);
        alert("Error delete user");
      });

    setAdminPassword("");
  };

  return (
    <div className="modify-user-page">
      <h1>{heading} User</h1>

      <div className="user-data-container">
        {showIdField && (
          <div className="user-data-form-group">
            <label htmlFor="id">ID:</label>
            <input type="text" id="modify-user-id" value={id} readOnly />
          </div>
        )}
        <div className="user-data-form-group">
          <label htmlFor="name">Name:</label>
          <input
            type="text"
            id="modify-user-name"
            value={name}
            onChange={handleNameChange}
            readOnly={isReadOnly}
          />
        </div>

        <div className="user-data-form-group">
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="modify-user-password"
            value={password}
            onChange={handlePasswordChange}
            readOnly={isReadOnly}
          />
        </div>

        <div className="user-data-form-group">
          <label htmlFor="phoneNumber">Phone Number:</label>
          <input
            type="text"
            id="modify-user-phone-number"
            value={phoneNumber}
            onChange={handlePhoneNumberChange}
            readOnly={isReadOnly}
          />
        </div>

        {showAdminOption && (
          <div className="user-data-form-group">
            <label>Admin:</label>
            <div className="user-data-radio-grp">
              <input
                type="radio"
                id="modify-user-radio-yes"
                checked={admin === true}
                name="admin"
                value={"yes"}
                onChange={handleAdminChange}
              />
              <label htmlFor="radio-yes">Yes</label>
              <input
                type="radio"
                checked={admin === false}
                id="modify-user-radio-no"
                name="admin"
                value={"no"}
                onChange={handleAdminChange}
              />
              <label htmlFor="radio-no">No</label>
            </div>
          </div>
        )}
      </div>

      <div className="modify-user-buttons">
        <div className="modify-user-btns-grp">
          <button
            type="submit"
            className="modify-user-ok-btn"
            onSubmit={handleRequestAdminPassword}
          >
            {okBtnLabel}
          </button>
          <button
            type="button"
            className="modify-user-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          {showDeleteButton && (
            <button
              type="button"
              className="modify-user-delete-btn"
              onClick={handleDelete}
            >
              Delete User
            </button>
          )}
        </div>

        <div className="modify-user-btns-grp">
          {/* Final submit button, shown only after popup is closed */}
          {showAddBtn && (
            <button
              type="button"
              className="modify-user-send-btn"
              onClick={handleSendRequest}
            >
              Send Request
            </button>
          )}
        </div>
      </div>

      <AdminPasswordPopup
        isOpen={isAdminPopupOpen}
        onClose={closePopup}
        onSubmit={handleGetAdminPassword}
      />
    </div>
  );
};

export default ModifyUserPage;
