import React, { useState, useEffect } from 'react';
import { api } from "../api.jsx";
import './Messenger.css';
import Logo from '../assets/Logo.svg';
import sendSymbol from '../assets/sendSymbol.svg';
import Header from './Header'

const Messenger = () => {

  const user_id = localStorage.getItem('user_id');
  const token = localStorage.getItem('token');
 
  const [chats, setChats] = useState([]);
  
    useEffect(() => {
    const fetchData = async () => {
      try {
        const chatRes = api.get(`/user/${user_id}/chats`)

        if (chatRes.status === 'fulfilled') {
          setChats(chatRes.value.data);
        } else {
          console.warn('Chatss endpoint not available – chats will be limited');
        }
      } catch (err) {
        console.error('Unexpected error', err);
      }
    };
    fetchData();
    console.log(chats);
  }, []);

  const api_url = import.meta.env.VITE_API_URL;
  
  const onWebsocketConnect = async (e) => {
    const response = await fetch(api_url + `/ws?access_token=${token}`, {
      method: 'GET',
    });
  }

  return (
  <>
    {/* Верхняя панель */}
    <Header />
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
              <div className="chat-name">Alex K.</div>
              <div className="chat-preview">Ok, let's on 19:00</div>
            </div>
            <div className="chat-time">14:32</div>
          </div>

          <div className="chat-item">
            <div
              className="chat-avatar"
            />
            <div className="chat-info">
              <div className="chat-name">Project team</div>
              <div className="chat-preview">Mary: don't forget about meating</div>
            </div>
            <div className="chat-time">12:05</div>
          </div>

          <div className="chat-item">
            <div
              className="chat-avatar"
            />
            <div className="chat-info">
              <div className="chat-name">Friends</div>
              <div className="chat-preview">You: Gestern Foto</div>
            </div>
            <div className="chat-time">yesterday</div>
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
          <div className="chat-name">Alex K.</div>
        </div>

        <div className="messages">
          <div className="message received">
            Hi! How are you?
            <div className="message-time">14:20</div>
          </div>

          <div className="message sent">
            Ok, thanks! And you?
            <div className="message-time">14:21</div>
          </div>

          <div className="message received">
            Ok too. Will we meet tomorrow?
            <div className="message-time">14:25</div>
          </div>

          <div className="message sent">
            Yeah, let'\''s at 19:00
            <div className="message-time">14:32</div>
          </div>

          {/* Дополнительные сообщения добавляйте здесь */}
        </div>

        <div className="message-input-wrapper">
          <input
            type="text"
            className="message-input"
            placeholder="Write message..."
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

export default Messenger;
