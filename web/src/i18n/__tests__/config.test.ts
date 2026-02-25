import { describe, expect, it } from "vitest";
import i18n from "../config";

describe("i18n config", () => {
	it("initializes with zh-CN as default language", () => {
		expect(i18n.language).toBe("zh-CN");
	});

	it("has en as fallback language", () => {
		expect(i18n.options.fallbackLng).toEqual(["en"]);
	});

	it("translates zh-CN keys", () => {
		expect(i18n.t("common.save")).toBe("保存");
		expect(i18n.t("nav.dashboard")).toBe("仪表盘");
		expect(i18n.t("auth.login")).toBe("登录");
	});

	it("translates en keys when language is switched", async () => {
		await i18n.changeLanguage("en");
		expect(i18n.t("common.save")).toBe("Save");
		expect(i18n.t("nav.dashboard")).toBe("Dashboard");
		expect(i18n.t("auth.login")).toBe("Login");
		await i18n.changeLanguage("zh-CN");
	});

	it("has all required translation namespaces", () => {
		const enKeys = Object.keys(
			i18n.getResourceBundle("en", "translation"),
		);
		expect(enKeys).toContain("common");
		expect(enKeys).toContain("nav");
		expect(enKeys).toContain("auth");
		expect(enKeys).toContain("dashboard");
		expect(enKeys).toContain("posts");
		expect(enKeys).toContain("media");
		expect(enKeys).toContain("settings");
		expect(enKeys).toContain("errors");
		expect(enKeys).toContain("messages");
	});

	it("has matching keys between en and zh-CN", () => {
		const en = i18n.getResourceBundle("en", "translation");
		const zhCN = i18n.getResourceBundle("zh-CN", "translation");
		const enTopKeys = Object.keys(en).sort();
		const zhCNTopKeys = Object.keys(zhCN).sort();
		expect(enTopKeys).toEqual(zhCNTopKeys);

		for (const key of enTopKeys) {
			const enSubKeys = Object.keys(en[key]).sort();
			const zhCNSubKeys = Object.keys(zhCN[key]).sort();
			expect(enSubKeys).toEqual(zhCNSubKeys);
		}
	});
});
