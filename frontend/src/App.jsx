import React from 'react';
import { Routes, BrowserRouter, Route, Navigate, Outlet } from 'react-router-dom';
import {AuthPage} from './components/AuthPage';
import {EditProfile} from './components/EditProfile';
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
            <Route path="/my" element={<EditProfile
              initialDescription={""}
              initialHobbyTitles={["hdbvfl"]}
              initialKnownLanguageNames={["English"]}
              initialLearnLanguageNames={["English"]}
              onSave={(data) => {
                updateUser(data);
                navigate('/profile');
              }}
              onCancel={() => navigate(-1)}
            />} />
        </Route>
        </Routes>
    </BrowserRouter>
  );
}

export default App;
