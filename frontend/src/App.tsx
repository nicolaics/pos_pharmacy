import React from "react";
import LandingPage from "./components/Home/Home";
import LoginPage from "./components/Login/Login";
import { Routes, Route } from "react-router-dom";
import UserPage from "./components/User/User";
import ProtectedRoute from "./ProtectedRoute";
import AddUserPage from "./components/User/AddUser";

const App: React.FC = () => {
  return (
    <div className="App">
      <Routes>
        <Route path="/" element={<LoginPage />} />

        <Route
          path="/home"
          element={<ProtectedRoute children={<LandingPage />} admin={false} />}
        />

        <Route
          path="/user"
          element={<ProtectedRoute children={<UserPage />} admin={false} />}
        />

        <Route
          path="/user/add"
          element={<ProtectedRoute children={<AddUserPage />} admin={true} />}
        />

        { /* FOR TESTING */ }
        {/* <Route
          path="/user/add"
          element={<AddUserPage />}
        /> */}

      </Routes>
    </div>
  );
};

export default App;
