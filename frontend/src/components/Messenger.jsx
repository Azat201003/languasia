import React, { useState } from 'react';
import './Messenger.css';
import Logo from '../assets/Logo.svg';
import sendSymbol from '../assets/sendSymbol.svg';

const Messenger = () => {

  const user_id = localStorage.getItem('user_id');
  const token = localStorage.getItem('token');

  const api_url = import.meta.env.VITE_API_URL;
  
  const onWebsocketConnect = async (e) => {
    const response = await fetch(api_url + `/ws?access_token=${token}`, {
      method: 'GET',
    });
  }


  return (
  <>
    {/* Верхняя панель */}
    <header className="top-header">
      <img src={Logo} alt="Logo" className="logo" />
      <button className="profile-btn" aria-label="Профиль" />
    </header>

    {/* Основной контент */}
    <div className="main-content">
      {/* Левая панель: список чатов */}
      <aside className="chats-sidebar">
        <div className="search-background">
          <input
            type="text"
            className="search-input"
            placeholder="Поиск по чатам"
          />
        </div>
        <div className="chats-list">
          <div className="chat-item active">
            <div
              className="chat-avatar"
            />
            <div className="chat-info">
              <div className="chat-name">Алексей Петров</div>
              <div className="chat-preview">Ок, давай в 19:00</div>
            </div>
            <div className="chat-time">14:32</div>
          </div>

          <div className="chat-item">
            <div
              className="chat-avatar"
            />
            <div className="chat-info">
              <div className="chat-name">Команда проекта</div>
              <div className="chat-preview">Мария: не забудьте про встречу</div>
            </div>
            <div className="chat-time">12:05</div>
          </div>

          <div className="chat-item">
            <div
              className="chat-avatar"
            />
            <div className="chat-info">
              <div className="chat-name">Друзья</div>
              <div className="chat-preview">Ты: фото с вчерашнего дня</div>
            </div>
            <div className="chat-time">Вчера</div>
          </div>

          {/* Дополнительные чаты добавляйте здесь */}
        </div>
      </aside>

      {/* Правая панель: переписка */}
      <main className="chat-area">
        <div className="chat-header">
          <div
            className="chat-avatar"
          />
          <div className="chat-name">Алексей Петров</div>
        </div>

        <div className="messages">
          <div className="message received">
            Привет! Как дела?
            <div className="message-time">14:20</div>
          </div>

          <div className="message sent">
            Нормально, спасибо! А у тебя?
            <div className="message-time">14:21</div>
          </div>

          <div className="message received">
            Тоже хорошо. Встретимся сегодня вечером?
            <div className="message-time">14:25</div>
          </div>

          <div className="message sent">
            Да, давай в 19:00
            <div className="message-time">14:32</div>
          </div>

          {/* Дополнительные сообщения добавляйте здесь */}
        </div>

        <div className="message-input-wrapper">
          <input
            type="text"
            className="message-input"
            placeholder="Напишите сообщение..."
          />
          <button className="send-btn" aria-label="Отправить"  onClick={onWebsocketConnect}>
            <img 
              src={sendSymbol} 
              className="send-symbol"
            />
          </button>
        </div>
      </main>
    </div>
  </>
);
};

export {Messenger};
