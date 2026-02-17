import { useAuth } from "../context/AuthContext";

export default function StudentDashboard() {
  const { user, logout } = useAuth();

  return (
    <div>
      <h1>Student Dashboard</h1>
      <p>Welcome, {user?.username}</p>
      <p>Role: {user?.role}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
