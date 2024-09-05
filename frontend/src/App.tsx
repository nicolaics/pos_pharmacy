import React from "react";
import LandingPage from "./components/Home/Home";
import LoginPage from "./components/Login/Login";
import { Routes, Route } from "react-router-dom";

const App: React.FC = () => {
  return (
    <div className="App">      
      <Routes>
        <Route path="/" element={<LoginPage/>} />
        <Route path="/Home" element={<LandingPage/>} />
      </Routes>
    </div>
  );
};

export default App;
