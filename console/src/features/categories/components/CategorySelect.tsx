import { useEffect, useState } from 'react';
import { ChevronDown } from 'lucide-react';
import type { Category } from '../../types/categories';

interface CategorySelectProps {
  categories: Category[];
  value?: string;
  onChange: (categoryId: string | undefined) => void;
  placeholder?: string;
  disabled?: boolean;
}

interface FlatCategory extends Category {
  depth: number;
}

function flattenCategories(categories: Category[], depth: number = 0): FlatCategory[] {
  const result: FlatCategory[] = [];
  for (const category of categories) {
    result.push({ ...category, depth });
    if (category.children && category.children.length > 0) {
      result.push(...flattenCategories(category.children, depth + 1));
    }
  }
  return result;
}

export function CategorySelect({
  categories,
  value,
  onChange,
  placeholder = 'Select a category',
  disabled = false,
}: CategorySelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const flatCategories = flattenCategories(categories);

  const selectedCategory = flatCategories.find((c) => c.id === value);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('[data-category-select]')) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [isOpen]);

  return (
    <div data-category-select className="relative">
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <span className={selectedCategory ? '' : 'text-muted-foreground'}>
          {selectedCategory
            ? `${'  '.repeat(selectedCategory.depth)}${selectedCategory.name}`
            : placeholder}
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
            {flatCategories.map((category) => (
              <button
                key={category.id}
                type="button"
                onClick={() => {
                  onChange(category.id);
                  setIsOpen(false);
                }}
                className={`flex w-full items-center px-3 py-2 text-sm hover:bg-muted ${
                  value === category.id ? 'bg-muted' : ''
                }`}
              >
                <span style={{ paddingLeft: `${category.depth * 16}px` }}>
                  {category.name}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
