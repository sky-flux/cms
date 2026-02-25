import { beforeEach, describe, expect, it } from "vitest";
import { useAuthStore } from "../auth-store";

describe("auth-store", () => {
	beforeEach(() => {
		useAuthStore.setState(useAuthStore.getInitialState());
	});

	describe("initial state", () => {
		it("has null user", () => {
			expect(useAuthStore.getState().user).toBeNull();
		});

		it("has null accessToken", () => {
			expect(useAuthStore.getState().accessToken).toBeNull();
		});

		it("is not authenticated", () => {
			expect(useAuthStore.getState().isAuthenticated).toBe(false);
		});
	});

	describe("setAuth", () => {
		const mockUser = {
			id: "test-uuid",
			email: "admin@example.com",
			displayName: "Admin",
			avatarUrl: "https://example.com/avatar.png",
		};

		it("sets user and token", () => {
			useAuthStore.getState().setAuth(mockUser, "jwt-token-123");
			const state = useAuthStore.getState();
			expect(state.user).toEqual(mockUser);
			expect(state.accessToken).toBe("jwt-token-123");
		});

		it("sets isAuthenticated to true", () => {
			useAuthStore.getState().setAuth(mockUser, "jwt-token-123");
			expect(useAuthStore.getState().isAuthenticated).toBe(true);
		});
	});

	describe("clearAuth", () => {
		it("clears user, token, and sets isAuthenticated to false", () => {
			const mockUser = {
				id: "test-uuid",
				email: "admin@example.com",
				displayName: "Admin",
				avatarUrl: "",
			};
			useAuthStore.getState().setAuth(mockUser, "jwt-token-123");
			useAuthStore.getState().clearAuth();
			const state = useAuthStore.getState();
			expect(state.user).toBeNull();
			expect(state.accessToken).toBeNull();
			expect(state.isAuthenticated).toBe(false);
		});
	});

	describe("setUser", () => {
		it("updates user without changing token or auth status", () => {
			const originalUser = {
				id: "test-uuid",
				email: "admin@example.com",
				displayName: "Admin",
				avatarUrl: "",
			};
			useAuthStore.getState().setAuth(originalUser, "jwt-token-123");

			const updatedUser = {
				id: "test-uuid",
				email: "admin@example.com",
				displayName: "Updated Admin",
				avatarUrl: "https://example.com/new-avatar.png",
			};
			useAuthStore.getState().setUser(updatedUser);

			const state = useAuthStore.getState();
			expect(state.user).toEqual(updatedUser);
			expect(state.accessToken).toBe("jwt-token-123");
			expect(state.isAuthenticated).toBe(true);
		});
	});
});
