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
import ViewPatientPage from "./components/Patient/LandingPage/Patient";
import ModifyPatientPage from "./components/Patient/Modify/Modify";
import ViewDoctorPage from "./components/Doctor/LandingPage/Doctor";
import ModifyDoctorPage from "./components/Doctor/Modify/Modify";

export const BACKEND_BASE_URL = "localhost:9988/api/v1";

const App: React.FC = () => {
  return (
    <div className="App">
      <Routes>
        { /* NOT PROTECTED */ }
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />

        { /* ----------------FOR TESTING PURPOSE ONLY------------------------ */ }
        <Route path="/home" element={<HomePage />} />
        <Route path="/user" element={<UserLandingPage />} />
        <Route path="/user/view" element={<ViewUserPage />} />
        <Route path="/user/detail" element={<ModifyUserPage />} />

        <Route path="/customer" element={<ViewCustomerPage />} />
        <Route path="/customer/detail" element={<ModifyCustomerPage />} />

        <Route path="/supplier" element={<ViewSupplierPage />} />
        <Route path="/supplier/detail" element={<ModifySupplierPage />} />

        <Route path="/patient" element={<ViewPatientPage />} />
        <Route path="/patient/detail" element={<ModifyPatientPage />} />

        <Route path="/doctor" element={<ViewDoctorPage />} />
        <Route path="/doctor/detail" element={<ModifyDoctorPage />} />
        { /* ----------------------------------------------------------------- */ }

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
            <ProtectedRoute children={<ViewCustomerPage />} admin={false} />
          }
        />
        <Route
          path="/customer/detail"
          element={
            <ProtectedRoute children={<ModifyCustomerPage />} admin={false} />
          }
        /> */}

        {/* SUPPLIER ROUTE */}
        {/* <Route
          path="/supplier"
          element={
            <ProtectedRoute children={<ViewSupplierPage />} admin={false} />
          }
        />
        <Route
          path="/supplier/detail"
          element={
            <ProtectedRoute children={<ModifySupplierPage />} admin={false} />
          }
        /> */}

        {/* PATIENT ROUTE */}
        {/* <Route
          path="/patient"
          element={
            <ProtectedRoute children={<ViewPatientPage />} admin={false} />
          }
        />
        <Route
          path="/patient/detail"
          element={
            <ProtectedRoute children={<ModifyPatientPage />} admin={false} />
          }
        /> */}

        {/* Doctor ROUTE */}
        {/* <Route
          path="/doctor"
          element={
            <ProtectedRoute children={<ViewDoctorPage />} admin={false} />
          }
        />
        <Route
          path="/doctor/detail"
          element={
            <ProtectedRoute children={<ModifyDoctorPage />} admin={false} />
          }
        /> */}

      </Routes>
    </div>
  );
};

export default App;
