import React, { useEffect, useState } from "react";

import "./Modify.css";
import { useLocation, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../../App";

const ModifyPatientPage: React.FC = () => {
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

      const patientURL = `http://${BACKEND_BASE_URL}/patient/${state.id}`; // Set the URL or handle this logic
      fetch(patientURL, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      })
        .then((response) =>
          response.json().then((data) => {
            if (!response.ok) {
              throw new Error("Unable to modify patient data");
            }

            console.log(data["data"]);

            setId(data[0]["data"].id);
            setName(data[0]["data"].name);
          })
        )
        .catch((error) => {
          console.error("Error load selected patient:", error);
          alert("Error load selected patient");
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
      const url = `http://${BACKEND_BASE_URL}/patient`;

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
              throw new Error("Unable to modify patient data");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error modify patient:", error);
          alert("Error modify patient");
        });
    } else {
      const url = `http://${BACKEND_BASE_URL}/patient`;

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
              throw new Error("Invalid credentials or network issue");
            }

            console.log(data);
          })
        )
        .catch((error) => {
          console.error("Error adding new patient:", error);
          alert("Error adding new patient");
        });
    }

    // Reset the state
    navigate("/patient");
  };

  const handleCancel = (e: any) => {
    navigate("/patient");
  };

  const handleDelete = (e: any) => {
    e.preventDefault();

    const token = sessionStorage.getItem("token");
    const url = `http://${BACKEND_BASE_URL}/patient`;

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
            throw new Error("Unable to delete patient data");
          }

          console.log(data);
          navigate("/patient");
        })
      )
      .catch((error) => {
        console.error("Error delete patient:", error);
        alert("Error delete patient");
      });
  };

  return (
    <div className="modify-patient-page">
      <h1>{heading} Patient</h1>

      <div className="patient-data-container">
        { showIdField && (
        <div className="patient-data-form-group">
          <label htmlFor="id">ID:</label>
          <input type="text" id="modify-patient-id" value={id} readOnly />
        </div>
        )}
        <div className="patient-data-form-group">
          <label htmlFor="name">Name:</label>
          <input
            type="text"
            id="modify-patient-name"
            value={name}
            onChange={handleNameChange}
          />
        </div>
      </div>

      <div className="modify-patient-buttons">
        <div className="modify-patient-btns-grp">
        <button
            type="button"
            className="modify-patient-delete-btn"
            onClick={handleDelete}
          >
            Delete Patient
          </button>

          <button
            type="button"
            className="modify-patient-cancel-btn"
            onClick={handleCancel}
          >
            Cancel
          </button>

          <button
            type="button"
            className="modify-patient-ok-btn"
            onClick={handleSendRequest}
          >
            {okBtnLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ModifyPatientPage;
