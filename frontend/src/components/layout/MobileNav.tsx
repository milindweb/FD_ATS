import { NavLink } from "react-router-dom";
import { useEscapeKey } from "../../lib/hooks";
import { Icon } from "../../lib/icons";
import { navItems } from "../../lib/nav";
import { brand } from "../../lib/brandInfo";
import { IconButton } from "../ui/Button";
import aartiLogo from "../../assets/images/logo-ats.svg";

export interface MobileNavProps {
  open: boolean;
  onClose: () => void;
}

/** Drawer navigation for small screens using the same nav data as the header. */
export function MobileNav({ open, onClose }: MobileNavProps) {
  useEscapeKey(open, onClose);

  if (!open) return null;

  return (
    <div className="hs-drawer is-open">
      <div className="hs-drawer__overlay" onClick={onClose} />
      <div className="hs-drawer__panel" role="dialog" aria-modal="true" aria-label="Navigation menu">
        <div className="hs-drawer__head">
          <span className="hs-header__brand">
            <img className="hs-header__logo" src={aartiLogo} alt="" aria-hidden="true" />
            <span>{brand.product}</span>
          </span>
          <IconButton icon="close" label="Close menu" onClick={onClose} />
        </div>

        <nav className="hs-drawer__nav" aria-label="Main navigation">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) => `hs-navlink ${isActive ? "is-active" : ""}`}
              onClick={onClose}
            >
              <Icon name={item.icon} size={16} />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="hs-drawer__foot">
          Designed by {brand.developer}
          <br />
          {brand.copyright}
        </div>
      </div>
    </div>
  );
}
