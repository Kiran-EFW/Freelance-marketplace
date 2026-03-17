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
	{
		code: 'as',
		name: 'Assamese',
		nativeName: '\u0985\u09B8\u09AE\u09C0\u09AF\u09BC\u09BE',
		direction: 'ltr',
		speechCode: 'as-IN',
		region: 'indian'
	},
	{
		code: 'bn',
		name: 'Bengali',
		nativeName: '\u09AC\u09BE\u0982\u09B2\u09BE',
		direction: 'ltr',
		speechCode: 'bn-IN',
		region: 'indian'
	},
	{
		code: 'brx',
		name: 'Bodo',
		nativeName: '\u092C\u0921\u093C\u094B',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'doi',
		name: 'Dogri',
		nativeName: '\u0921\u094B\u0917\u0930\u0940',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'gu',
		name: 'Gujarati',
		nativeName: '\u0A97\u0AC1\u0A9C\u0AB0\u0ABE\u0AA4\u0AC0',
		direction: 'ltr',
		speechCode: 'gu-IN',
		region: 'indian'
	},
	{
		code: 'ks',
		name: 'Kashmiri',
		nativeName: '\u0915\u0949\u0936\u0941\u0930',
		direction: 'ltr',
		speechCode: 'ur-IN',
		region: 'indian'
	},
	{
		code: 'kok',
		name: 'Konkani',
		nativeName: '\u0915\u094B\u0902\u0915\u0923\u0940',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'mai',
		name: 'Maithili',
		nativeName: '\u092E\u0948\u0925\u093F\u0932\u0940',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'mni',
		name: 'Manipuri',
		nativeName: '\u09AE\u09C8\u09A4\u09C8\u09B2\u09CB\u09A8\u09CD',
		direction: 'ltr',
		speechCode: 'bn-IN',
		region: 'indian'
	},
	{
		code: 'mr',
		name: 'Marathi',
		nativeName: '\u092E\u0930\u093E\u0920\u0940',
		direction: 'ltr',
		speechCode: 'mr-IN',
		region: 'indian'
	},
	{
		code: 'ne',
		name: 'Nepali',
		nativeName: '\u0928\u0947\u092A\u093E\u0932\u0940',
		direction: 'ltr',
		speechCode: 'ne-NP',
		region: 'indian'
	},
	{
		code: 'or',
		name: 'Odia',
		nativeName: '\u0B13\u0B21\u0B3C\u0B3F\u0B06',
		direction: 'ltr',
		speechCode: 'or-IN',
		region: 'indian'
	},
	{
		code: 'pa',
		name: 'Punjabi',
		nativeName: '\u0A2A\u0A70\u0A1C\u0A3E\u0A2C\u0A40',
		direction: 'ltr',
		speechCode: 'pa-IN',
		region: 'indian'
	},
	{
		code: 'sa',
		name: 'Sanskrit',
		nativeName: '\u0938\u0902\u0938\u094D\u0915\u0943\u0924\u092E\u094D',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'sat',
		name: 'Santali',
		nativeName: '\u1CA5\u1C9F\u1CA8\u1C9B\u1C9F\u1CAC\u1C98',
		direction: 'ltr',
		speechCode: 'hi-IN',
		region: 'indian'
	},
	{
		code: 'sd',
		name: 'Sindhi',
		nativeName: '\u0633\u0646\u068C\u064A',
		direction: 'ltr',
		speechCode: 'sd-IN',
		region: 'indian'
	},
	{
		code: 'ur',
		name: 'Urdu',
		nativeName: '\u0627\u0631\u062F\u0648',
		direction: 'rtl',
		speechCode: 'ur-IN',
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
