import React, { useState } from "react";
import "./Login.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import { BACKEND_BASE_URL } from "../../App";


// const name = 'admin1';
// const password = 'dnP9K5RMjV1l';

function login(name: string, password: string, navigate: NavigateFunction) {
  const url = `http://${BACKEND_BASE_URL}/user/login`;
  
  fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name: name,
      password: password,
    }),
  })
    .then((response) =>
      response.json().then((data) => {
        if (!response.ok) {
          throw new Error("Invalid credentials or network issue");
        }
        console.log(data);
        sessionStorage.setItem("token", data["token"]);
        navigate("/home");
      })
    )
    .catch((error) => {
      console.error("Error during sign-in:", error);
      alert("Invalid credentials:" + error); // Show pop-up message
    });
}

const LoginPage: React.FC = () => {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");

  const navigate = useNavigate();

  const handleNameChange = (event: any) => {
    setName(event.target.value);
  };

  const handlePasswordChange = (event: any) => {
    setPassword(event.target.value);
  };

  const handleSubmit = (event: any) => {
    event.preventDefault();

    if (name === "" || password === "") {
      alert(`name and password cannot be empty`);
      return;
    }

    login(name, password, navigate);

    setName("");
    setPassword("");
  };

  return (
    <div className="login-card">
      <h1>Login to POS System</h1>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Name"
          value={name}
          onChange={handleNameChange}
          className="login-form-input"
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={handlePasswordChange}
          className="login-form-input"
        />
        <button type="submit" className="login-btn">
          Login
        </button>
      </form>
    </div>
  );
};

export default LoginPage;
