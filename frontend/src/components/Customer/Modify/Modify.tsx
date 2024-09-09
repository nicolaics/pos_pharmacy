import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";

const ModifyCustomerPage: React.FC = () => {
  const navigate = useNavigate();
  const state = useLocation().state;

  const [id, setId] = useState("");
  const [name, setName] = useState("");

  const [okBtnLabel, setOkBtnLabel] = useState("Modify");

  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state) {
      setOkBtnLabel("Modify");
      
      // TODO: add query for searching customers
      const customerURL = "http://localhost:19230/api/v1/customer?id="; // Set the URL or handle this logic
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

            console.log(data["data"]);

            setId(data["data"].id);
            setName(data["data"].name);
          })
        )
        .catch((error) => {
          console.error("Error load selected customer:", error);
          alert("Error load selected customer");
        });
    } else {
      setOkBtnLabel("Add");
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
      const url = "http://localhost:19230/api/v1/customer";

      fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({
          name: name,
        }),
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify customer data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify customer:", error);
          alert("Error modify customer");
        });
    } else {
      const url = "http://localhost:19230/api/v1/customer";

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
              throw new Error("Invalid credentials or network issue");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error adding new customer:", error);
          alert("Error adding new customer");
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
    const url = "http://localhost:19230/api/v1/customer";

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
      <h1>Add User</h1>

      <div className="customer-data">
        <div className="id-number">
          <h3>ID: </h3>
          <input type="text" value={id} readOnly />
        </div>

        <div className="customer-name">
          <h3>Name: </h3>
          <input type="text" value={name} onChange={handleNameChange} />
        </div>
      </div>

      <div className="customer-ok-cancel-delete-btn">
        <button type="button" onClick={handleSendRequest}>
          {okBtnLabel}
        </button>

        <button type="button" onClick={handleCancel}>
          Cancel
        </button>

        <button type="button" onClick={handleDelete}>
          Delete User
        </button>
      </div>
    </div>
  );
};

export default ModifyCustomerPage;
