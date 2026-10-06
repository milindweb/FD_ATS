import { NavLink } from "react-router-dom";
import { navItems } from "../../lib/nav";
import { cycleTheme, currentTheme } from "../../lib/theme";
import { Icon } from "../../lib/icons";
import { IconButton } from "../ui/Button";
import { useState } from "react";
import aartiLogo from "../../assets/images/logo-ats.svg";

export interface AppHeaderProps {
  onMenuOpen: () => void;
  onLogout?: () => void;
}

/** Sticky header: brand, main navigation, theme toggle, mobile menu. */
export function AppHeader({ onMenuOpen, onLogout }: AppHeaderProps) {
  const [theme, setTheme] = useState(currentTheme());

  const toggleTheme = () => setTheme(cycleTheme());

  return (
    <header className="hs-header">
      <IconButton
        icon="menu"
        label="Open menu"
        className="hs-menu-btn"
        onClick={onMenuOpen}
      />

      <NavLink to="/" className="hs-header__brand">
        <img className="hs-header__logo" src={aartiLogo} alt="" aria-hidden="true" />
        <span>Fixed Deposit Management</span>
      </NavLink>

      <nav className="hs-header__nav" aria-label="Main navigation">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === "/"}
            className={({ isActive }) => `hs-navlink ${isActive ? "is-active" : ""}`}
          >
            <Icon name={item.icon} size={15} />
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="hs-header__actions">
        {onLogout && <IconButton icon="logout" label="Log out" onClick={onLogout} />}
        <IconButton
          icon={theme === "dark" ? "moon" : "sun"}
          label={`Theme: ${theme}. Click to change.`}
          onClick={toggleTheme}
        />
      </div>
    </header>
  );
}
