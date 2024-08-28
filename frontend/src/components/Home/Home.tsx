import React from "react";
import {
  FaLock,
  FaShoppingCart,
  FaPills,
  FaReceipt,
  FaUser,
  FaClock
} from "react-icons/fa";
import "./Home.css";

const LandingPage: React.FC = () => {
  return (
    <div className="landing-page">
      <h1>Welcome!</h1>
      <div className="grid-container">
        <div className="grid-item">
          <FaLock size={50} />
          <h2>Admin</h2>
        </div>
        <div className="grid-item">
          <FaUser size={50} />
          <h2>Customer</h2>
        </div>
        <div className="grid-item">
          <FaShoppingCart size={50} />
          <h2>Supplier</h2>
        </div>
        <div className="grid-item">
          <FaPills size={50} />
          <h2>Inventory</h2>
        </div>
        <div className="grid-item">
          <FaReceipt size={50} />
          <h2>Invoice</h2>
        </div>
        <div className="grid-item">
          <FaClock size={50} />
          <h2>Purchasing</h2>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
