import React, { useState, useContext } from 'react';
import './AuthPage.css';
import Logo from '../assets/Logo.svg';
import { authContext } from "../contexts/AuthContext.jsx";
import { api, baseURL } from "../api.jsx";
import { redirect, useNavigate } from "react-router-dom";

const AuthPage = () => {

  const navigate = useNavigate();
  
  const api_url = baseURL;
  
  const [key, setKey] = useState(0);

  const triggerAnimation = () => {
    setKey(prev => prev + 1);
  };
  
  const [isLogin, setIsLogin] = useState(true);
  const [isSimilar, setIsSimilar] = useState(false);
  const [passwordError, setPasswordError] = useState('');

  const [registerData, setRegisterData] = useState({
    login: '',
    password: '',
    confirmPassword: '',
  });


  const handleChange = (e) => {
    const { name, value } = e.target;
    setRegisterData(prev => ({
      ...prev,
      [name]: value.replace(/\s/g, ''),
    }));
    setPasswordError('');
  };


  const onRegisterSubmit = async (e) => {
    e.preventDefault();

    if (registerData.password !== registerData.confirmPassword) {
      setPasswordError('Passwords do not match');
      triggerAnimation();
      return;
    }
    
    if (registerData.password.length < 2) {
      setPasswordError('Password is too short');
      triggerAnimation();
      return;
    }
    
    const response = await fetch(api_url + '/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: registerData.login,
        password: registerData.password,
      }),
    });

    if (response.status == 422) {
      setPasswordError('User already exists');
      triggerAnimation();
      return;
    }

    const loginResponse = await fetch(api_url + '/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: registerData.login,
        password: registerData.password,
      }),
    });
    
    if (loginResponse.status == 422){
      setPasswordError('User does not exist');
      triggerAnimation();
      return;
    }

    if (loginResponse.status == 401) {
      setPasswordError('Wrong password');
      triggerAnimation();
      return;
    }
    
    const data = await loginResponse.json();
    
    localStorage.setItem('token', data.jwt_token);
    localStorage.setItem('user_id', data.user_id);
    localStorage.setItem('refresh_token', data.refresh_token);

    navigate("/");
    
  };
  
  const onLoginSubmit = async (e) => {
    e.preventDefault();
      
    const loginResponse = await fetch(api_url + '/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: registerData.login,
        password: registerData.password,
      }),
    });
    
    if (loginResponse.status == 422){
      setPasswordError('User does not exist');
      triggerAnimation();
      return;
    }

    if (loginResponse.status == 401) {
      setPasswordError('Wrong password');
      triggerAnimation();
      return;
    }
    
    const data = await loginResponse.json();
    
    localStorage.setItem('token', data.jwt_token);
    localStorage.setItem('user_id', data.user_id);
    localStorage.setItem('refresh_token', data.refresh_token);

    navigate("/");

  };

  return (
    <div className="auth-background">
      <div className="auth-card">
        <div className="auth-header">
          <img src={Logo} alt="LanguasiA" className="logo" />
        </div>

        <div className="auth-tabs">
          <button
            className={isLogin ? 'active' : ''}
            onClick={() => setIsLogin(true)}
          >
            Login
          </button>
          <button
            className={!isLogin ? 'active' : ''}
            onClick={() => setIsLogin(false)}
          >
            Register
          </button>

	  <div 
    	    className="glider"
    	      style={{
      	        transform: isLogin ? 'translateX(0%)' : 'translateX(105.5%)'
    	    }}
  	  />
        </div>

        <div className="auth-content">
          {isLogin ? (
            <form className="auth-form" onSubmit={onLoginSubmit}>
              <input name="login" value={registerData.login} type="text" placeholder="Username" onChange={handleChange} required />
              <input name="password" value={registerData.password} type="password" placeholder="Password" onChange={handleChange} />
              {passwordError && <p key={key} className="passwordError error-animate">{passwordError}</p>}
              <button type="submit" className="submit-btn">
                Log in
              </button>
            </form>
          ) : (
            <form className="auth-form" onSubmit={onRegisterSubmit}>
              <input name="login" value={registerData.login} type="text" placeholder="Username" onChange={handleChange} required />
              <input name="password" value={registerData.password} type="password" placeholder="Password" onChange={handleChange} required />
              <input name="confirmPassword" value={registerData.confirmPassword} type="password" placeholder="Confirm password" onChange={handleChange} required />
              {passwordError && <p key={key} className="passwordError error-animate">{passwordError}</p>}
              <button type="submit" className="submit-btn">
                Create account
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};

export { AuthPage };
