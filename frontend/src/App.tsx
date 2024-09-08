import React from "react";
import LandingPage from "./components/Home/Home";
import LoginPage from "./components/Login/Login";
import { Routes, Route } from "react-router-dom";
import UserPage from "./components/User/User";

const App: React.FC = () => {
  return (
    <div className="App">      
      <Routes>
        <Route path="/" element={<LoginPage/>} />
        <Route path="/home" element={<LandingPage/>} />
        <Route path="/user" element={<UserPage/>} />
      </Routes>
    </div>
  );
};

export default App;
