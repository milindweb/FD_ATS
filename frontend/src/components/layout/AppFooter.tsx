import { Link } from "react-router-dom";
import { brand } from "../../lib/brandInfo";

/** Footer shown on every screen (SRS §33). */
export function AppFooter() {
  return (
    <footer className="hs-footer">
      <span>
        Designed by{" "}
        <a href={brand.website} target="_blank" rel="noreferrer">
          {brand.developer}
        </a>
      </span>
      <span>{brand.copyright}</span>
      <span className="hs-footer__links">
        <Link to="/privacy">Privacy Policy</Link>
        <Link to="/about">About</Link>
      </span>
    </footer>
  );
}
