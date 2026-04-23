import React, { useState, useEffect, useRef } from 'react';
import './Messenger.css';
import Logo from '../assets/Logo.svg';
import sendSymbol from '../assets/sendSymbol.svg';
import {api, baseURL, wsURL} from '../api';
import Header from './Header';
import UserIcon from './UserIcon'

import styled from 'styled-components';
import SimpleBar from 'simplebar-react';
import 'simplebar-react/dist/simplebar.min.css';

const api_url = baseURL;
const ws_url = wsURL;

const Scroll = styled(SimpleBar)`
  .simplebar-track.simplebar-vertical {
    background-color: #00000000;
    pointer-events: auto;
    z-index: 2;
  }
  .simplebar-scrollbar::before {
    background-color: #5e5e5e !important;
    transition: opacity 0.8s !important;
  }
`;


function ChatList({ userid, activeChat, setActiveChat, setChatMessages, searchQuery }) {

  const [chats, setChats] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {

    const loadChats = async () => {
        setLoading(true);
        setError(null);

        const response = await api.get(api_url + `/users/${userid}/chats`);

        const rawChats = await response.data;

        console.log(rawChats);

        setChats(rawChats);

        // setChats(response.data);

        setLoading(false);
    }
    loadChats()
  }, [userid]);

  if (loading) return <div className="chats-status">Loading chats...</div>;
  if (error)   return <div className="chats-status">Error: {error}</div>;

  const filteredChats = chats.filter(chat =>
    chat.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
   <>
    {filteredChats.length === 0 ? (
      <p className="chats-status">You have no chats</p>
    ) : (filteredChats.map((chat, index) => (
      <div key={index} className={`chat-item ${activeChat?.chat_id == chat.chat_id ? "active" : ""}`} onClick={() => (setActiveChat(chat))}>
        <UserIcon className="chat-avatar" color={chat.color} size="52" />
        <div className="chat-info">
          <div className="chat-name">{chat.title}</div>
          <div className="chat-preview"></div>
        </div>
        <div className="chat-time"></div>
      </div>)))}
      </>
  );
}


const Messenger = () => {

    const [activeChat, setActiveChat] = useState(null);

    const [message, setMessage] = useState('');

    const [chatMessages, setChatMessages] = useState({});

    const [searchQuery, setSearchQuery] = useState('');

    const user_id = parseInt(localStorage.getItem('user_id'));
    const token = localStorage.getItem('token');



    // const websocket = useRef(new WebSocket(`${ws_url}/ws` + `?access_token=${token}`));
    const websocket = useRef(null);

    useEffect(() => {
        const ws = new WebSocket(`${ws_url}/ws?access_token=${token}`);

        websocket.current = ws;

        ws.onopen = () => {
            console.log('WebSocket connected');
        };

        websocket.current.onmessage = (event) => {
            if (event.data.chat_id && event.data.chat_id !== activeChat?.chat_id) return;

            const newMessage = JSON.parse(event.data);

            if (newMessage.type != "pong") {
                setChatMessages(newChatMessages =>{

	var newChatMessages = chatMessages
	if (chatMessages[newMessage.chat_id] == undefined) {
		chatMessages[newMessage.chat_id] = []
	}
	const index = newChatMessages[newMessage.chat_id].findIndex(x => x.message_id > newMessage.message_id);
      	if (index === -1) {
	  newChatMessages[newMessage.chat_id].push(newMessage);
	} else {
	  newChatMessages[newMessage.chat_id].splice(index, 0, newMessage);
	}
          })

	//chatMessages[newMessage.chat_id] = [..chatMessages[newMessage.chat_id]]
        //setChatMessages(prev => [newMessage, ...prev].sort((a, b) => a.message_id - b.message_id));
      }
    }

    return () => {
      ws.close();
    };
  }, [token]);


  const onPing = async (e) => {
    const message = {
      type: "ping",
    };
    setTimeout(() => {websocket.current.send(JSON.stringify(message))}, 2000);
  }

  const onMessageSend = async(e, chatId) => {
    const websocketMessage = {
      type: "chat",
      chat_id: chatId,
      content: message
    };
    websocket.current.send(JSON.stringify(websocketMessage));
    setMessage('');
  }



  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      onMessageSend(e, activeChat.chat_id);
    }
  };

  const loadChat = async(chatId) => {
    const loadChatMessage = {
      type: "recieve_messages",
      chat_id: chatId,
      from_message_id: 0,
      limit: 50
    };
    websocket.current.send(JSON.stringify(loadChatMessage));
		console.log(chatMessages, activeChat, chatMessages[activeChat]);
  }

  // console.log(setTimeout((e) => loadChat(e, 6), 2000));

  useEffect(() => {
    if (activeChat) {
      if (chatMessages[activeChat.chat_id] == undefined)
      setChatMessages(prev => ({ ...prev, [activeChat.chat_id]: [] }));

      console.log("some", chatMessages);

      loadChat(activeChat.chat_id);

    }
	console.log(chatMessages, activeChat, chatMessages[activeChat]);
  }, [activeChat]);

  
  const bottomRef = useRef(null);

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatMessages]);

  const handleScroll = (e) => {
    const {scrollTop, scrollHeight, clientHeight} = e.target;
    console.log(e.target);
  }

  return (
  <>
    {/* Верхняя панель */}
    <Header/>
    {/* Основной контент */}

    
    
    <div className="main-content">
      {/* Левая панель: список чатов */}
      <aside className="chats-sidebar">
        <div className="search-background">
          <input
            type="text"
            className="search-input"
            placeholder="Поиск по чатам"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <Scroll 
          style={{ maxHeight: '100%', flex: 1, overflowY: 'auto' }}
        >
        <div className="chats-list">
          <ChatList userid={user_id} activeChat={activeChat} setActiveChat={setActiveChat} setChatMessages={setChatMessages} searchQuery={searchQuery}/>
        </div>
        </Scroll>
      </aside>

      {/* Правая панель: переписка */}
    {activeChat != null && (
      <main className="chat-area">
        
        <div className="chat-header">
          <UserIcon
            className="chat-avatar"
            color={activeChat.color}
          />
          <div className="chat-name">{activeChat.title}</div>
        </div>

        <Scroll 
          style={{ maxHeight: '100%', flex: 1, overflowY: 'auto' }}
        >

        <div className="messages" onScroll={handleScroll}>

          {(chatMessages[activeChat.chat_id]||[]).map((msg, index) => (<div className={`message ${msg.sender_id != user_id ? 'received' : 'sent'}`} key={msg.message_id}> {msg.content} <div className="message-time">{(parseInt(msg.created_at.split("T")[1].split("Z")[0].split(":")[0] - parseInt(new Date().getTimezoneOffset())/60)) + ":" + msg.created_at.split("T")[1].split("Z")[0].split(":")[1]}</div> </div>))}
          
          <div ref={bottomRef} />
        </div>

        </Scroll>

        <div className="message-input-wrapper">
          <input
            type="text"
            value={message}
            className="message-input"
            placeholder="Write a message..."
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <button className="send-btn" aria-label="Отправить"  onClick={(e) => onMessageSend(e, activeChat.chat_id)}>
            <img 
              src={sendSymbol} 
              className="send-symbol"
            />
          </button>

        </div>
      </main>
        )}
    </div>
  </>
);
};

export default Messenger;
