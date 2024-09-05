import React, { useEffect, useState } from "react";
import {
  FaLock,
  FaShoppingCart,
  FaPills,
  FaReceipt,
  FaUser,
  FaClock
} from "react-icons/fa";

import { IoIosLogOut } from "react-icons/io"
import "./Home.css";
import { useNavigate } from "react-router-dom";

const LandingPage: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false); // Track authentication
  const [error, setError] = useState<string | null>(null); // Track error state if needed
  
  const navigate = useNavigate();
  
  const logout = () => {
    navigate("/", { replace: true });
  }

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
        <div className="logout-item" onClick={logout}>
          <IoIosLogOut size={50} />
          <h2>Logout</h2>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
