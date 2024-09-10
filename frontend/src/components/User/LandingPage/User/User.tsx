import React from "react";
import {
  FaHome,
  FaSearch,
} from "react-icons/fa";
import { BsPersonFillAdd, BsPersonFillGear } from "react-icons/bs";

import "./User.css";
import { useNavigate } from "react-router-dom";


// use window.location.href if the files have been moved to the server

const UserLandingPage: React.FC = () => {
  const navigate = useNavigate();

  const view = () => {
    navigate("/user/view");
  };

  const modifyMyData = () => {
    navigate("/user/detail", {state: 
      {reqType: "modify"}
    });
  };

  const createUser = () => {
    navigate("/user/detail", {state: 
      {reqType: "add"}
    });
  }

  const returnToHome = () => {
    navigate("/home");
  };

  return (
    <div className="user-landing-page">
      <h1>User</h1>
      <div className="user-grid-container">
        <div className="user-grid-item" onClick={modifyMyData}>
          <BsPersonFillGear size={50} />
          <h2>Modify My Information</h2>
        </div>
        <div className="user-grid-item" onClick={view}>
          <FaSearch size={50} />
          <h2>View All Users</h2>
        </div>
        <div className="user-grid-item" onClick={createUser}>
          <BsPersonFillAdd size={50} />
          <h2>Create User</h2>
        </div>
        <div className="user-grid-item" onClick={returnToHome}>
          <FaHome size={50} />
          <h2>Back to Home</h2>
        </div>
      </div>
    </div>
  );
};

export default UserLandingPage;
