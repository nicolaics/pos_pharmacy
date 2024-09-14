import React, { useState } from "react";
import Modal from "react-modal";

import "./AdminPasswordPopup.css";
import { FaCheckCircle } from "react-icons/fa";

// Set the root element for the modal (for accessibility reasons)
Modal.setAppElement("#root");

interface ModalComponentProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (password: string) => void; // Callback for when the user submits the form
}

const AdminPasswordPopup: React.FC<ModalComponentProps> = ({
  isOpen,
  onClose,
  onSubmit,
}) => {
  const [adminPassword, setAdminPassword] = useState("");

  if (!isOpen) {
    return null;
  }

  const handleAdminPasswordChange = (event: any) => {
    setAdminPassword(event.target.value);
  };

  const handleSubmit = () => {
    onSubmit(adminPassword); // Pass the password back to the parent component
    setAdminPassword(""); // Clear the input field after submission
    onClose(); // Close the popup
  };

  return (
    <Modal isOpen={isOpen} onRequestClose={onClose} className={"popup"}>
      <h2>Enter your admin password: </h2>
      <div className="admin-password-container">
        <input
          type="password"
          value={adminPassword}
          onChange={handleAdminPasswordChange}
          id="admin-password-input"
        />

          <button onClick={handleSubmit} className="admin-password-ok-btn">
            <FaCheckCircle size={20} id="admin-password-ok-icon"/>
            Ok
          </button>

        <button onClick={onClose} className="close-popup-btn">
          &times;
        </button>
      </div>
    </Modal>
  );
};

export default AdminPasswordPopup;
