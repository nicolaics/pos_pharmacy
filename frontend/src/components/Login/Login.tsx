import React, { useState } from "react";
import "./Login.css";
import { NavigateFunction, useNavigate } from "react-router-dom";
import {
  ApplyMiddleware,
  AuthMiddleware,
  RequestContext,
} from "../../Middleware";

// so we will define our constants here
const url = "http://localhost:19230/api/v1/user/login";

// const username = 'admin1';
// const password = 'dnP9K5RMjV1l';

function login(username: string, password: string, navigate: NavigateFunction) {
  const requestContext: RequestContext = {
    url: url,
    options: {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: username,
        password: password,
      }),
    },
  };

  ApplyMiddleware(AuthMiddleware, requestContext).then((response) =>
    response
      .json()
      .then((data) => {
        if (!response.ok) {
          alert("wrong credentials");
          throw response;
        }

        console.log(data);
        sessionStorage.setItem("token", data["token"]);
        navigate("/home");
      })
      .catch((error) => {
        console.error("Error fetching user data:", error);
      })
  );
}

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const navigate = useNavigate();

  const handleUsernameChange = (event: any) => {
    setUsername(event.target.value);
  };

  const handlePasswordChange = (event: any) => {
    setPassword(event.target.value);
  };

  // TODO: fetch from API
  const handleSubmit = (event: any) => {
    event.preventDefault();

    if (username === "" || password === "") {
      alert(`username and password cannot be empty`);
      return;
    }

    login(username, password, navigate);

    setUsername("");
    setPassword("");
  };

  return (
    <div className="login-grid">
      <h1>Login</h1>
      <span className="login-form">
        <form onSubmit={handleSubmit}>
          <label id="username">Username: </label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={handleUsernameChange}
          />
          <br />

          <label id="password">Password: </label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={handlePasswordChange}
          />
          <br />

          <input type="submit" value={"Login"} />
        </form>
      </span>
    </div>
  );
};

export default LoginPage;
