import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DndContext,
  closestCenter,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  ChevronRight,
  ChevronDown,
  Pencil,
  Plus,
  Trash2,
  GripVertical,
  FolderTree,
} from 'lucide-react';
import type { CategoryNode, ReorderItem } from '@/lib/content-api';

interface CategoryTreeProps {
  categories: CategoryNode[];
  onEdit: (category: CategoryNode) => void;
  onAddChild: (parentId: string) => void;
  onDelete: (category: CategoryNode) => void;
  onReorder: (orders: ReorderItem[]) => void;
}

export function CategoryTree({
  categories,
  onEdit,
  onAddChild,
  onDelete,
  onReorder,
}: CategoryTreeProps) {
  const { t } = useTranslation();
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const toggleExpand = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      // Find sibling list that contains both items
      const findSiblings = (nodes: CategoryNode[]): CategoryNode[] | null => {
        const ids = nodes.map((n) => n.id);
        if (ids.includes(String(active.id)) && ids.includes(String(over.id))) {
          return nodes;
        }
        for (const node of nodes) {
          const result = findSiblings(node.children);
          if (result) return result;
        }
        return null;
      };

      const siblings = findSiblings(categories);
      if (!siblings) return;

      const oldIndex = siblings.findIndex((n) => n.id === active.id);
      const newIndex = siblings.findIndex((n) => n.id === over.id);
      if (oldIndex === -1 || newIndex === -1) return;

      // Compute new sort_order values
      const reordered = [...siblings];
      const [moved] = reordered.splice(oldIndex, 1);
      reordered.splice(newIndex, 0, moved);

      const orders: ReorderItem[] = reordered.map((node, i) => ({
        id: node.id,
        sort_order: i + 1,
      }));
      onReorder(orders);
    },
    [categories, onReorder],
  );

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor),
  );

  if (categories.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <FolderTree className="h-12 w-12 mb-4 opacity-50" />
        <p>{t('content.noCategoriesFound')}</p>
      </div>
    );
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <TreeLevel
        nodes={categories}
        depth={0}
        expandedIds={expandedIds}
        onToggleExpand={toggleExpand}
        onEdit={onEdit}
        onAddChild={onAddChild}
        onDelete={onDelete}
      />
    </DndContext>
  );
}

interface TreeLevelProps {
  nodes: CategoryNode[];
  depth: number;
  expandedIds: Set<string>;
  onToggleExpand: (id: string) => void;
  onEdit: (category: CategoryNode) => void;
  onAddChild: (parentId: string) => void;
  onDelete: (category: CategoryNode) => void;
}

function TreeLevel({
  nodes,
  depth,
  expandedIds,
  onToggleExpand,
  onEdit,
  onAddChild,
  onDelete,
}: TreeLevelProps) {
  return (
    <SortableContext
      items={nodes.map((n) => n.id)}
      strategy={verticalListSortingStrategy}
    >
      <div className="space-y-1">
        {nodes.map((node) => (
          <TreeNode
            key={node.id}
            node={node}
            depth={depth}
            expandedIds={expandedIds}
            onToggleExpand={onToggleExpand}
            onEdit={onEdit}
            onAddChild={onAddChild}
            onDelete={onDelete}
          />
        ))}
      </div>
    </SortableContext>
  );
}

interface TreeNodeProps {
  node: CategoryNode;
  depth: number;
  expandedIds: Set<string>;
  onToggleExpand: (id: string) => void;
  onEdit: (category: CategoryNode) => void;
  onAddChild: (parentId: string) => void;
  onDelete: (category: CategoryNode) => void;
}

function TreeNode({
  node,
  depth,
  expandedIds,
  onToggleExpand,
  onEdit,
  onAddChild,
  onDelete,
}: TreeNodeProps) {
  const { t } = useTranslation();
  const hasChildren = node.children.length > 0;
  const isExpanded = expandedIds.has(node.id);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
  } = useSortable({ id: node.id });

  const style = {
    transform: CSS?.Transform?.toString(transform) ?? undefined,
    transition: transition ?? undefined,
  };

  return (
    <div ref={setNodeRef} style={style}>
      <div
        className="flex items-center gap-2 rounded-md border bg-card px-3 py-2 hover:bg-accent/50 transition-colors"
        style={{ marginLeft: `${depth * 24}px` }}
      >
        <button
          className="cursor-grab text-muted-foreground hover:text-foreground"
          aria-label="Drag to reorder"
          {...attributes}
          {...listeners}
        >
          <GripVertical className="h-4 w-4" />
        </button>

        {hasChildren ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={() => onToggleExpand(node.id)}
            aria-label="Toggle children"
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </Button>
        ) : (
          <span className="w-6" />
        )}

        <span className="font-medium flex-1">{node.name}</span>

        <Badge variant="secondary" className="text-xs">
          {t('content.postCount', { count: node.post_count })}
        </Badge>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0"
            onClick={() => onEdit(node)}
            aria-label="Edit Category"
          >
            <Pencil className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0"
            onClick={() => onAddChild(node.id)}
            aria-label="Add Subcategory"
          >
            <Plus className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0 text-destructive hover:text-destructive"
            onClick={() => onDelete(node)}
            aria-label="Delete Category"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {hasChildren && isExpanded && (
        <TreeLevel
          nodes={node.children}
          depth={depth + 1}
          expandedIds={expandedIds}
          onToggleExpand={onToggleExpand}
          onEdit={onEdit}
          onAddChild={onAddChild}
          onDelete={onDelete}
        />
      )}
    </div>
  );
}
