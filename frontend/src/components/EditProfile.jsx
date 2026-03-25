import React, { useState, useEffect } from 'react';
import './EditProfile.css';
import { api } from "../api.jsx";
import { useNavigate } from 'react-router-dom';
import Header from './Header';

const EditProfile = ({
  initialName = '',
  initialColor = '',
  initialDescription = '',
  initialPassword = '',
  initialPasswordConfirmation = '',
  initialHobbyTitles = [],
  initialKnownLanguageNames = [],
  initialLearnLanguageNames = [],
  onSave,
}) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [userId, setUserId] = useState(null);
  const [activeTab, setActiveTab] = useState('common'); // 'common' or 'security'

  // Form fields
  const [color, setColor] = useState(initialColor);
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
  const [password, setPassword] = useState(initialPassword);
  const [passwordConfirmation, setPasswordConfirmation] = useState(initialPasswordConfirmation);

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

  // Password validation error
  const [passwordError, setPasswordError] = useState('');

  const onSignOut = async () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user_id");
    localStorage.removeItem("refresh_token");
    window.location.reload();
  };

  const onCancel = async () => {
    navigate("/");
    window.location.reload();
  };

  const fetchData = async () => {
    try {
      const userRes = await api.get('/my');
      const user = userRes.data;

      setName(user.nickname || '');
      setColor(user.color || '');
      setDescription(user.description || '');
      setInitialHobbyIds(user.hobby_title_ids || []);
      setInitialKnownLanguageIds(user.known_language_ids || []);
      setInitialLearnLanguageIds(user.learn_language_ids || []);

      const [hobbiesRes, languagesRes] = await Promise.all([
        api.get('/hobbies'),
        api.get('/languages'),
      ]);

      const hobbiesData = hobbiesRes.data;
      const languagesData = languagesRes.data;

      setAllHobbies(hobbiesData);
      setAllLanguages(languagesData);

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
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    const storedUserId = localStorage.getItem('user_id');

    if (!token || !storedUserId) {
      setError('Authentication required. Please log in.');
      setLoading(false);
      return;
    }

    setUserId(parseInt(storedUserId));
    fetchData();
  }, []);

  const addItem = (value, setter, inputSetter, all, key) => {
    if (value.trim()) {
      if (all.map(el => el[key]).includes(value)) {
        setter(prev => [...prev, value.trim()]);
        inputSetter('');
      }
    }
  };

  const removeItem = (array, index, setter) => {
    setter(array.filter((_, i) => i !== index));
  };

  const handleCommonSubmit = async (e) => {
    e.preventDefault();
    setError(null);

    const token = localStorage.getItem('token');
    if (!token || !userId) {
      setError('Authentication missing');
      return;
    }

    const getChanges = (currentTitles, initialIds, allItems, titleKey, idKey) => {
      const currentIds = currentTitles
        .map(title => {
          const item = allItems.find(i => i[titleKey] === title);
          return item ? item[idKey] : null;
        })
        .filter(id => id !== null);

      const added = currentIds.filter(id => !initialIds.includes(id));
      const removed = initialIds.filter(id => !currentIds.includes(id));
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

    try {
      await api.patch(`/users/${userId}`, payload);
      await fetchData();
      //if (onSave) onSave();
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to save changes');
    }
  };

  const handleSecuritySubmit = async (e) => {
    e.preventDefault();
    setPasswordError('');
    setError(null);

    // Validate passwords
    if (password || passwordConfirmation) {
      if (password !== passwordConfirmation) {
        setPasswordError('Passwords do not match');
        return;
      }
      if (password.length < 2) {
        setPasswordError('Password must be at least 6 characters');
        return;
      }
    } else {
      // No password change requested
      setError('No password provided');
      return;
    }

    try {
      await api.patch(`/users/${userId}`, { password });
      // Clear fields on success
      setPassword('');
      setPasswordConfirmation('');
      //if (onSave) onSave();
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to update password');
    }
  };

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

  if (loading) return <div>Loading...</div>;

  return (
    <>
      <Header />
      <div className="profile-background">
        <div className="profile-card">
          {/* Left tabs */}
          <div className="profile-tabs">
            <button
              className={`tab-button ${activeTab === 'common' ? 'active' : ''}`}
              onClick={() => setActiveTab('common')}
            >
              Common
            </button>
            <button
              className={`tab-button ${activeTab === 'security' ? 'active' : ''}`}
              onClick={() => setActiveTab('security')}
            >
              Security
            </button>
          </div>

          {/* Right content with fade animation (key triggers re-mount) */}
          <div key={activeTab} className="profile-tab-content">
            {activeTab === 'common' ? (
              <form onSubmit={handleCommonSubmit}>
                {/* Color */}
                <div className="section name-section">
                  <div className="color-wrapper">
                    <input
                      type="color"
                      value={color}
                      onChange={(e) => setColor(e.target.value)}
                      className="view-profile"
                    />
                  </div>
                </div>

                {/* Name */}
                <div className="section name-section">
                  <div className="section-title">Nickname</div>
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
                  <div className="section-title">Bio</div>
                  <textarea
                    className="description-textarea"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Enter your description..."
                    rows={3}
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
                            addItem(newHobby, setHobbies, setNewHobby, allHobbies, "title");
                          }
                        }}
                        placeholder="Add hobby..."
                        className="chip-input"
                      />
                      <button
                        type="button"
                        onClick={() => addItem(newHobby, setHobbies, setNewHobby, allHobbies, "title")}
                        className="chip-add-btn"
                      >
                        +
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
                            addItem(newKnownLang, setKnownLanguages, setNewKnownLang, allLanguages, "name");
                          }
                        }}
                        placeholder="Enter language..."
                        className="chip-input"
                      />
                      <button
                        type="button"
                        onClick={() => addItem(newKnownLang, setKnownLanguages, setNewKnownLang, allLanguages, "name")}
                        className="chip-add-btn"
                      >
                        +
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
                            addItem(newLearnLang, setLearnLanguages, setNewLearnLang, allLanguages, "name");
                          }
                        }}
                        placeholder="Add language..."
                        className="chip-input"
                      />
                      <button
                        type="button"
                        onClick={() => addItem(newLearnLang, setLearnLanguages, setNewLearnLang, allLanguages, "name")}
                        className="chip-add-btn"
                      >
                        +
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
                  {allHobbies.map((item) => (
                    <option key={item.hobby_id} value={item.title} />
                  ))}
                </datalist>
                <datalist id="languages-options">
                  {allLanguages.map((item) => (
                    <option key={item.language_id} value={item.name} />
                  ))}
                </datalist>
              </form>
            ) : (
              <form onSubmit={handleSecuritySubmit}>
                {/* Password */}
                <div className="section password-section">
                  <div className="section-title">New Password</div>
                  <input
                    type="password"
                    className="password-input"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter new password (min. 6 characters)"
                  />
                </div>

                {/* Password confirmation */}
                <div className="section password-section">
                  <div className="section-title">Confirm Password</div>
                  <input
                    type="password"
                    className="password-input"
                    value={passwordConfirmation}
                    onChange={(e) => setPasswordConfirmation(e.target.value)}
                    placeholder="Re-enter new password"
                  />
                </div>

                {/* Footer buttons */}
                <div className="profile-footer" style={{ display: 'flex', gap: '16px' }}>
                  <button type="submit" className="edit-btn">
                    Update password
                  </button>
                  <button type="button" className="edit-btn cancel-btn" onClick={onCancel}>
                    Cancel
                  </button>
                  <button type="button" className="edit-btn signout-btn" onClick={onSignOut}>
                    Sign out
                  </button>
                </div>

                {/* Password-specific error */}
                {passwordError && <div className="error-message">{passwordError}</div>}
              </form>
            )}

            {/* General error display (appears in both tabs) */}
            {error && <div className="error-message">{error}</div>}
          </div>
        </div>
      </div>
    </>
  );
};

export {EditProfile};
