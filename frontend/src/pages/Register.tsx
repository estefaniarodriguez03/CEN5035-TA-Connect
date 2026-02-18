import { useState } from "react";
import { register } from "../api/auth";
import AuthLayout from "../components/AuthLayout";
import { useNavigate } from "react-router-dom";

export default function Register() {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"student" | "ta">("student");
  const [message, setMessage] = useState("");
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    try {
      await register(username, email, password, role);

      setMessage("Registration successful! Redirecting to login...");

      setTimeout(() => {
        navigate("/login", { state: { fromRegister: true } });
      }, 1500);


    } catch (err: any) {
      // Show error message
      if (err?.response?.data?.error) {
        setMessage(`Registration failed: ${err.response.data.error}`);
      } else {
        setMessage("Registration failed. The username or email is already in use.");
      }
    }
  }

  return (
    <AuthLayout
      title="Register"
      footerText="Already have an account?"
      footerLink="/login"
      footerLinkText="Login"
    >
      {/* Inline message */}
      {message && <p className="success-message">{message}</p>}

      <form onSubmit={handleSubmit}>
        <input
          placeholder="Name"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />

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

        <select value={role} onChange={(e) => setRole(e.target.value as any)}>
          <option value="student">Student</option>
          <option value="ta">TA</option>
        </select>

        <button type="submit">Register</button>
      </form>
    </AuthLayout>
  );
}
