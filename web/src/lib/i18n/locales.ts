export interface Locale {
	code: string;
	name: string;
	nativeName: string;
	direction: 'ltr' | 'rtl';
	speechCode: string;
	region: 'indian' | 'european' | 'global';
}

export const locales: Locale[] = [
	// Global
	{
		code: 'en',
		name: 'English',
		nativeName: 'English',
		direction: 'ltr',
		speechCode: 'en-US',
		region: 'global'
	},

	// Indian
	{
		code: 'hi',
		name: 'Hindi',
		nativeName: '\u0939\u093F\u0928\u094D\u0926\u0940',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'ta',
		name: 'Tamil',
		nativeName: '\u0BA4\u0BAE\u0BBF\u0BB4\u0BCD',
		direction: 'ltr',
		speechCode: 'ta-IN',
		region: 'indian'
	},
	{
		code: 'te',
		name: 'Telugu',
		nativeName: '\u0C24\u0C46\u0C32\u0C41\u0C17\u0C41',
		direction: 'ltr',
		speechCode: 'te-IN',
		region: 'indian'
	},
	{
		code: 'kn',
		name: 'Kannada',
		nativeName: '\u0C95\u0CA8\u0CCD\u0CA8\u0CA1',
		direction: 'ltr',
		speechCode: 'kn-IN',
		region: 'indian'
	},
	{
		code: 'ml',
		name: 'Malayalam',
		nativeName: '\u0D2E\u0D32\u0D2F\u0D3E\u0D33\u0D02',
		direction: 'ltr',
		speechCode: 'ml-IN',
		region: 'indian'
	},

	// European
	{
		code: 'de',
		name: 'German',
		nativeName: 'Deutsch',
		direction: 'ltr',
		speechCode: 'de-DE',
		region: 'european'
	},
	{
		code: 'fr',
		name: 'French',
		nativeName: 'Fran\u00E7ais',
		direction: 'ltr',
		speechCode: 'fr-FR',
		region: 'european'
	},
	{
		code: 'es',
		name: 'Spanish',
		nativeName: 'Espa\u00F1ol',
		direction: 'ltr',
		speechCode: 'es-ES',
		region: 'european'
	},
	{
		code: 'it',
		name: 'Italian',
		nativeName: 'Italiano',
		direction: 'ltr',
		speechCode: 'it-IT',
		region: 'european'
	},
	{
		code: 'pt',
		name: 'Portuguese',
		nativeName: 'Portugu\u00EAs',
		direction: 'ltr',
		speechCode: 'pt-PT',
		region: 'european'
	},
	{
		code: 'nl',
		name: 'Dutch',
		nativeName: 'Nederlands',
		direction: 'ltr',
		speechCode: 'nl-NL',
		region: 'european'
	},
	{
		code: 'pl',
		name: 'Polish',
		nativeName: 'Polski',
		direction: 'ltr',
		speechCode: 'pl-PL',
		region: 'european'
	},
	{
		code: 'ro',
		name: 'Romanian',
		nativeName: 'Rom\u00E2n\u0103',
		direction: 'ltr',
		speechCode: 'ro-RO',
		region: 'european'
	},
	{
		code: 'sv',
		name: 'Swedish',
		nativeName: 'Svenska',
		direction: 'ltr',
		speechCode: 'sv-SE',
		region: 'european'
	},
	{
		code: 'da',
		name: 'Danish',
		nativeName: 'Dansk',
		direction: 'ltr',
		speechCode: 'da-DK',
		region: 'european'
	},
	{
		code: 'fi',
		name: 'Finnish',
		nativeName: 'Suomi',
		direction: 'ltr',
		speechCode: 'fi-FI',
		region: 'european'
	},
	{
		code: 'no',
		name: 'Norwegian',
		nativeName: 'Norsk',
		direction: 'ltr',
		speechCode: 'nb-NO',
		region: 'european'
	},
	{
		code: 'el',
		name: 'Greek',
		nativeName: '\u0395\u03BB\u03BB\u03B7\u03BD\u03B9\u03BA\u03AC',
		direction: 'ltr',
		speechCode: 'el-GR',
		region: 'european'
	},
	{
		code: 'cs',
		name: 'Czech',
		nativeName: '\u010Ce\u0161tina',
		direction: 'ltr',
		speechCode: 'cs-CZ',
		region: 'european'
	},
	{
		code: 'hu',
		name: 'Hungarian',
		nativeName: 'Magyar',
		direction: 'ltr',
		speechCode: 'hu-HU',
		region: 'european'
	},
	{
		code: 'bg',
		name: 'Bulgarian',
		nativeName: '\u0411\u044A\u043B\u0433\u0430\u0440\u0441\u043A\u0438',
		direction: 'ltr',
		speechCode: 'bg-BG',
		region: 'european'
	},
	{
		code: 'hr',
		name: 'Croatian',
		nativeName: 'Hrvatski',
		direction: 'ltr',
		speechCode: 'hr-HR',
		region: 'european'
	},
	{
		code: 'sk',
		name: 'Slovak',
		nativeName: 'Sloven\u010Dina',
		direction: 'ltr',
		speechCode: 'sk-SK',
		region: 'european'
	},
	{
		code: 'sl',
		name: 'Slovenian',
		nativeName: 'Sloven\u0161\u010Dina',
		direction: 'ltr',
		speechCode: 'sl-SI',
		region: 'european'
	},
	{
		code: 'lt',
		name: 'Lithuanian',
		nativeName: 'Lietuvi\u0173',
		direction: 'ltr',
		speechCode: 'lt-LT',
		region: 'european'
	},
	{
		code: 'lv',
		name: 'Latvian',
		nativeName: 'Latvie\u0161u',
		direction: 'ltr',
		speechCode: 'lv-LV',
		region: 'european'
	},
	{
		code: 'et',
		name: 'Estonian',
		nativeName: 'Eesti',
		direction: 'ltr',
		speechCode: 'et-EE',
		region: 'european'
	},
	{
		code: 'uk',
		name: 'Ukrainian',
		nativeName: '\u0423\u043A\u0440\u0430\u0457\u043D\u0441\u044C\u043A\u0430',
		direction: 'ltr',
		speechCode: 'uk-UA',
		region: 'european'
	},
	{
		code: 'ru',
		name: 'Russian',
		nativeName: '\u0420\u0443\u0441\u0441\u043A\u0438\u0439',
		direction: 'ltr',
		speechCode: 'ru-RU',
		region: 'european'
	},
	{
		code: 'tr',
		name: 'Turkish',
		nativeName: 'T\u00FCrk\u00E7e',
		direction: 'ltr',
		speechCode: 'tr-TR',
		region: 'european'
	}
];

export const localeMap: Record<string, Locale> = Object.fromEntries(
	locales.map((locale) => [locale.code, locale])
);

export const defaultLocale = 'en';

export function getLocalesByRegion(): Record<string, Locale[]> {
	const grouped: Record<string, Locale[]> = {};

	for (const locale of locales) {
		if (!grouped[locale.region]) {
			grouped[locale.region] = [];
		}
		grouped[locale.region].push(locale);
	}

	return grouped;
}
