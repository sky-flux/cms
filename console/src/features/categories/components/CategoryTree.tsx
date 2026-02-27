import { useState } from 'react';
import { ChevronRight, ChevronDown, Pencil, Trash } from 'lucide-react';
import type { Category } from '../../types/categories';

interface CategoryTreeProps {
  categories: Category[];
  onEdit: (categoryId: string) => void;
  onDelete: (categoryId: string) => void;
  onSelect: (categoryId: string) => void;
}

interface CategoryItemProps {
  category: Category;
  depth: number;
  onEdit: (categoryId: string) => void;
  onDelete: (categoryId: string) => void;
  onSelect: (categoryId: string) => void;
}

function CategoryItem({ category, depth, onEdit, onDelete, onSelect }: CategoryItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasChildren = category.children && category.children.length > 0;

  return (
    <div>
      <div
        className={`flex items-center justify-between py-2 px-4 hover:bg-muted/50 cursor-pointer ${
          depth === 0 ? 'pl-4' : `pl-${(depth + 1) * 4}`
        }`}
      >
        <div className="flex items-center gap-2">
          {hasChildren ? (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="p-1 hover:bg-muted rounded"
              aria-label={isExpanded ? 'collapse' : 'expand'}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </button>
          ) : (
            <span className="w-6" />
          )}
          <span
            onClick={() => onSelect(category.id)}
            className="font-medium"
          >
            {category.name}
          </span>
          {category.description && (
            <span className="text-sm text-muted-foreground">- {category.description}</span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => onEdit(category.id)}
            className="p-2 hover:bg-muted rounded-md"
            title="Edit"
            aria-label="Edit"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={() => onDelete(category.id)}
            className="p-2 hover:bg-muted rounded-md text-destructive"
            title="Delete"
            aria-label="Delete"
          >
            <Trash className="h-4 w-4" />
          </button>
        </div>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {category.children!.map((child) => (
            <CategoryItem
              key={child.id}
              category={child}
              depth={depth + 1}
              onEdit={onEdit}
              onDelete={onDelete}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function CategoryTree({ categories, onEdit, onDelete, onSelect }: CategoryTreeProps) {
  if (categories.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No categories found.
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <div className="bg-muted/50 py-2 px-4">
        <div className="grid grid-cols-12 gap-4 text-sm font-medium">
          <div className="col-span-8">Name</div>
          <div className="col-span-4 text-right">Actions</div>
        </div>
      </div>
      <div>
        {categories.map((category) => (
          <CategoryItem
            key={category.id}
            category={category}
            depth={0}
            onEdit={onEdit}
            onDelete={onDelete}
            onSelect={onSelect}
          />
        ))}
      </div>
    </div>
  );
}
