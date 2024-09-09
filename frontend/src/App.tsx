import React from "react";
import LandingPage from "./components/Home/Home";
import LoginPage from "./components/Login/Login";
import { Routes, Route } from "react-router-dom";
import ProtectedRoute from "./ProtectedRoute";
import UserLandingPage from "./components/User/LandingPage/User/User";
import ViewUserPage from "./components/User/View/ViewUser";
import ModifyUserPage from "./components/User/Modify/Modify";
import ViewCustomerPage from "./components/Customer/LandingPage/Customer";
import ModifyCustomerPage from "./components/Customer/Modify/Modify";

const App: React.FC = () => {
  return (
    <div className="App">
      <Routes>
        <Route path="/" element={<LoginPage />} />

        {/* FOR TESTING PURPOSE ONLY */}
        <Route path="/home" element={<LandingPage />} />
        <Route path="/user" element={<UserLandingPage />} />
        <Route path="/user/view" element={<ViewUserPage />} />
        <Route path="/user/detail" element={<ModifyUserPage />} />

        <Route path="/customer" element={<ViewCustomerPage />} />
        <Route path="/customer/detail" element={<ModifyCustomerPage />} />


        {/* <Route
          path="/home"
          element={<ProtectedRoute children={<LandingPage />} admin={false} />}
        /> */}

        {/* <Route
          path="/user"
          element={<ProtectedRoute children={<UserLandingPage />} admin={false} />}
        /> */}

        {/* <Route
          path="/user/view"
          element={<ProtectedRoute children={<ViewUserPage />} admin={true} />}
        /> */}

        {/* <Route
          path="/user/detail"
          element={<ProtectedRoute children={<ModifyUserPage />} admin={false} />}
        /> */}

        {/* <Route
          path="/user/create"
          element={<ProtectedRoute children={<ModifyUserPage />} admin={true} />}
        /> */}
      </Routes>
    </div>
  );
};

export default App;
