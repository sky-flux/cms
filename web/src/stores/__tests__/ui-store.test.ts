import { beforeEach, describe, expect, it } from "vitest";
import { useUIStore } from "../ui-store";

describe("ui-store", () => {
	beforeEach(() => {
		useUIStore.setState(useUIStore.getInitialState());
	});

	describe("initial state", () => {
		it("has system theme by default", () => {
			expect(useUIStore.getState().theme).toBe("system");
		});

		it("has zh-CN locale by default", () => {
			expect(useUIStore.getState().locale).toBe("zh-CN");
		});

		it("has empty siteSlug", () => {
			expect(useUIStore.getState().siteSlug).toBe("");
		});

		it("has sidebar open by default", () => {
			expect(useUIStore.getState().sidebarOpen).toBe(true);
		});
	});

	describe("setTheme", () => {
		it("updates theme to light", () => {
			useUIStore.getState().setTheme("light");
			expect(useUIStore.getState().theme).toBe("light");
		});

		it("updates theme to dark", () => {
			useUIStore.getState().setTheme("dark");
			expect(useUIStore.getState().theme).toBe("dark");
		});
	});

	describe("setLocale", () => {
		it("updates locale to en", () => {
			useUIStore.getState().setLocale("en");
			expect(useUIStore.getState().locale).toBe("en");
		});

		it("updates locale back to zh-CN", () => {
			useUIStore.getState().setLocale("en");
			useUIStore.getState().setLocale("zh-CN");
			expect(useUIStore.getState().locale).toBe("zh-CN");
		});
	});

	describe("setSiteSlug", () => {
		it("updates siteSlug", () => {
			useUIStore.getState().setSiteSlug("my_blog");
			expect(useUIStore.getState().siteSlug).toBe("my_blog");
		});
	});

	describe("toggleSidebar", () => {
		it("toggles sidebar from open to closed", () => {
			useUIStore.getState().toggleSidebar();
			expect(useUIStore.getState().sidebarOpen).toBe(false);
		});

		it("toggles sidebar from closed to open", () => {
			useUIStore.getState().toggleSidebar();
			useUIStore.getState().toggleSidebar();
			expect(useUIStore.getState().sidebarOpen).toBe(true);
		});
	});

	describe("setSidebarOpen", () => {
		it("sets sidebar to a specific state", () => {
			useUIStore.getState().setSidebarOpen(false);
			expect(useUIStore.getState().sidebarOpen).toBe(false);
			useUIStore.getState().setSidebarOpen(true);
			expect(useUIStore.getState().sidebarOpen).toBe(true);
		});
	});
});
