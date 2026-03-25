import { useAuth } from "../context/AuthContext";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";

interface TALayoutProps {
  children: React.ReactNode;
  sidebar?: React.ReactNode;
  activeTab?: 'dashboard' | 'office-hours' | 'queue';
  onTabChange?: (tab: 'dashboard' | 'office-hours' | 'queue') => void;
}

export default function TALayout({ children, sidebar, activeTab = 'dashboard', onTabChange }: TALayoutProps) {
  const { logout } = useAuth();

  return (
    <div className="ta-dashboard" style={{ minHeight: '100vh', width: '100%', display: 'block' }}>
      <nav className="navbar">
        <div className="navbar-left">
          <div className="logo">
            <img src={ufLogo} alt="UF Logo" className="logo-icon" />
            <span className="logo-text">TA Connect</span>
          </div>
          <div className="nav-tabs">
            <button
              className={`nav-tab ${activeTab === 'dashboard' ? 'active' : ''}`}
              onClick={() => onTabChange?.('dashboard')}
            >
              Dashboard
            </button>
            <button
              className={`nav-tab ${activeTab === 'office-hours' ? 'active' : ''}`}
              onClick={() => onTabChange?.('office-hours')}
            >
              My Office Hours
            </button>
            <button
              className={`nav-tab ${activeTab === 'queue' ? 'active' : ''}`}
              onClick={() => onTabChange?.('queue')}
            >
              Queue
            </button>
          </div>
        </div>
        <div className="navbar-right">
          <div className="notification-icon">
            <img src={whiteNotificationIcon} alt="Notifications" />
          </div>
          <button className="profile-icon" onClick={logout}>
            <img src={whiteProfileIcon} alt="Profile" />
          </button>
        </div>
      </nav>
      <div className="dashboard-content">
        {children}
        {sidebar && <aside className="sidebar">{sidebar}</aside>}
      </div>
    </div>
  );
}