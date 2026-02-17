import { useAuth } from "../context/AuthContext";

export default function TADashboard() {
  const { user, logout } = useAuth();

  return (
    <div>
      <h1>TA Dashboard</h1>
      <p>Welcome, {user?.username}</p>
      <p>Role: {user?.role}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
