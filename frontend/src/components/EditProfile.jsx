import React, { useState, useEffect } from 'react';
import './EditProfile.css';
import { api } from "../api.jsx";
// import Logo from '../assets/Logo.svg'; // not used
import Settings from '../assets/settings.svg';
import Header from './Header';

const EditProfile = ({
  initialName = '',
  initialColor = '',
  initialDescription = '',
  initialHobbyTitles = [],
  initialKnownLanguageNames = [],
  initialLearnLanguageNames = [],
  onSave,
  onCancel,
}) => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [userId, setUserId] = useState(null);

  // Modal state
  const [isOpen, setIsOpen] = useState(false);
  const [openedParam, setOpenedParam] = useState('');

  // Form fields
  const [color, setColor] = useState(initialColor);
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);

  // Lists (as titles for display)
  const [hobbies, setHobbies] = useState(initialHobbyTitles);
  const [knownLanguages, setKnownLanguages] = useState(initialKnownLanguageNames);
  const [learnLanguages, setLearnLanguages] = useState(initialLearnLanguageNames);

  // Master data for mapping
  const [allHobbies, setAllHobbies] = useState([]);       // { id, title }
  const [allLanguages, setAllLanguages] = useState([]);   // { id, name }

  // Initial IDs for change detection
  const [initialHobbyIds, setInitialHobbyIds] = useState([]);
  const [initialKnownLanguageIds, setInitialKnownLanguageIds] = useState([]);
  const [initialLearnLanguageIds, setInitialLearnLanguageIds] = useState([]);

  // Input fields for new items
    const [newHobby, setNewHobby] = useState('');
    const [newKnownLang, setNewKnownLang] = useState('');
    const [newLearnLang, setNewLearnLang] = useState('');


  const onSignOut = async () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user_id");
    localStorage.removeItem("refresh_token");
    window.location.reload();
  }

  useEffect(() => {
    const token = localStorage.getItem('token');
    const storedUserId = localStorage.getItem('user_id');

    if (!token || !storedUserId) {
      setError('Authentication required. Please log in.');
      setLoading(false);
      return;
    }

    setUserId(parseInt(storedUserId));

    const fetchData = async () => {
      try {
        // Fetch current user
        const userRes = await api.post('/users', {
          user_id: parseInt(storedUserId), page_size: 1,
        });

        console.log(userRes);

        const users = userRes.data;
        const user = users[0];

        // Populate form
        setName(user.nickname || '');
        setColor(user.color || '');
        setDescription(user.description || '');
        setInitialHobbyIds(user.hobby_title_ids || []);
        setInitialKnownLanguageIds(user.known_language_ids || []);
        setInitialLearnLanguageIds(user.learn_language_ids || []);

        // Fetch master data (hobbies and languages)
        const [hobbiesRes, languagesRes] = await Promise.all([
          api.get('/hobbies'),
          api.get('/languages'),
        ]);

        const hobbiesData = await hobbiesRes.data;
        const languagesData = await languagesRes.data;

        setAllHobbies(hobbiesData);
        setAllLanguages(languagesData);

        // Map IDs to titles for display
        setHobbies(
          hobbiesData
            .filter(h => user.hobby_title_ids.includes(h.hobby_id))
            .map(h => h.title)
        );
        setKnownLanguages(
          languagesData
            .filter(l => user.known_language_ids.includes(l.language_id))
            .map(l => l.name)
        );
        setLearnLanguages(
          languagesData
            .filter(l => user.learn_language_ids.includes(l.language_id))
            .map(l => l.name)
        );
        console.log("hobbies: ", hobbies);
        console.log("hobbiesData: ", hobbiesData);
        console.log("user.hobby_title_id: ", user.hobby_title_ids);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const addItem = (value, setter, inputSetter) => {
    if (value.trim()) {
      setter(prev => [...prev, value.trim()]);
      inputSetter('');
    }
  };

  const removeItem = (array, index, setter) => {
    setter(array.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const token = localStorage.getItem('token');
    if (!token || !userId) {
      alert('Authentication missing');
      return;
    }

    // Helper to get added/removed IDs based on titles
    const getChanges = (currentTitles, initialIds, allItems, titleKey, idKey) => {
      console.log(currentTitles, initialIds, allItems, titleKey, idKey);
      const currentIds = currentTitles
        .map(title => {
          const item = allItems.find(i => i[titleKey] === title);
          return item ? item[idKey] : null;
        })
        .filter(id => id !== null);
      
      console.log(currentIds);

      const added = currentIds.filter(id => !initialIds.includes(id));
      const removed = initialIds.filter(id => !currentIds.includes(id));

      console.log(added, removed);
      return { added, removed };
    };

    const hobbyChanges = getChanges(hobbies, initialHobbyIds, allHobbies, 'title', 'hobby_id');
    const knownLangChanges = getChanges(knownLanguages, initialKnownLanguageIds, allLanguages, 'name', 'language_id');
    const learnLangChanges = getChanges(learnLanguages, initialLearnLanguageIds, allLanguages, 'name', 'language_id');

    const payload = {
      description,
      color,
      nickname: name,
      add_hobbies: hobbyChanges.added.map(id => ({ hobby_id: id })),
      delete_hobbies: hobbyChanges.removed.map(id => ({ hobby_id: id })),
      add_languages: [
        ...knownLangChanges.added.map(id => ({ language_id: id, is_known: true })),
        ...learnLangChanges.added.map(id => ({ language_id: id, is_known: false })),
      ],
      delete_languages: [
        ...knownLangChanges.removed.map(id => ({ language_id: id, is_known: true })),
        ...learnLangChanges.removed.map(id => ({ language_id: id, is_known: false })),
      ],
    };

      const response = await api.patch(`/users/${userId}`, 
        payload,
      );

      if (onSave) onSave();
  };

  const openWindow = (param) => {
    setOpenedParam(param);
    setIsOpen(true);
  };

  const closeModal = () => {
    setIsOpen(false);
    setOpenedParam('');
  };

  // Render chips with remove button
  const renderRemovableChips = (items, setter, isLearn = false) => {
    return items.map((item, index) => (
      <span key={index} className={`chip ${isLearn ? 'learn' : ''}`}>
        {item}
        <span className="chip-close" onClick={() => removeItem(items, index, setter)}>
          ×
        </span>
      </span>
    ));
  };

  if (loading) return <div className="profile-background">Loading...</div>;
  if (error) return <div className="profile-background">Error: {error}</div>;

  return (
    <>
      <Header />
      <div className="profile-background">
        <div className="profile-card">
          <div className="profile-content">
            <form onSubmit={handleSubmit}>
              {/* Name and color */}
              <div className="section name-section">
                <div className="color-wrapper">
                  <input
                    type="color"
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    className="view-profile"
                  />
                </div>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Enter nickname..."
                  className="name-input"
                />
              </div>

              {/* About */}
              <div className="section">
                <div className="section-title">О себе</div>
                <textarea
                  className="description-textarea"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Enter your description..."
                  rows={5}
                />
              </div>

              {/* Hobbies */}
              <div className="section">
                <div className="section-title">Hobbies</div>
                <div>
                  <div className="chip-input-group">
                    <input
                      type="text"
                      list="hobbies-options"
                      value={newHobby}
                      onChange={(e) => setNewHobby(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          addItem(newHobby, setHobbies, setNewHobby);
                        }
                      }}
                      placeholder="Add hobby..."
                      className="chip-input"
                    />
                    <button
                      type="button"
                      onClick={() => addItem(newHobby, setHobbies, setNewHobby)}
                      className="chip-add-btn"
                    >
                      <img src={Settings} className="settings-icon" alt="add" />
                    </button>
                  </div>
                  <div className="chips">{renderRemovableChips(hobbies, setHobbies)}</div>
                </div>
              </div>

              {/* Known languages */}
              <div className="section">
                <div className="section-title">Known languages</div>
                <div>
                  <div className="chip-input-group">
                    <input
                      type="text"
                      list="languages-options"
                      value={newKnownLang}
                      onChange={(e) => setNewKnownLang(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          addItem(newKnownLang, setKnownLanguages, setNewKnownLang);
                        }
                      }}
                      placeholder="Enter language..."
                      className="chip-input"
                    />
                    <button
                      type="button"
                      onClick={() => addItem(newKnownLang, setKnownLanguages, setNewKnownLang)}
                      className="chip-add-btn"
                    >
                      <img src={Settings} className="settings-icon" alt="add" />
                    </button>
                  </div>
                  <div className="chips">{renderRemovableChips(knownLanguages, setKnownLanguages)}</div>
                </div>
              </div>

              {/* Learning languages */}
              <div className="section">
                <div className="section-title">Learn languages</div>
                <div>
                  <div className="chip-input-group">
                    <input
                      list="languages-options"
                      type="text"
                      value={newLearnLang}
                      onChange={(e) => setNewLearnLang(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          addItem(newLearnLang, setLearnLanguages, setNewLearnLang);
                        }
                      }}
                      placeholder="Add language..."
                      className="chip-input"
                    />
                    <button
                      type="button"
                      onClick={() => addItem(newLearnLang, setLearnLanguages, setNewLearnLang)}
                      className="chip-add-btn"
                    >
                      <img src={Settings} className="settings-icon" alt="add" />
                    </button>
                  </div>
                  <div className="chips">{renderRemovableChips(learnLanguages, setLearnLanguages, true)}</div>
                </div>
              </div>

              {/* Footer buttons */}
              <div className="profile-footer" style={{ display: 'flex', gap: '16px' }}>
                <button type="submit" className="edit-btn">
                  Save changes
                </button>
                <button type="button" className="edit-btn cancel-btn" onClick={onCancel}>
                  Cancel
                </button>
                <button type="button" className="edit-btn signout-btn" onClick={onSignOut}>
                  Sign out
                </button>
              </div>
                <datalist id="hobbies-options">
                  {allHobbies.map((item, index) => (
                    <option value={item.title} />
                  ))}
                </datalist>
                <datalist id="languages-options">
                  {allLanguages.map((item, index) => (
                    <option value={item.name} />
                  ))}
                </datalist>
            </form>
          </div>
        </div>

        {/* Modal for adding items (currently not used but can be extended) */}
        {isOpen && (
          <>
            <div className="edit-background" onClick={closeModal} />
            <div className="edit-card">
              <form>
                <div className="section">
                  <div className="section-title">
                    {openedParam === 'hobbies' && 'Хобби'}
                    {openedParam === 'known_languages' && 'Знаю языки'}
                    {openedParam === 'learn_languages' && 'Изучаю языки'}
                  </div>
                  <div className="chip-input-group">
                    <input
                      placeholder="Добавить..."
                      className="chip-input"
                      type="text"
                    />
                    <button type="button" className="chip-add-btn">+</button>
                  </div>
                  <div className="chips">
                    {/* Dynamically render chips based on openedParam */}
                  </div>
                </div>
                <div className="profile-footer" style={{ display: 'flex', gap: '16px' }}>
                  <button type="submit" className="edit-btn">Сохранить изменения</button>
                  <button type="button" className="edit-btn cancel-btn" onClick={closeModal}>
                    Отмена
                  </button>
                </div>
              </form>
            </div>
          </>
        )}
      </div>
    </>
  );
};

export { EditProfile };
