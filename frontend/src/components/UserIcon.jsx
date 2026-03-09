import React from 'react';
import "./UserIcon.css";
import UserLogoFilter from "../assets/UserLogoFilter.svg";

function UserIcon({ color }) {
  console.log(UserLogoFilter)
  const coloredStyle = {
    backgroundColor: color,
    maskImage: `url("${UserLogoFilter}")`,
    WebkitMaskImage: `url("${UserLogoFilter}")`,
    maskSize: "contain",
    WebkitMaskSize: "contain",
    maskRepeat: "no-repeat",
    WebkitMaskRepeat: "no-repeat",
    width: "42px",
    height: "42px",
    border: "none"           // remove default button border
  };

  return (
    <div class="profile-btn-bg">
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
