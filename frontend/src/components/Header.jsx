import React, { useState, useContext, useEffect } from 'react';
import Logo from '../assets/Logo.svg';
import "./Header.css";
import api from "../api.jsx";
import { redirect, useNavigate } from "react-router-dom";

const Header = () => {
  const [color, setColor] = useState('#ffffff');
  const [nickname, setNickname] = useState('');

  const navigate = useNavigate();
  
  const api_url = import.meta.env.VITE_API_URL;

  const user_id = localStorage.getItem('user_id');

  useEffect(() => {
    const fetchUserColor = async () => {
        const userInfo = await fetch(api_url + `/users`, {method: "POST", headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ user_id: parseInt(user_id) }),});
        const data = await userInfo.json();
        setColor(data[0].color);
        setNickname(data[0].nickname);
    };

    fetchUserColor();

  }, [api_url, user_id]);
  
  const onProfile = async (e) => {
    navigate("/my");
  }

  const profileStyle = {backgroundColor: color}

  return (
    <>
      <header className="top-header">
        <img src={Logo} alt="Logo" className="logo" />
	<a href="/">Messenger</a>
        <a href="/search">Search People</a>
        <div className="profile-container">
          <div className="nickname">{ nickname }</div>
          <button style={profileStyle} className="profile-btn" aria-label="Профиль" onClick={onProfile} />
        </div>
      </header>
    </>
  );
};

export {Header};

