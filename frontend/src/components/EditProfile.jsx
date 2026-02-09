import React, { useState } from 'react';
import './EditProfile.css';
import Logo from '../assets/Logo.svg';
import Settings from '../assets/settings.svg';

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
  
  const [isOpen, setIsOpen] = useState(false);
  const [openedParam, setOpenedParam] = useState('');
  
  const [color, setColor] = useState(initialColor);
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);

  const [hobbies, setHobbies] = useState(initialHobbyTitles);
  const [knownLanguages, setKnownLanguages] = useState(initialKnownLanguageNames);
  const [learnLanguages, setLearnLanguages] = useState(initialLearnLanguageNames);

  const [newHobby, setNewHobby] = useState('');
  const [newKnownLang, setNewKnownLang] = useState('');
  const [newLearnLang, setNewLearnLang] = useState('');

  const addItem = (value, setter, inputSetter) => {
    if (value.trim()) {
      setter(prev => [...prev, value.trim()]);
      inputSetter('');
    }
  };

  const removeItem = (array, index, setter) => {
    setter(array.filter((_, i) => i !== index));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (onSave) {
      onSave({
        description,
        hobbyTitles: hobbies,
        knownLanguageNames: knownLanguages,
        learnLanguageNames: learnLanguages,
      });
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

  const openWindow = (name) => {
    setIsOpen(true);
    if (name === "known_languages") {
      
    }
    return;
  }

  return (
    <>
    <header className="top-header">
      <img src={Logo} alt="Logo" className="logo" />
      <button className="profile-btn" aria-label="Профиль" />
    </header>
    <div className="profile-background">
      <div className="profile-card">


        <div className="profile-content">
          <form onSubmit={handleSubmit}>
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
                  value={newHobby}
                  onChange={(e) => setNewHobby(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      addItem(newHobby, setHobbies, setNewHobby);
                    }
                  }}
                  placeholder="Введите имя..."
                  className="name-input"
                />
            </div>
            <div className="section">
              <div className="section-title">О себе</div>
              <textarea
                className="description-textarea"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Расскажите о себе, интересах и целях изучения языков..."
                rows={5}
              />
            </div>

            <div className="section">
              <div className="section-title">Хобби</div>
              <div className="chip-input-group">
                <button
                  type="button"
                  onClick={() => addItem(newHobby, setHobbies, setNewHobby)}
                  className="chip-add-btn"
                >
                  <img src={Settings} className="settings-icon" />
                </button>
              </div>
              
            </div>

            <div className="section">
              <div className="section-title">Знаю языки</div>
              <div className="chip-input-group">
                <button
                  type="button"
                  onClick={() => openWindow("known_languages")}
                  className="chip-add-btn"
                >
                  <img src={Settings} className="settings-icon" />
                </button>
              </div>

            </div>

            <div className="section">
              <div className="section-title">Изучаю языки</div>
              <div className="chip-input-group">
                <button
                  type="button"
                  onClick={() => addItem(newLearnLang, setLearnLanguages, setNewLearnLang)}
                  className="chip-add-btn"
                >
                  <img src={Settings} className="settings-icon" />
                </button>
              </div>

            </div>

            <div className="profile-footer" style={{ display: 'flex', gap: '16px' }}>
              <button type="submit" className="edit-btn">
                Сохранить изменения
              </button>
              <button type="button" className="edit-btn cancel-btn" onClick={onCancel}>
                Отмена
              </button>
            </div>
          </form>
        </div>
      </div>
    {isOpen && (
<>
<div className="edit-background" />
<div className="edit-card">
  <form className="">
    <div className="section">
      <div className="section-title">Хобби</div>
      <div className="chip-input-group">
        <input
          placeholder="Добавить хобби..."
          className="chip-input"
          type="text"
        />
        <button type="button" className="chip-add-btn">+</button>
      </div>
      <div className="chips">
        <span className="chip">
          hdbvfl
          <span className="chip-close">×</span>
        </span>
      </div>
    </div>
    <div className="profile-footer" style={{ display: "flex", gap: "16px" }}>
      <button type="submit" className="edit-btn">Сохранить изменения</button>
      <button type="button" className="edit-btn cancel-btn">Отмена</button>
    </div>
  </form>
</div>
</>
)}
    </div>
</>
  );
};

export {EditProfile};
