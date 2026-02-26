import { useState, useEffect } from 'react';
import { Check, ChevronDown } from 'lucide-react';
import { categoriesApi } from '@/lib/content-api';
import type { CategoryNode } from '@/lib/content-api';
import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from '@/components/ui/command';
import { cn } from '@/lib/utils';

interface FlatCategory {
  id: string;
  name: string;
  depth: number;
}

function flattenTree(nodes: CategoryNode[], depth = 0): FlatCategory[] {
  const result: FlatCategory[] = [];
  for (const node of nodes) {
    result.push({ id: node.id, name: node.name, depth });
    if (node.children?.length) {
      result.push(...flattenTree(node.children, depth + 1));
    }
  }
  return result;
}

interface CategorySelectProps {
  value: string[];
  onChange: (ids: string[]) => void;
}

export function CategorySelect({ value, onChange }: CategorySelectProps) {
  const [open, setOpen] = useState(false);
  const [categories, setCategories] = useState<FlatCategory[]>([]);

  useEffect(() => {
    categoriesApi.tree().then((res) => {
      if (res?.data) {
        setCategories(flattenTree(res.data));
      }
    });
  }, []);

  const selectedNames = categories
    .filter((c) => value.includes(c.id))
    .map((c) => c.name);

  function toggle(id: string) {
    if (value.includes(id)) {
      onChange(value.filter((v) => v !== id));
    } else {
      onChange([...value, id]);
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="button"
          aria-label="Select categories"
          className="w-full justify-between font-normal"
        >
          <span className="truncate">
            {selectedNames.length > 0
              ? selectedNames.join(', ')
              : 'Select categories'}
          </span>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[280px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search categories..." />
          <CommandList>
            <CommandEmpty>No categories found.</CommandEmpty>
            <CommandGroup>
              {categories.map((cat) => (
                <CommandItem
                  key={cat.id}
                  value={cat.name}
                  onSelect={() => toggle(cat.id)}
                  style={{ paddingLeft: `${(cat.depth * 16) + 8}px` }}
                >
                  <Check
                    data-testid="category-check"
                    className={cn(
                      'mr-2 h-4 w-4',
                      value.includes(cat.id) ? 'opacity-100' : 'opacity-0',
                    )}
                  />
                  {cat.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
