import React, { useState, useContext, useEffect } from 'react';
import Logo from '../assets/Logo.svg';
import "./Header.css";
import {api} from "../api.jsx";
import UserIcon from "./UserIcon.jsx"
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

  return (
    <>
      <header className="top-header">
	  <a href="/"> <img src={Logo} alt="Logo" className="logo" /> </a>
      <a href="/">Messenger</a>
        <a href="/search">Search People</a>
        <div className="profile-container">
          <a href="/my" className="nickname">{ nickname }</a>
          <a href="/my"> <UserIcon color={color} /> </a>
        </div>
      </header>
    </>
  );
};

export default Header;

