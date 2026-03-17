import { defaultLocale, localeMap } from './locales';

const STORAGE_KEY = 'seva-locale';

// Use an object wrapper so we can export it without triggering
// Svelte 5's "cannot export reassigned $state" restriction.
const i18nState = $state({
	locale: defaultLocale,
	translations: {} as Record<string, any>,
	fallback: {} as Record<string, any>,
	loading: false
});

// Readonly derived getters for external use
export function getCurrentLocale(): string {
	return i18nState.locale;
}

// Direct property access for reactive reads in templates
export const i18n = i18nState;

async function loadTranslations(locale: string): Promise<void> {
	i18nState.loading = true;

	try {
		const module = await import(`./translations/${locale}.json`);
		i18nState.translations = module.default;
		i18nState.locale = locale;
	} catch (error) {
		console.warn(`Failed to load translations for "${locale}", falling back to "${defaultLocale}".`);

		if (locale !== defaultLocale) {
			try {
				const fallbackModule = await import(`./translations/${defaultLocale}.json`);
				i18nState.translations = fallbackModule.default;
				i18nState.locale = defaultLocale;
			} catch (fallbackError) {
				console.error('Failed to load fallback translations:', fallbackError);
			}
		}
	} finally {
		i18nState.loading = false;
	}
}

function getNestedValue(obj: Record<string, any>, path: string): string | undefined {
	const keys = path.split('.');
	let current: any = obj;

	for (const key of keys) {
		if (current === undefined || current === null || typeof current !== 'object') {
			return undefined;
		}
		current = current[key];
	}

	return typeof current === 'string' ? current : undefined;
}

function t(key: string, params?: Record<string, string | number>): string {
	let value = getNestedValue(i18nState.translations, key);

	if (value === undefined) {
		value = getNestedValue(i18nState.fallback, key);
	}

	if (value === undefined) {
		return key;
	}

	if (params) {
		for (const [paramKey, paramValue] of Object.entries(params)) {
			value = value.replaceAll(`{${paramKey}}`, String(paramValue));
		}
	}

	return value;
}

async function initLocale(): Promise<void> {
	// Load English as fallback first
	try {
		const fallbackModule = await import(`./translations/${defaultLocale}.json`);
		i18nState.fallback = fallbackModule.default;
	} catch (error) {
		console.error('Failed to load fallback (English) translations:', error);
	}

	let locale = defaultLocale;

	// Check localStorage for saved preference
	if (typeof window !== 'undefined') {
		const saved = localStorage.getItem(STORAGE_KEY);

		if (saved && localeMap[saved]) {
			locale = saved;
		} else if (typeof navigator !== 'undefined' && navigator.language) {
			const browserLang = navigator.language.split('-')[0];

			if (localeMap[browserLang]) {
				locale = browserLang;
			}
		}
	}

	await loadTranslations(locale);
}

async function setLocale(locale: string): Promise<void> {
	if (typeof window !== 'undefined') {
		localStorage.setItem(STORAGE_KEY, locale);
	}

	await loadTranslations(locale);
}

// Export currentLocale as a getter property for backward compatibility
// Components can read `currentLocale` reactively since it reads from $state
export { t, initLocale, setLocale, loadTranslations };

// For components that imported `currentLocale` directly, provide a derived-like export
// They should use `i18n.locale` instead, but we re-export for convenience
export const currentLocale = {
	get value() {
		return i18nState.locale;
	}
};
