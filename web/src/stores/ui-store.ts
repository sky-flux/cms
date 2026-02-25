import { create } from "zustand";

export type Theme = "light" | "dark" | "system";
export type Locale = "zh-CN" | "en";

interface UIState {
	theme: Theme;
	locale: Locale;
	siteSlug: string;
	sidebarOpen: boolean;
	setTheme: (theme: Theme) => void;
	setLocale: (locale: Locale) => void;
	setSiteSlug: (slug: string) => void;
	toggleSidebar: () => void;
	setSidebarOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>()((set) => ({
	theme: "system",
	locale: "zh-CN",
	siteSlug: "",
	sidebarOpen: true,
	setTheme: (theme) => set({ theme }),
	setLocale: (locale) => set({ locale }),
	setSiteSlug: (siteSlug) => set({ siteSlug }),
	toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
	setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
}));
