import { create } from "zustand";

export interface AuthUser {
	id: string;
	email: string;
	displayName: string;
	avatarUrl: string;
}

interface AuthState {
	user: AuthUser | null;
	accessToken: string | null;
	isAuthenticated: boolean;
	setAuth: (user: AuthUser, token: string) => void;
	clearAuth: () => void;
	setUser: (user: AuthUser) => void;
}

export const useAuthStore = create<AuthState>()((set) => ({
	user: null,
	accessToken: null,
	isAuthenticated: false,
	setAuth: (user, token) =>
		set({ user, accessToken: token, isAuthenticated: true }),
	clearAuth: () =>
		set({ user: null, accessToken: null, isAuthenticated: false }),
	setUser: (user) => set({ user }),
}));
