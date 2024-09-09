import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import AdminPasswordPopup from "../../AdminPasswordPopup/AdminPasswordPopup";

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

  const reqType = state.reqType;

  const [isAdminPopupOpen, setAdminPopup] = useState(false);
  const [showAddBtn, setshowAddBtn] = useState(false);

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state.reqType === "modify") {
      setOkBtnLabel("Modify");

      if (state.admin) {
        setShowAdminOption(true);
      } else {
        setShowAdminOption(false);
      }

      const currentUserURL ="http://localhost:19230/api/v1/user/current"; // Set the URL or handle this logic
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
      const url = "http://localhost:19230/api/v1/user/modify";

      fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          id: state.id,
          newData: {
            name: name,
            password: password,
            phoneNumber: phoneNumber,
            admin: admin,
            adminPassword: adminPassword
          }
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
    }

    // Reset the state
    setshowAddBtn(false);
    setAdminPassword("");
    navigate("/user");
  };

  const handleCancel = (e: any) => {
    navigate("/user");
  };

  return (
    <div className="modify-user-page">
      <h1>Add User</h1>

      <div className="user-data">
        <div className="id-number">
          <h3>ID: </h3>
          <label>{id}</label>
        </div>
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

        {showAdminOption && (
          <div className="admin">
            <h3>Admin: </h3>
            <input
              type="radio"
              id="radio-yes"
              checked={admin == true ? true : false}
              name="admin"
              value={"yes"}
              onChange={handleAdminChange}
            />
            <label htmlFor="radio-yes">Yes</label>
            <input
              type="radio"
              checked={admin == false ? true : false}
              id="radio-no"
              name="admin"
              value={"no"}
              onChange={handleAdminChange}
            />
            <label htmlFor="radio-no">No</label>
          </div>
        )}
      </div>

      <div className="buttons">
        <div className="ok-cancel-btn">
          <button type="submit" onSubmit={handleRequestAdminPassword}>
            {okBtnLabel}
          </button>
          <button type="button" onClick={handleCancel}>
            Cancel
          </button>
        </div>

        <div className="final-ok-btn">
          {/* Final submit button, shown only after popup is closed */}
          {showAddBtn && (
            <button type="button" onClick={handleSendRequest}>
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
