import { beforeEach, describe, expect, it } from "vitest";
import { useEditorStore } from "../editor-store";

describe("editor-store", () => {
	beforeEach(() => {
		useEditorStore.setState(useEditorStore.getInitialState());
	});

	describe("initial state", () => {
		it("has empty drafts", () => {
			expect(useEditorStore.getState().drafts).toEqual({});
		});
	});

	describe("saveDraft", () => {
		it("saves a draft for a given postId", () => {
			const content = JSON.stringify({ blocks: [{ type: "paragraph" }] });
			useEditorStore.getState().saveDraft("post-1", content);
			expect(useEditorStore.getState().drafts["post-1"]).toBe(content);
		});

		it("overwrites existing draft for the same postId", () => {
			useEditorStore.getState().saveDraft("post-1", "draft-v1");
			useEditorStore.getState().saveDraft("post-1", "draft-v2");
			expect(useEditorStore.getState().drafts["post-1"]).toBe("draft-v2");
		});

		it("saves multiple drafts for different postIds", () => {
			useEditorStore.getState().saveDraft("post-1", "content-1");
			useEditorStore.getState().saveDraft("post-2", "content-2");
			const drafts = useEditorStore.getState().drafts;
			expect(drafts["post-1"]).toBe("content-1");
			expect(drafts["post-2"]).toBe("content-2");
		});
	});

	describe("getDraft", () => {
		it("returns the draft for an existing postId", () => {
			useEditorStore.getState().saveDraft("post-1", "saved-content");
			expect(useEditorStore.getState().getDraft("post-1")).toBe(
				"saved-content",
			);
		});

		it("returns undefined for a non-existent postId", () => {
			expect(useEditorStore.getState().getDraft("no-such-post")).toBeUndefined();
		});
	});

	describe("clearDraft", () => {
		it("removes a specific draft", () => {
			useEditorStore.getState().saveDraft("post-1", "content-1");
			useEditorStore.getState().saveDraft("post-2", "content-2");
			useEditorStore.getState().clearDraft("post-1");
			expect(useEditorStore.getState().drafts["post-1"]).toBeUndefined();
			expect(useEditorStore.getState().drafts["post-2"]).toBe("content-2");
		});

		it("does nothing if postId does not exist", () => {
			useEditorStore.getState().saveDraft("post-1", "content-1");
			useEditorStore.getState().clearDraft("no-such-post");
			expect(useEditorStore.getState().drafts["post-1"]).toBe("content-1");
		});
	});

	describe("clearAllDrafts", () => {
		it("removes all drafts", () => {
			useEditorStore.getState().saveDraft("post-1", "content-1");
			useEditorStore.getState().saveDraft("post-2", "content-2");
			useEditorStore.getState().saveDraft("post-3", "content-3");
			useEditorStore.getState().clearAllDrafts();
			expect(useEditorStore.getState().drafts).toEqual({});
		});
	});
});
