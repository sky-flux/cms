import { create } from "zustand";

interface EditorState {
	drafts: Record<string, string>;
	saveDraft: (postId: string, content: string) => void;
	getDraft: (postId: string) => string | undefined;
	clearDraft: (postId: string) => void;
	clearAllDrafts: () => void;
}

export const useEditorStore = create<EditorState>()((set, get) => ({
	drafts: {},
	saveDraft: (postId, content) =>
		set((state) => ({ drafts: { ...state.drafts, [postId]: content } })),
	getDraft: (postId) => get().drafts[postId],
	clearDraft: (postId) =>
		set((state) => {
			const { [postId]: _, ...rest } = state.drafts;
			return { drafts: rest };
		}),
	clearAllDrafts: () => set({ drafts: {} }),
}));
