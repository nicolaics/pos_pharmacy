import React, { useState } from "react";
import Modal from "react-modal";

import "./AdminPasswordPopup.css";

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
    <Modal
      isOpen={isOpen}
      onRequestClose={onClose}
      className={"popup"}
    >
      <h2>Enter your admin password: </h2>
      <div>
        <input
          type="password"
          value={adminPassword}
          onChange={handleAdminPasswordChange}
        />
      </div>
      <button onClick={handleSubmit} id="submit-btn">
        Ok
      </button>
      <button onClick={onClose} id="close-btn">
        &times;
      </button>
    </Modal>
  );
};

export default AdminPasswordPopup;
