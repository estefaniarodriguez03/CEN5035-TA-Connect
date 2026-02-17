import { useState, useEffect } from "react";
import { login } from "../api/auth";
import { useAuth } from "../context/AuthContext";
import { useNavigate, useLocation } from "react-router-dom";
import AuthLayout from "../components/AuthLayout";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [welcomeMessage, setWelcomeMessage] = useState("");

  // Top-level hooks
  const { login: authLogin } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // Check if redirected from registration
  useEffect(() => {
    if (location.state?.fromRegister) {
      setWelcomeMessage("Registration successful! Please log in below.");
    }
  }, [location.state]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    try {
      const data = await login(email, password);

      // Use context instead of localStorage directly
      authLogin(data.token, data.user);

      // Role-based redirect
      if (data.user.role === "ta") {
        navigate("/ta");
      } else {
        navigate("/student");
      }
    } catch (err) {
      setWelcomeMessage("Login failed. Check your email and password.");
    }
  }

  return (
    <AuthLayout
      title="Login"
      footerText="Don't have an account?"
      footerLink="/register"
      footerLinkText="Register"
    >
      {/* Show welcome message or login error */}
      {welcomeMessage && <p className="success-message">{welcomeMessage}</p>}

      <form onSubmit={handleSubmit}>
        <input
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />

        <button type="submit">Login</button>
      </form>
    </AuthLayout>
  );
}
