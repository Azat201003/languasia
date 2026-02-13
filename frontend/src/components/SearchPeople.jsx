import React, { useState, useEffect } from 'react';
import './SearchPeople.css';
import api from "../api.jsx";
import Header from './Header'

const SearchPeople = () => {
    /*
  const [users, setUsers] = useState([]);
  const [languages, setLanguages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchString, setSearchString] = useStat("")

  const updateQuery = () => {
    const fetchData = async () => {
      try {
        const response = await fetch(
          'https://95.165.132.221/api/users'
        );

        if (!response.ok) {
          throw new Error('Network response was not ok');
        }

        const data = await response.json();
        setUsers(data);
      } catch (error) {
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  });
  useEffect(updateQuery, []);
  useEffect(() => {
      try {
        const response = await fetch(
          'https://95.165.132.221/api/languages'
        );

        if (!response.ok) {
          throw new Error('Network response was not ok');
        }

        const data = await response.json();
        setUsers(data);
      } catch (error) {
        setError(error.message);
      } finally {
        setLoading(false);
      }
    
  }, []);

  if (loading) return <div>Loading todos...</div>;
  if (error) return <div>Error: {error}</div>;

    */

  return (
    <>
      <Header />
      <div className="main-content">
        <div className="search-sidebar">
          <div className="search-form">
            <div className="search-field">
              <label className="search-label">Поисковая строка</label>
              <input type="text" className="search-input" placeholder="Имя, ник, ключевые слова"/>
            </div>
            <div className="search-field">
              <label className="search-label">Хобби</label>
              <input type="text" className="search-input" placeholder="Через запятую: чтение, йога, путешествия"/>
            </div>
            <div className="search-field">
              <label className="search-label">Известные языки</label>
              <input type="text" className="search-input" list="language-list" placeholder="Через запятую: русский, английский" />
            </div>
            <div className="search-field">
              <label className="search-label">Изучаемые языки</label>
              <input type="text" className="search-input" list="language-list" placeholder="Через запятую: испанский, немецкий" />
            </div>
            <button className="search-button">Искать</button>
          </div>

          <datalist id="language-list">
            <option value="Английский" />
          </datalist>
        </div>

        <div className="results-area">
          <div className="results-header">Результаты поиска</div>
          <div className="results-list">
            <div className="chat-item active">
              <div className="chat-avatar"></div>
              <div className="chat-info">
                <div className="chat-name">Анна</div>
                <div className="chat-preview">Хобби: чтение, йога, путешествия • Знает: русский, английский • Учит: испанский, французский</div>
              </div>
              <div className="chat-status">Онлайн</div>
            </div>

            <div className="chat-item">
              <div className="chat-avatar"></div>
              <div className="chat-info">
                <div className="chat-name">Максим</div>
                <div className="chat-preview">Хобби: футбол, программирование • Знает: русский, английский, немецкий • Учит: японский</div>
              </div>
              <div className="chat-status">2 ч. назад</div>
            </div>

            <div className="chat-item">
              <div className="chat-avatar"></div>
              <div className="chat-info">
                <div className="chat-name">София</div>
                <div className="chat-preview">Хобби: рисование, кулинария • Знает: английский, французский • Учит: русский</div>
              </div>
              <div className="chat-status">Вчера</div>
            </div>

            <div className="chat-item">
              <div className="chat-avatar"></div>
              <div className="chat-info">
                <div className="chat-name">Дмитрий</div>
                <div className="chat-preview">Хобби: музыка, бег • Знает: русский, английский • Учит: итальянский</div>
              </div>
              <div className="chat-status">3 дня назад</div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export { SearchPeople };
