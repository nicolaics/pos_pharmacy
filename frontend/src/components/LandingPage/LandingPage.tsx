import React from "react";
import { Link, useNavigate } from "react-router-dom";
import "./LandingPage.css";

const LandingPage: React.FC = () => {
  const navigate = useNavigate();

  const goToLogin = () => {
    navigate("/login");
  };

  return (
    <div className="landing-page-card">
      <div className="landing-page-text">
        <h1>Welcome to Our POS System</h1>
        <p>
          Streamline your business operations with our easy-to-use Point of Sale
          system.
        </p>
      </div>
      <button className="go-to-login-btn" onClick={goToLogin}>
        Go to Login
      </button>
    </div>
  );
};

export default LandingPage;
