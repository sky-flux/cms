import { useCallback, useMemo } from 'react';
import { Checkbox } from '@/components/ui/checkbox';

export interface TreeNode {
  id: string;
  label: string;
  children: TreeNode[];
}

interface PermissionTreeProps {
  items: TreeNode[];
  checkedIds: string[];
  onChange: (ids: string[]) => void;
}

function getAllDescendantIds(node: TreeNode): string[] {
  const ids: string[] = [node.id];
  for (const child of node.children) {
    ids.push(...getAllDescendantIds(child));
  }
  return ids;
}

function getCheckState(
  node: TreeNode,
  checkedSet: Set<string>,
): 'checked' | 'unchecked' | 'indeterminate' {
  if (node.children.length === 0) {
    return checkedSet.has(node.id) ? 'checked' : 'unchecked';
  }
  const childStates = node.children.map((c) => getCheckState(c, checkedSet));
  if (childStates.every((s) => s === 'checked') && checkedSet.has(node.id)) return 'checked';
  if (childStates.some((s) => s === 'checked' || s === 'indeterminate')) return 'indeterminate';
  return 'unchecked';
}

function TreeNodeRow({
  node,
  checkedSet,
  onToggle,
  depth = 0,
}: {
  node: TreeNode;
  checkedSet: Set<string>;
  onToggle: (node: TreeNode) => void;
  depth?: number;
}) {
  const state = getCheckState(node, checkedSet);

  return (
    <div>
      <label
        className="flex items-center gap-2 py-1 hover:bg-muted/50 rounded px-2 cursor-pointer"
        style={{ paddingLeft: `${depth * 24 + 8}px` }}
      >
        <Checkbox
          checked={state === 'checked' ? true : state === 'indeterminate' ? 'indeterminate' : false}
          onCheckedChange={() => onToggle(node)}
        />
        <span className="text-sm">{node.label}</span>
      </label>
      {node.children.map((child) => (
        <TreeNodeRow
          key={child.id}
          node={child}
          checkedSet={checkedSet}
          onToggle={onToggle}
          depth={depth + 1}
        />
      ))}
    </div>
  );
}

export function PermissionTree({ items, checkedIds, onChange }: PermissionTreeProps) {
  const checkedSet = useMemo(() => new Set(checkedIds), [checkedIds]);

  const handleToggle = useCallback(
    (node: TreeNode) => {
      const allIds = getAllDescendantIds(node);
      const currentState = getCheckState(node, checkedSet);
      const newSet = new Set(checkedIds);

      if (currentState === 'checked') {
        for (const id of allIds) newSet.delete(id);
      } else {
        for (const id of allIds) newSet.add(id);
      }

      onChange(Array.from(newSet));
    },
    [checkedIds, checkedSet, onChange],
  );

  if (items.length === 0) {
    return <p className="text-sm text-muted-foreground py-4">No permissions available</p>;
  }

  return (
    <div className="space-y-0.5">
      {items.map((item) => (
        <TreeNodeRow
          key={item.id}
          node={item}
          checkedSet={checkedSet}
          onToggle={handleToggle}
        />
      ))}
    </div>
  );
}
