import { useEffect, useState } from 'react';
import { ChevronDown } from 'lucide-react';
import type { Tag } from '../../types/tags';

interface TagSelectProps {
  tags: Tag[];
  value?: string;
  onChange: (tagId: string | undefined) => void;
  placeholder?: string;
  disabled?: boolean;
}

export function TagSelect({
  tags,
  value,
  onChange,
  placeholder = 'Select a tag',
  disabled = false,
}: TagSelectProps) {
  const [isOpen, setIsOpen] = useState(false);

  const selectedTag = tags.find((t) => t.id === value);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('[data-tag-select]')) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [isOpen]);

  return (
    <div data-tag-select className="relative">
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <span className={selectedTag ? '' : 'text-muted-foreground'}>
          {selectedTag ? selectedTag.name : placeholder}
        </span>
        <ChevronDown className="h-4 w-4 opacity-50" />
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-background shadow-md">
          <div className="py-1">
            <button
              type="button"
              onClick={() => {
                onChange(undefined);
                setIsOpen(false);
              }}
              className="flex w-full items-center px-3 py-2 text-sm hover:bg-muted"
            >
              <span className="text-muted-foreground">{placeholder}</span>
            </button>
            {tags.map((tag) => (
              <button
                key={tag.id}
                type="button"
                onClick={() => {
                  onChange(tag.id);
                  setIsOpen(false);
                }}
                className={`flex w-full items-center px-3 py-2 text-sm hover:bg-muted ${
                  value === tag.id ? 'bg-muted' : ''
                }`}
              >
                {tag.name}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
