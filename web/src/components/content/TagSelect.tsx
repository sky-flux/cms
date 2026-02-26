import { useState, useEffect, useCallback, useRef } from 'react';
import { Plus, X } from 'lucide-react';
import { tagsApi } from '@/lib/content-api';
import type { Tag } from '@/lib/content-api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from '@/components/ui/command';

interface TagSelectProps {
  value: string[];
  onChange: (ids: string[]) => void;
  allTags?: Tag[];
}

export function TagSelect({ value, onChange, allTags = [] }: TagSelectProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [suggestions, setSuggestions] = useState<Tag[]>([]);
  const [creating, setCreating] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Resolve tag names from value IDs
  const resolvedTags = allTags.filter((t) => value.includes(t.id));
  // Also include suggestions that match value IDs (for newly created tags)
  const allKnown = [...allTags, ...suggestions];
  const selectedTags = value.map((id) => {
    const found = allKnown.find((t) => t.id === id);
    return found || { id, name: id, slug: '', post_count: 0, created_at: '' };
  });

  const fetchSuggestions = useCallback((q: string) => {
    if (!q.trim()) {
      setSuggestions([]);
      return;
    }
    tagsApi.suggest(q).then((res) => {
      if (res?.data) {
        setSuggestions(res.data);
      }
    });
  }, []);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      fetchSuggestions(search);
    }, 200);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [search, fetchSuggestions]);

  function selectTag(id: string) {
    if (!value.includes(id)) {
      onChange([...value, id]);
    }
    setOpen(false);
    setSearch('');
  }

  function removeTag(id: string) {
    onChange(value.filter((v) => v !== id));
  }

  async function createAndSelect(name: string) {
    setCreating(true);
    try {
      const res = await tagsApi.create({ name });
      if (res?.data) {
        setSuggestions((prev) => [...prev, res.data]);
        selectTag(res.data.id);
      }
    } finally {
      setCreating(false);
    }
  }

  const hasExactMatch = suggestions.some(
    (s) => s.name.toLowerCase() === search.toLowerCase(),
  );
  const showCreate = search.trim().length > 0 && !hasExactMatch;
  // Filter out already-selected tags from suggestions
  const filteredSuggestions = suggestions.filter((s) => !value.includes(s.id));

  return (
    <div className="flex flex-col gap-2">
      {/* Selected tags as badges */}
      {selectedTags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {selectedTags.map((tag) => (
            <Badge key={tag.id} variant="secondary" className="gap-1">
              {tag.name}
              <button
                type="button"
                aria-label={`Remove ${tag.name}`}
                onClick={() => removeTag(tag.id)}
                className="ml-1 rounded-full outline-none hover:bg-muted"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}

      {/* Add tag popover */}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            role="button"
            aria-label="Add tag"
            className="w-fit"
          >
            <Plus className="mr-1 h-4 w-4" />
            Add tag
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[250px] p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              placeholder="Search tags..."
              value={search}
              onValueChange={setSearch}
            />
            <CommandList>
              {filteredSuggestions.length === 0 && !showCreate && (
                <CommandEmpty>No tags found.</CommandEmpty>
              )}
              <CommandGroup>
                {filteredSuggestions.map((tag) => (
                  <CommandItem
                    key={tag.id}
                    value={tag.name}
                    onSelect={() => selectTag(tag.id)}
                  >
                    {tag.name}
                    <span className="ml-auto text-xs text-muted-foreground">
                      {tag.post_count}
                    </span>
                  </CommandItem>
                ))}
                {showCreate && (
                  <CommandItem
                    value={`create-${search}`}
                    onSelect={() => createAndSelect(search)}
                    disabled={creating}
                  >
                    <Plus className="mr-2 h-4 w-4" />
                    Create "{search}"
                  </CommandItem>
                )}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}
