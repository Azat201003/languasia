import React, { useState, useEffect } from 'react';
import './SearchPeople.css';
import { api } from "../api.jsx";
import Header from './Header';
import Scroll from "./Scroll"

const SearchPeople = () => {
  const [users, setUsers] = useState([]);
  const [languages, setLanguages] = useState([]);
  const [hobbies, setHobbies] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Selected IDs for each filter category
  const [selectedHobbyIds, setSelectedHobbyIds] = useState([]);
  const [selectedKnownLangIds, setSelectedKnownLangIds] = useState([]);
  const [selectedLearnLangIds, setSelectedLearnLangIds] = useState([]);

  // Input values for the chip fields
  const [hobbyInput, setHobbyInput] = useState('');
  const [knownLangInput, setKnownLangInput] = useState('');
  const [learnLangInput, setLearnLangInput] = useState('');

  // Free‑text search string
  const [searchString, setSearchString] = useState('');

  // Fetch languages and hobbies on mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [langRes, hobbyRes] = await Promise.allSettled([
          api.get('/languages'),
          api.get('/hobbies')
        ]);

        if (langRes.status === 'fulfilled') {
          setLanguages(langRes.value.data);
        } else {
          console.error('Failed to fetch languages', langRes.reason);
          setError('Не удалось загрузить список языков');
        }

        if (hobbyRes.status === 'fulfilled') {
          setHobbies(hobbyRes.value.data);
        } else {
          console.warn('Hobbies endpoint not available – hobby filtering will be limited');
        }
      } catch (err) {
        console.error('Unexpected error', err);
      }
    };
    fetchData();
  }, []);

  const createChat = async (goalId) => {
    api.post("/chats", {
      title: "doesn't matter",
      goal_id: goalId,
      type: "Direct"
    });
  };

  // Helper: get language name by ID
  const getLanguageName = (id) => {
    const lang = languages.find(l => l.language_id === id);
    return lang ? lang.name : id;
  };

  // Helper: get hobby title by ID
  const getHobbyTitle = (id) => {
    const hobby = hobbies.find(h => h.hobby_id === id);
    return hobby ? hobby.title : id;
  };

  // Handlers for adding items from the input field
  const handleAddHobby = () => {
    const trimmed = hobbyInput.trim();
    if (!trimmed) return;
    const hobby = hobbies.find(h => h.title.toLowerCase() === trimmed.toLowerCase());
    if (hobby && !selectedHobbyIds.includes(hobby.hobby_id)) {
      setSelectedHobbyIds(prev => [...prev, hobby.hobby_id]);
    }
    setHobbyInput('');
  };

  const handleAddKnownLang = () => {
    const trimmed = knownLangInput.trim();
    if (!trimmed) return;
    const lang = languages.find(l => l.name.toLowerCase() === trimmed.toLowerCase());
    if (lang && !selectedKnownLangIds.includes(lang.language_id)) {
      setSelectedKnownLangIds(prev => [...prev, lang.language_id]);
    }
    setKnownLangInput('');
  };

  const handleAddLearnLang = () => {
    const trimmed = learnLangInput.trim();
    if (!trimmed) return;
    const lang = languages.find(l => l.name.toLowerCase() === trimmed.toLowerCase());
    if (lang && !selectedLearnLangIds.includes(lang.language_id)) {
      setSelectedLearnLangIds(prev => [...prev, lang.language_id]);
    }
    setLearnLangInput('');
  };

  const handleSearch = async () => {
    setLoading(true);
    setError(null);

    const filter = {};

    if (searchString) filter.search_string = searchString;

    if (selectedHobbyIds.length > 0) {
      filter.hobbies = selectedHobbyIds;
    }

    if (selectedKnownLangIds.length > 0) {
      filter.known_languages = selectedKnownLangIds;
    }

    if (selectedLearnLangIds.length > 0) {
      filter.learn_languages = selectedLearnLangIds;
    }

    filter.page_size = 100;
    filter.page_number = 1;

    try {
      const response = await api.post('/users', filter);
      setUsers(response.data);
    } catch (err) {
      console.error('Search failed', err);
      setError('Ошибка при выполнении поиска');
    } finally {
      setLoading(false);
    }
  };

  // Format user preview line
  const formatPreview = (user) => {
    const parts = [];

    if (user.hobby_title_ids && user.hobby_title_ids.length > 0) {
      const hobbyNames = user.hobby_title_ids.map(id => getHobbyTitle(id)).join(', ');
      parts.push(`Hobbies: ${hobbyNames}`);
    }

    if (user.known_language_ids && user.known_language_ids.length > 0) {
      const langNames = user.known_language_ids.map(id => getLanguageName(id)).join(', ');
      parts.push(`Known languages: ${langNames}`);
    }

    if (user.learn_language_ids && user.learn_language_ids.length > 0) {
      const langNames = user.learn_language_ids.map(id => getLanguageName(id)).join(', ');
      parts.push(`Learning languages: ${langNames}`);
    }

    return parts.join(' • ');
  };

  // Render chips with remove button (edit‑profile style)
  const renderChips = (ids, type) => {
    const items = type === 'hobby'
      ? ids.map(id => ({ id, name: getHobbyTitle(id) }))
      : ids.map(id => ({ id, name: getLanguageName(id) }));

    return items.map(item => (
      <span key={item.id} className={`chip ${type === 'learn' ? 'learn' : ''}`}>
        {item.name}
        <span
          className="chip-close"
          onClick={() => {
            if (type === 'hobby') {
              setSelectedHobbyIds(prev => prev.filter(id => id !== item.id));
            } else if (type === 'known') {
              setSelectedKnownLangIds(prev => prev.filter(id => id !== item.id));
            } else {
              setSelectedLearnLangIds(prev => prev.filter(id => id !== item.id));
            }
          }}
        >
          ×
        </span>
      </span>
    ));
  };

  return (
    <>
      <Header />
      <div className="main-content">
        <div className="search-sidebar">
          <div className="search-form">
            {/* Free‑text search */}
            <div className="search-field">
              <label className="search-label">Search string</label>
              <input
                type="text"
                className="search-input"
                placeholder="Name, keywords, etc"
                value={searchString}
                onChange={(e) => setSearchString(e.target.value)}
              />
            </div>

            {/* Hobbies field with chip input */}
            <div className="search-field">
              <label className="search-label">Hobbies</label>
              <div className="chip-input-group">
                <input
                  type="text"
                  list="hobbies-options"
                  value={hobbyInput}
                  onChange={(e) => setHobbyInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddHobby();
                    }
                  }}
                  placeholder="Add hobby..."
                  className="chip-input"
                />
                <button
                  type="button"
                  onClick={handleAddHobby}
                  className="chip-add-btn"
                  disabled={!hobbies.length}
                  title={!hobbies.length ? 'List of hobbies is unavailable' : 'Add hobby'}
                >
                  +
                </button>
              </div>
              <div className="chips">
                {renderChips(selectedHobbyIds, 'hobby')}
              </div>
            </div>

            {/* Known languages field */}
            <div className="search-field">
              <label className="search-label">Known languages</label>
              <div className="chip-input-group">
                <input
                  type="text"
                  list="languages-options"
                  value={knownLangInput}
                  onChange={(e) => setKnownLangInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddKnownLang();
                    }
                  }}
                  placeholder="Add language..."
                  className="chip-input"
                />
                <button
                  type="button"
                  onClick={handleAddKnownLang}
                  className="chip-add-btn"
                  disabled={!languages.length}
                  title={!languages.length ? 'List of languages is unavailable' : 'Add language'}
                >
                  +
                </button>
              </div>
              <div className="chips">
                {renderChips(selectedKnownLangIds, 'known')}
              </div>
            </div>

            {/* Learning languages field */}
            <div className="search-field">
              <label className="search-label">Learning languages</label>
              <div className="chip-input-group">
                <input
                  type="text"
                  list="languages-options"
                  value={learnLangInput}
                  onChange={(e) => setLearnLangInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddLearnLang();
                    }
                  }}
                  placeholder="Add language..."
                  className="chip-input"
                />
                <button
                  type="button"
                  onClick={handleAddLearnLang}
                  className="chip-add-btn"
                  disabled={!languages.length}
                  title={!languages.length ? 'List of languages is unavailable' : 'Add language'}
                >
                  +
                </button>
              </div>
              <div className="chips">
                {renderChips(selectedLearnLangIds, 'learn')}
              </div>
            </div>

            <button className="search-button" onClick={handleSearch} disabled={loading}>
              {loading ? 'Search...' : 'Search'}
            </button>

            {/* Datalists for suggestions */}
            <datalist id="hobbies-options">
              {hobbies.map(hobby => (
                <option key={hobby.hobby_id} value={hobby.title} />
              ))}
            </datalist>
            <datalist id="languages-options">
              {languages.map(lang => (
                <option key={lang.language_id} value={lang.name} />
              ))}
            </datalist>
          </div>
        </div>

        <div className="results-area">
          <div className="results-header">Search result</div>
          <Scroll
            style={{ maxHeight: '100%', flex: 1, overflowY: 'auto' }}
          >
              <div className="results-list">
                {loading && <div className="no-results">Loading...</div>}
                {error && <div className="no-results">Error: {error}</div>}
                {!loading && !error && users.length === 0 && (
                  <div className="no-results">Nothing found</div>
                )}
                {!loading && !error && users.map((user) => (
                  <div key={user.user_id} className="chat-item">
                    <div
                      className="chat-avatar"
                      style={{ backgroundColor: user.color || '#333' }}
                    />
                    <div className="chat-info">
                      <div className="chat-name">
                        {user.nickname || user.username || 'User'}
                      </div>
                      <div className="chat-preview">{formatPreview(user)}</div>
                      {user.description && (
                        <div className="chat-preview" style={{ marginTop: 4 }}>
                          {user.description}
                        </div>
                      )}
                    </div>
                    <div className="chat-status"> </div>
                    <button className="chat-create-btn" onClick={() => createChat(user.user_id)}>
                      Create chat
                    </button>
                  </div>
                ))}
            </div>
        </Scroll>
        </div>
      </div>
    </>
  );
};

export { SearchPeople };
