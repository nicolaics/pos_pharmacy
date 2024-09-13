import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../../App";

const ModifyCustomerPage: React.FC = () => {
  const navigate = useNavigate();
  const state = useLocation().state;

  const [id, setId] = useState("");
  const [name, setName] = useState("");

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");
  const [showIdField, setShowIdField] = useState(false);

  var heading = "";
  if (state) {
    heading = "Modify";
  } else {
    heading = "Add";
  }

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state) {
      setOkBtnLabel("Modify");
      setShowIdField(true);

      const customerURL = `http://${BACKEND_BASE_URL}/customer/${state.id}`; // Set the URL or handle this logic
      fetch(customerURL, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify customer data");
            }

            console.log(data[0]);

            setId(data[0]["data"].id);
            setName(data[0]["data"].name);
          })
        )
        .catch((error) => {
          console.error("Error load selected customer:", error);
          alert("Error load selected customer");
        });
    } else {
      setOkBtnLabel("Add");
      setShowIdField(false);
    }
  }, [state]); // Dependency array ensures this effect only runs when reqType changes

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handleSendRequest = (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault(); // Prevent form submission

    // Handle form submission logic here
    const token = sessionStorage.getItem("token");

    if (state) {
      const url = `http://${BACKEND_BASE_URL}/customer`;

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
          },
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to add customer data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error add customer:", error);
          alert("Error add customer");
        });
    } else {
      const url = `http://${BACKEND_BASE_URL}/customer`;

      fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          name: name
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
          console.error("Error add customer:", error);
          alert("Error add customer");
        });
    }

    // Reset the state
    navigate("/customer");
  };

  const handleCancel = (e: any) => {
    navigate("/customer");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();

    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/customer`;

    fetch(url, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        id: Number(id),
        name: name,
      }),
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Unable to delete customer data");
          }

          console.log(data);
          navigate("/customer");
        })
      )
      .catch((error) => {
        console.error("Error delete customer:", error);
        alert("Error delete customer");
      });
  };

  return (
    <div className="modify-customer-page">
      <h1>{heading} Customer</h1>

      <div className="customer-data-container">
        {showIdField && (
        <div className="customer-data-form-group">
          <label htmlFor="id">ID:</label>
          <input type="text" id="modify-customer-id" value={id} readOnly />
        </div>
        )}
        <div className="customer-data-form-group">
          <label htmlFor="name">Name:</label>
          <input
            type="text"
            id="modify-customer-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>
      </div>

      <div className="modify-customer-buttons">
        <div className="modify-customer-btns-grp">
        <button
            type="button"
            className="modify-customer-delete-btn"
            onClick={handleDelete}
          >
            Delete Customer
          </button>

          <button
            type="button"
            className="modify-customer-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="button"
            className="modify-customer-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifyCustomerPage;
