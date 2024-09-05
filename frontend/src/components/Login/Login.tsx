import React, { useState } from "react";
import "./Login.css";
import { useNavigate } from "react-router-dom";
import LandingPage from "../Home/Home";

// so we will define our constants here
const url = 'http://localhost:19230/api/v1/cashier/login';

const username = 'admin';
const password = 'admin1234!';

// now we configure the fetch request to the url endpoint.
// we should probably put it inside a separate function since
// you're using a browser, you probably will bind this request
// to a click event or something.
async function login() {
  const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            name: username,
            password: password,
        }),
    });
    // An important thing to note is that an error response will not throw
    // an error so if the result is not okay we should throw the error
    if (!response.ok) {
        throw response;
    }
    return await response.json();
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

    const handleSubmit = (event: any) => {
        event.preventDefault();

        if (username === "" || password === "") {
            alert(`username and password cannot be empty`);
            return;
        }

        // alert(`You typed ${username}\n ${password}`);
        // console.log(login());
        // setUsername("");
        // setPassword("");
        navigate("/Home");
    };

    return (
        <div className="login-grid">
            <h1>Login</h1>
            <span className="login-form">
                <form onSubmit={handleSubmit}>
                    <label id="username">Username: </label>
                    <input type="text" id="username" value={username} onChange={handleUsernameChange}/>
                    <br />
                        
                    <label id="password">Password: </label>
                    <input type="password" id="password" value={password} onChange={handlePasswordChange} />
                    <br />
                    
                    <input type="submit" value={"Login"} />
                </form>
            </span>
        </div>
    );
};

export default LoginPage;