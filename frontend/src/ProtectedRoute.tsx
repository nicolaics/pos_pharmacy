// ProtectedRoute.tsx
import React from "react";
import { useNavigate } from "react-router-dom";
import { AuthMiddleware } from "./Middleware";

interface ProtectedRouteProps {
  children: React.ReactNode;
  admin: boolean;
}

// Higher-order component to apply the authMiddleware before rendering the page
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, admin }) => {
  const navigate = useNavigate();

  React.useEffect(() => {
    // Run middleware when the component is mounted
    AuthMiddleware(navigate, admin);
  }, [navigate]);

  return <>{children}</>; // Render the children if the middleware passes
};

export default ProtectedRoute;
