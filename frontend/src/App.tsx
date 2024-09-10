import React from "react";
import { Routes, Route } from "react-router-dom";
import ProtectedRoute from "./ProtectedRoute";
import UserLandingPage from "./components/User/LandingPage/User/User";
import ViewUserPage from "./components/User/View/ViewUser";
import ModifyUserPage from "./components/User/Modify/Modify";
import ViewCustomerPage from "./components/Customer/LandingPage/Customer";
import ModifyCustomerPage from "./components/Customer/Modify/Modify";
import HomePage from "./components/Home/Home";
import LandingPage from "./components/LandingPage/LandingPage";
import LoginPage from "./components/Login/Login";
import ViewSupplierPage from "./components/Supplier/LandingPage/Supplier";
import ModifySupplierPage from "./components/Supplier/Modify/Modify";

export const BACKEND_BASE_URL = "localhost:90808/api/v1";

const App: React.FC = () => {
  return (
    <div className="App">
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />

        {/* FOR TESTING PURPOSE ONLY */}
        <Route path="/home" element={<HomePage />} />
        <Route path="/user" element={<UserLandingPage />} />
        <Route path="/user/view" element={<ViewUserPage />} />
        <Route path="/user/detail" element={<ModifyUserPage />} />

        <Route path="/customer" element={<ViewCustomerPage />} />
        <Route path="/customer/detail" element={<ModifyCustomerPage />} />

        <Route path="/supplier" element={<ViewSupplierPage />} />
        <Route path="supplier/detail" element={<ModifySupplierPage />} />

        {/* <Route
          path="/home"
          element={<ProtectedRoute children={<HomePage />} admin={false} />}
        /> */}

        {/* USER ROUTE */}
        {/* <Route
          path="/user"
          element={<ProtectedRoute children={<UserLandingPage />} admin={false} />}
        />
        <Route
          path="/user/view"
          element={<ProtectedRoute children={<ViewUserPage />} admin={true} />}
        />
        <Route
          path="/user/detail"
          element={<ProtectedRoute children={<ModifyUserPage />} admin={false} />}
        />
        <Route
          path="/user/create"
          element={<ProtectedRoute children={<ModifyUserPage />} admin={true} />}
        /> */}

        {/* CUSTOMER ROUTE */}
        {/* <Route
          path="/customer"
          element={
            <ProtectedRoute children={<ViewCustomerPage />} admin={true} />
          }
        />
        <Route
          path="/customer/detail"
          element={
            <ProtectedRoute children={<ModifyCustomerPage />} admin={true} />
          }
        /> */}


      </Routes>
    </div>
  );
};

export default App;
