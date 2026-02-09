import React from 'react';
import Logo from '../Logo.svg';

function Header() {
  return (
    <header>
      <img className="logo" src={Logo} alt="LanguasiA" />
      <nav>
        <a href="#" className="active">Messenger</a>
        <a href="#">Find people</a>
      </nav>
      <div className="profile-icon">👤</div>
    </header>
  );
}

export default Header;