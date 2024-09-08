import React, {
  SyntheticEvent,
  useEffect,
  useLayoutEffect,
  useState,
} from "react";
import {
  FaLock,
  FaShoppingCart,
  FaPills,
  FaReceipt,
  FaUser,
  FaClock,
} from "react-icons/fa";
import { FaUserDoctor } from "react-icons/fa6";
import { IoIosLogOut } from "react-icons/io";
import { MdSick } from "react-icons/md";

import "./Home.css";
import { Navigate, NavigateFunction, useNavigate } from "react-router-dom";
import { ApplyMiddleware, AuthMiddleware, RequestContext } from "../../Middleware";

// TODO: think about back navigation
// TODO: if the user goes back, reverify the token in the local storage
// TODO: to the URL
const LandingPage: React.FC = () => {
  const navigate = useNavigate();

  const logout = () => {
    const token = sessionStorage.getItem("token");
    const logoutURL = "http://localhost:19230/api/v1/user/logout";

    const requestContext: RequestContext = {
      url: logoutURL,
      options: {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer " + token,
        },
      },
    };

    console.log("token ", token);

    ApplyMiddleware(AuthMiddleware, requestContext).then((response) =>
      response
        .json()
        .then((data) => {
          if (!response.ok) {
            alert("token invalid");
            throw response;
          }

          console.log(data);
          sessionStorage.removeItem("token");
          navigate("/", { replace: true });
        })
        .catch((error) => {
          console.error("Error logging out:", error);
        })
    );
  };

  const user = () => {
    navigate("/user")
  };


  return (
    <div className="landing-page">
      <h1>Welcome!</h1>
      <div className="grid-container">
        <div className="grid-item" onClick={user}>
          <FaLock size={50} />
          <h2>User</h2>
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
          <MdSick size={50} />
          <h2>Patient</h2>
        </div>
        <div className="grid-item">
          <FaUserDoctor size={50} />
          <h2>Doctor</h2>
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
        <div className="grid-item" onClick={logout}>
          <IoIosLogOut size={50} />
          <h2>Logout</h2>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
