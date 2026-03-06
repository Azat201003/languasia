import React from 'react';
import { Routes, BrowserRouter, Route, useNavigate, Navigate, Outlet } from 'react-router-dom';
import {AuthPage} from './components/AuthPage';
import {EditProfile} from './components/EditProfile';
import {SearchPeople} from './components/SearchPeople';
import Messenger from './components/Messenger';


function ProtectedRoute() {
  const token = localStorage.getItem('token');

  if (!token) {
    return <Navigate to="/auth" replace />;
  }

  return <Outlet />;
}


function App() {
  return (
    <BrowserRouter>
        <Routes>
          <Route path="/auth" element={< AuthPage/>} />
          <Route element={<ProtectedRoute />}>
          <Route path="/" element={< Messenger />} />
          <Route path="/messenger" element={< Messenger />} />
          <Route path="/my" element={<EditProfile
             initialDescription={""}
             initialHobbyTitles={["hdbvfl"]}
             initialKnownLanguageNames={["English"]}
             initialLearnLanguageNames={["English"]}
             onSave={(data) => {
               updateUser(data);
               navigate('/profile');
             }}
           />} />
          <Route path="/search" element={<SearchPeople/>} />
          </Route>
        </Routes>
    </BrowserRouter>
  );
}

export default App;
