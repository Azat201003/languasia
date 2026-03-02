import React from 'react';
import Logo from '../Logo.svg';
import './Header.css';
import { useNavigate } from 'react-router-dom';

const Header = () => {
  const navigate = useNavigate();
  
  const onProfile = async (e) => {
    navigate("/my")
  }

  return (
    <header className="top-header">
      <a href="/">
          <img src={Logo} alt="Logo" className="logo"/>
      </a>
      <a href="/messanger/">Messenger</a>
      <a href="/search/">Find users</a>
      <button className="profile-btn" aria-label="Профиль" onClick={onProfile} />
    </header>
  );
}

export default Header;
