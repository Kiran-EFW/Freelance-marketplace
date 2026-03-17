import { defaultLocale, localeMap } from './locales';

const STORAGE_KEY = 'seva-locale';

let currentLocale = $state(defaultLocale);
let translations = $state<Record<string, any>>({});
let loading = $state(false);
let fallbackTranslations = $state<Record<string, any>>({});

async function loadTranslations(locale: string): Promise<void> {
	loading = true;

	try {
		const module = await import(`./translations/${locale}.json`);
		translations = module.default;
		currentLocale = locale;
	} catch (error) {
		console.warn(`Failed to load translations for "${locale}", falling back to "${defaultLocale}".`);

		if (locale !== defaultLocale) {
			try {
				const fallbackModule = await import(`./translations/${defaultLocale}.json`);
				translations = fallbackModule.default;
				currentLocale = defaultLocale;
			} catch (fallbackError) {
				console.error('Failed to load fallback translations:', fallbackError);
			}
		}
	} finally {
		loading = false;
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
	let value = getNestedValue(translations, key);

	if (value === undefined) {
		value = getNestedValue(fallbackTranslations, key);
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
		fallbackTranslations = fallbackModule.default;
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

export {
	currentLocale,
	translations,
	loading,
	t,
	initLocale,
	setLocale,
	loadTranslations
};
