import React from 'react';
import "./UserIcon.css";
import UserLogoFilter from "../assets/UserLogoFilter.svg";

function UserIcon({ color, size=42 }) {
  console.log(UserLogoFilter)
  const coloredStyle = {
    backgroundColor: color,
    maskImage: `url("${UserLogoFilter}")`,
    WebkitMaskImage: `url("${UserLogoFilter}")`,
    maskSize: "contain",
    WebkitMaskSize: "contain",
    maskRepeat: "no-repeat",
    WebkitMaskRepeat: "no-repeat",
    filter: "brightness(70%)",
    width: `${size}px`,
    height: `${size}px`,
    border: "none"           // remove default button border
  };
  const backgroundStyles = {
    width: `${size}px`,
    height: `${size}px`,
    backgroundColor: color,
  };

  return (
    <div className="profile-btn-bg" style={backgroundStyles}>
        <button
          style={coloredStyle}
          className="profile-btn"
          type="button"           // prevent accidental form submits
          aria-label="Профиль"
        />
    </div>
  );
}

export default UserIcon;
