import React, { useState, useEffect } from 'react';
import './SearchPeople.css';
import { api } from "../api.jsx";
import Header from './Header';

// Reusable modal for selecting multiple items from a list
const SelectionModal = ({ isOpen, onClose, items, selectedIds, onConfirm, title }) => {
  const [localSelected, setLocalSelected] = useState(selectedIds);

  // Reset local state when modal opens with new props
  useEffect(() => {
    if (isOpen) {
      setLocalSelected(selectedIds);
    }
  }, [isOpen, selectedIds]);

  if (!isOpen) return null;

  const handleCheckboxChange = (id) => {
    setLocalSelected(prev =>
      prev.includes(id)
        ? prev.filter(itemId => itemId !== id)
        : [...prev, id]
    );
  };

  const handleConfirm = () => {
    onConfirm(localSelected);
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h3>{title}</h3>
        <div className="modal-list">
          {items.map(item => (
            <label key={item.language_id || item.hobby_id} className="modal-item">
              <input
                type="checkbox"
                checked={localSelected.includes(item.language_id || item.hobby_id)}
                onChange={() => handleCheckboxChange(item.language_id || item.hobby_id)}
              />
              {item.name || item.title}
            </label>
          ))}
        </div>
        <div className="modal-actions">
          <button onClick={handleConfirm}>Submit</button>
          <button onClick={onClose}>Cancel</button>
        </div>
      </div>
    </div>
  );
};

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

  // Modal visibility per category
  const [modalState, setModalState] = useState({
    hobby: false,
    knownLang: false,
    learnLang: false,
  });

  // Free‑text search string
  const [searchString, setSearchString] = useState('');

  // Fetch languages and hobbies on mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [langRes, hobbyRes] = await Promise.allSettled([
          api.get('/languages'),
          api.get('/hobbies') // may fail if endpoint doesn't exist
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
        "title": "doesn't matter",
        "goal_id": goalId,
        "type": "Direct"
    })
  }

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

  const handleSearch = async () => {
    setLoading(true);
    setError(null);

    const filter = {};

    if (searchString) filter.search_string = searchString;

    // Use selected IDs directly
    if (selectedHobbyIds.length > 0) {
      filter.hobbies = selectedHobbyIds;
    }

    if (selectedKnownLangIds.length > 0) {
      filter.known_languages = selectedKnownLangIds;
    }

    if (selectedLearnLangIds.length > 0) {
      filter.learn_languages = selectedLearnLangIds;
    }

    // Pagination defaults
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
      parts.push(`Learnt languages: ${langNames}`);
    }

    return parts.join(' • ');
  };

  // Render selected items as tags with remove button
  const renderTags = (ids, type) => {
    const items = type === 'hobby'
      ? ids.map(id => ({ id, name: getHobbyTitle(id) }))
      : ids.map(id => ({ id, name: getLanguageName(id) }));

    return items.map(item => (
      <span key={item.id} className="tag">
        {item.name}
        <button
          className="tag-remove"
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
          ✕
        </button>
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

            {/* Hobbies field with tags and plus button */}
            <div className="search-field">
              <label className="search-label">Hobbies</label>
              <div className="selection-container">
                <div className="tags-container">
                  {renderTags(selectedHobbyIds, 'hobby')}
                </div>
                <button
                  className="plus-button"
                  onClick={() => setModalState(prev => ({ ...prev, hobby: true }))}
                  disabled={!hobbies.length}
                  title={!hobbies.length ? 'List of hobbies is unavailable' : 'Add hobby'}
                >
                  +
                </button>
              </div>
            </div>

            {/* Known languages field */}
            <div className="search-field">
              <label className="search-label">Known languages</label>
              <div className="selection-container">
                <div className="tags-container">
                  {renderTags(selectedKnownLangIds, 'known')}
                </div>
                <button
                  className="plus-button"
                  onClick={() => setModalState(prev => ({ ...prev, knownLang: true }))}
                  disabled={!languages.length}
                  title={!languages.length ? 'List of languages is unavailable' : 'Add language'}
                >
                  +
                </button>
              </div>
            </div>

            {/* Learning languages field */}
            <div className="search-field">
              <label className="search-label">Learning langugaes</label>
              <div className="selection-container">
                <div className="tags-container">
                  {renderTags(selectedLearnLangIds, 'learn')}
                </div>
                <button
                  className="plus-button"
                  onClick={() => setModalState(prev => ({ ...prev, learnLang: true }))}
                  disabled={!languages.length}
                  title={!languages.length ? 'List of languages is unavailable' : 'Add language'}
                >
                  +
                </button>
              </div>
            </div>

            <button className="search-button" onClick={handleSearch} disabled={loading}>
              {loading ? 'Search...' : 'Search'}
            </button>
          </div>
        </div>

        <div className="results-area">
          <div className="results-header">Search result</div>
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
                  style={{
                    backgroundColor: user.color || '#333',
                  }}
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
                <button className="chat-create-btn" onClick={() => {
                    createChat(user.user_id);
                }}>Create chat</button>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Modals for selection */}
      <SelectionModal
        isOpen={modalState.hobby}
        onClose={() => setModalState(prev => ({ ...prev, hobby: false }))}
        items={hobbies}
        selectedIds={selectedHobbyIds}
        onConfirm={setSelectedHobbyIds}
        title="Choose hobbies"
      />

      <SelectionModal
        isOpen={modalState.knownLang}
        onClose={() => setModalState(prev => ({ ...prev, knownLang: false }))}
        items={languages}
        selectedIds={selectedKnownLangIds}
        onConfirm={setSelectedKnownLangIds}
        title="Choose known languages"
      />

      <SelectionModal
        isOpen={modalState.learnLang}
        onClose={() => setModalState(prev => ({ ...prev, learnLang: false }))}
        items={languages}
        selectedIds={selectedLearnLangIds}
        onConfirm={setSelectedLearnLangIds}
        title="Choose learnt languages"
      />
    </>
  );
};

export { SearchPeople };
