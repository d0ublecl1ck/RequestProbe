import React, { useEffect, useMemo, useState } from 'react';
import { ChevronDown, ExternalLink, Loader2 } from 'lucide-react';

export function OpenerSelect({
  options,
  selectedValue,
  onSelect,
  onOpen,
  disabled,
  loadingValues = {},
}) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const currentOpener = useMemo(
    () => options.find((item) => item.value === selectedValue) || options[0],
    [options, selectedValue],
  );
  const CurrentOpenerIcon = currentOpener.icon;
  const isOpeningCurrent = Boolean(loadingValues[currentOpener.value]);

  useEffect(() => {
    if (!isMenuOpen) {
      return undefined;
    }

    const handlePointerDown = (event) => {
      const menuRoot = document.querySelector('[data-opener-select-root="true"]');
      if (menuRoot && !menuRoot.contains(event.target)) {
        setIsMenuOpen(false);
      }
    };

    window.addEventListener('pointerdown', handlePointerDown);
    return () => window.removeEventListener('pointerdown', handlePointerDown);
  }, [isMenuOpen]);

  return (
    <div className="relative" data-opener-select-root="true">
      <div className="opener-select-shell">
        <button
          type="button"
          className="opener-select-main"
          onClick={() => onOpen(currentOpener.value)}
          disabled={disabled || Object.values(loadingValues).some(Boolean)}
        >
          <span className="opener-select-icon">
            {isOpeningCurrent ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <CurrentOpenerIcon className="h-4 w-4" />
            )}
          </span>
          <span className="opener-select-label">{currentOpener.label}</span>
        </button>
        <button
          type="button"
          className="opener-select-toggle"
          onClick={() => setIsMenuOpen((prev) => !prev)}
          disabled={disabled}
          aria-haspopup="menu"
          aria-expanded={isMenuOpen}
          aria-label="选择打开方式"
        >
          <ChevronDown className={`h-4 w-4 transition-transform ${isMenuOpen ? 'rotate-180' : ''}`} />
        </button>
      </div>

      {isMenuOpen ? (
        <div className="opener-select-menu" role="menu">
          {options.map((opener) => {
            const OpenerIcon = opener.icon;
            const isActive = opener.value === currentOpener.value;

            return (
              <button
                key={opener.value}
                type="button"
                role="menuitem"
                className={`opener-select-option ${isActive ? 'opener-select-option-active' : ''}`}
                onClick={() => {
                  onSelect(opener.value);
                  setIsMenuOpen(false);
                }}
              >
                <span className="opener-select-option-icon">
                  <OpenerIcon className="h-4 w-4" />
                </span>
                <span className="opener-select-option-text">
                  <span className="opener-select-option-title">{opener.label}</span>
                  <span className="opener-select-option-subtitle">{opener.subtitle}</span>
                </span>
                {isActive ? <ExternalLink className="opener-select-option-mark" /> : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}
