import React from "react";
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
import { GiMedicines } from "react-icons/gi";

import "./Home.css";
import { useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../App";

// use window.location.href if the files have been moved to the server

const HomePage: React.FC = () => {
  const navigate = useNavigate();

  const logout = () => {
    const token = sessionStorage.getItem("token");
    const logoutURL = `http://${BACKEND_BASE_URL}/user/logout`;

    fetch(logoutURL, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    })
      .then((response) =>
        response.json().then((data) => {
          if (!response.ok) {
            throw new Error("Invalid credentials or network issue");
          }
          console.log(data);
          sessionStorage.removeItem("token");
          navigate("/");
        })
      )
      .catch((error) => {
        console.error("Error during logout:", error);
      });
  };

  const user = () => {
    navigate("/user");
  };

  const customer = () => {
    navigate("/customer");
  };

  const supplier = () => {
    navigate("/supplier");
  };

  const patient = () => {
    navigate("/patient");
  };

  const doctor = () => {
    navigate("/doctor");
  };

  return (
    <div className="home-page">
      <h1>Welcome!</h1>
      <div className="home-grid-container">
        <div className="home-grid-item" onClick={user}>
          <FaLock size={50} />
          <h2>User</h2>
        </div>
        <div className="home-grid-item" onClick={customer}>
          <FaUser size={50} />
          <h2>Customer</h2>
        </div>
        <div className="home-grid-item" onClick={supplier}>
          <FaShoppingCart size={50} />
          <h2>Supplier</h2>
        </div>
        <div className="home-grid-item" onClick={patient}>
          <MdSick size={50} />
          <h2>Patient</h2>
        </div>
        <div className="home-grid-item" onClick={doctor}>
          <FaUserDoctor size={50} />
          <h2>Doctor</h2>
        </div>
        <div className="home-grid-item">
          <FaPills size={50} />
          <h2>Inventory</h2>
        </div>
        <div className="home-grid-item">
          <FaReceipt size={50} />
          <h2>Invoice</h2>
        </div>
        <div className="home-grid-item">
          <FaClock size={50} />
          <h2>Purchasing</h2>
        </div>
        <div className="home-grid-item">
          <GiMedicines size={50} />
          <h2>Production</h2>
        </div>
        <div className="home-grid-item-logout" onClick={logout}>
          <IoIosLogOut size={50} />
          <h2>Logout</h2>
        </div>
      </div>
    </div>
  );
};

export default HomePage;
