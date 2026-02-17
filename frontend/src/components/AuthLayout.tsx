import { Link } from "react-router-dom";

export default function AuthLayout({
  title,
  children,
  footerText,
  footerLink,
  footerLinkText,
}: {
  title: string;
  children: React.ReactNode;
  footerText: string;
  footerLink: string;
  footerLinkText: string;
}) {
  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-left">
          <h1>Welcome to TA Connect</h1>
          <p>
            Login or create an account to start your office hours experience.
          </p>
        </div>

        <div className="auth-right">
          <h2>{title}</h2>
          {children}

          <div className="auth-footer">
            {footerText} <Link to={footerLink}>{footerLinkText}</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
