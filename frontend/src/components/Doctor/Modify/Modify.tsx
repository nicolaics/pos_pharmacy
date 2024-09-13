import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../../App";

const ModifyDoctorPage: React.FC = () => {
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

  /*
  useEffect(() => {
    const token = sessionStorage.getItem("token");

    if (state) {
      setOkBtnLabel("Modify");
      setShowIdField(true);

      // TODO: add query for searching doctors
      const doctorURL = `http://${BACKEND_BASE_URL}/doctor?all=all`; // Set the URL or handle this logic
      fetch(doctorURL, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify doctor data");
            }

            console.log(data["data"]);

            setId(data["data"].id);
            setName(data["data"].name);
          })
        )
        .catch((error) => {
          console.error("Error load selected doctor:", error);
          alert("Error load selected doctor");
        });
    } else {
      setOkBtnLabel("Add");
      setShowIdField(false);
    }
  }, [state]); // Dependency array ensures this effect only runs when reqType changes
  */


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
      const url = `http://${BACKEND_BASE_URL}/doctor`;

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
              throw new Error("Unable to modify doctor data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify doctor:", error);
          alert("Error modify doctor");
        });
    } else {
      const url = `http://${BACKEND_BASE_URL}/doctor`;

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
              throw new Error("Invalid c redentials or network issue");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error adding new doctor:", error);
          alert("Error adding new doctor");
        });
    }

    // Reset the state
    navigate("/doctor");
  };

  const handleCancel = (e: any) => {
    navigate("/doctor");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();

    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/doctor`;

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
            throw new Error("Unable to delete doctor data");
          }

          console.log(data);
          navigate("/doctor");
        })
      )
      .catch((error) => {
        console.error("Error delete doctor:", error);
        alert("Error delete doctor");
      });
  };

  return (
    <div className="modify-doctor-page">
      <h1>{heading} Doctor</h1>

      <div className="doctor-data-container">
        { showIdField && (
        <div className="doctor-data-form-group">
          <label htmlFor="id">ID:</label>
          <input type="text" id="modify-doctor-id" value={id} readOnly />
        </div>
        )}
        <div className="doctor-data-form-group">
          <label htmlFor="name">Name:</label>
          <input
            type="text"
            id="modify-doctor-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>
      </div>

      <div className="modify-doctor-buttons">
        <div className="modify-doctor-btns-grp">
        <button
            type="button"
            className="modify-doctor-delete-btn"
            onClick={handleDelete}
          >
            Delete Doctor
          </button>

          <button
            type="button"
            className="modify-doctor-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="button"
            className="modify-doctor-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifyDoctorPage;
