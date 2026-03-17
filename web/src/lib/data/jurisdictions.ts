// ---------------------------------------------------------------------------
// Jurisdiction / Country configuration
// ---------------------------------------------------------------------------

export interface PaymentMethod {
	id: string; // 'upi', 'card', 'bank_transfer', 'cash', 'wallet', etc.
	name: string;
	icon: string; // lucide icon name
	isDefault: boolean;
	description: string;
}

export interface Jurisdiction {
	code: string; // ISO 3166-1 alpha-2 (IN, GB, US, DE, etc.)
	name: string; // English name
	nativeName: string; // Name in local language
	phoneCode: string; // e.g., '+91', '+44'
	phonePlaceholder: string; // e.g., '9876543210', '7911 123456'
	phoneMaxLength: number;
	currency: string; // ISO 4217 (INR, GBP, EUR, USD, etc.)
	currencySymbol: string; // Rs., £, €, $
	languages: string[]; // supported i18n locale codes
	flag: string; // emoji flag
	postcodeFormat: string; // regex pattern or description
	postcodePlaceholder: string;
	paymentMethods: PaymentMethod[];
	taxLabel: string; // GST, VAT, Tax, etc.
	taxRate: number; // default tax rate (decimal, e.g. 0.18 for 18%)
}

// ---------------------------------------------------------------------------
// Jurisdiction data
// ---------------------------------------------------------------------------

export const jurisdictions: Jurisdiction[] = [
	// -----------------------------------------------------------------------
	// India
	// -----------------------------------------------------------------------
	{
		code: 'IN',
		name: 'India',
		nativeName: '\u092D\u093E\u0930\u0924',
		phoneCode: '+91',
		phonePlaceholder: '9876543210',
		phoneMaxLength: 10,
		currency: 'INR',
		currencySymbol: '\u20B9',
		languages: ['en', 'hi', 'ta', 'te', 'kn', 'ml', 'as', 'bn', 'brx', 'doi', 'gu', 'ks', 'kok', 'mai', 'mni', 'mr', 'ne', 'or', 'pa', 'sa', 'sat', 'sd', 'ur'],
		flag: '\uD83C\uDDEE\uD83C\uDDF3',
		postcodeFormat: '\\d{6}',
		postcodePlaceholder: '560001',
		paymentMethods: [
			{
				id: 'upi',
				name: 'UPI',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Pay using any UPI app (GPay, PhonePe, Paytm)'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard, RuPay'
			},
			{
				id: 'cash',
				name: 'Cash',
				icon: 'Banknote',
				isDefault: false,
				description: 'Pay in cash after service completion'
			},
			{
				id: 'wallet',
				name: 'Paytm / Wallet',
				icon: 'Wallet',
				isDefault: false,
				description: 'Pay using Paytm or other digital wallets'
			},
			{
				id: 'net_banking',
				name: 'Net Banking',
				icon: 'Building2',
				isDefault: false,
				description: 'Pay via internet banking'
			}
		],
		taxLabel: 'GST',
		taxRate: 0.18
	},

	// -----------------------------------------------------------------------
	// Nepal
	// -----------------------------------------------------------------------
	{
		code: 'NP',
		name: 'Nepal',
		nativeName: '\u0928\u0947\u092A\u093E\u0932',
		phoneCode: '+977',
		phonePlaceholder: '9812345678',
		phoneMaxLength: 10,
		currency: 'NPR',
		currencySymbol: 'Rs.',
		languages: ['ne', 'en'],
		flag: '\uD83C\uDDF3\uD83C\uDDF5',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '44600',
		paymentMethods: [
			{
				id: 'esewa',
				name: 'eSewa',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Pay using eSewa digital wallet'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard'
			}
		],
		taxLabel: 'VAT',
		taxRate: 0.13
	},

	// -----------------------------------------------------------------------
	// United Kingdom
	// -----------------------------------------------------------------------
	{
		code: 'GB',
		name: 'United Kingdom',
		nativeName: 'United Kingdom',
		phoneCode: '+44',
		phonePlaceholder: '7911 123456',
		phoneMaxLength: 11,
		currency: 'GBP',
		currencySymbol: '\u00A3',
		languages: ['en'],
		flag: '\uD83C\uDDEC\uD83C\uDDE7',
		postcodeFormat: '[A-Z]{1,2}\\d[A-Z\\d]?\\s?\\d[A-Z]{2}',
		postcodePlaceholder: 'SW1A 1AA',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Amex'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Direct bank transfer via Faster Payments'
			},
			{
				id: 'apple_pay',
				name: 'Apple Pay',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pay using Apple Pay'
			}
		],
		taxLabel: 'VAT',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// United States
	// -----------------------------------------------------------------------
	{
		code: 'US',
		name: 'United States',
		nativeName: 'United States',
		phoneCode: '+1',
		phonePlaceholder: '(555) 123-4567',
		phoneMaxLength: 10,
		currency: 'USD',
		currencySymbol: '$',
		languages: ['en', 'es'],
		flag: '\uD83C\uDDFA\uD83C\uDDF8',
		postcodeFormat: '\\d{5}(-\\d{4})?',
		postcodePlaceholder: '10001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Amex, Discover'
			},
			{
				id: 'apple_pay',
				name: 'Apple Pay',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pay using Apple Pay'
			},
			{
				id: 'venmo',
				name: 'Venmo',
				icon: 'Wallet',
				isDefault: false,
				description: 'Pay using Venmo'
			},
			{
				id: 'cash',
				name: 'Cash',
				icon: 'Banknote',
				isDefault: false,
				description: 'Pay in cash after service completion'
			}
		],
		taxLabel: 'Sales Tax',
		taxRate: 0.0
	},

	// -----------------------------------------------------------------------
	// Germany
	// -----------------------------------------------------------------------
	{
		code: 'DE',
		name: 'Germany',
		nativeName: 'Deutschland',
		phoneCode: '+49',
		phonePlaceholder: '151 12345678',
		phoneMaxLength: 12,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['de', 'en'],
		flag: '\uD83C\uDDE9\uD83C\uDDEA',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '10115',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Girocard'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'SEPA direct debit or bank transfer'
			},
			{
				id: 'paypal',
				name: 'PayPal',
				icon: 'Wallet',
				isDefault: false,
				description: 'Pay using PayPal'
			}
		],
		taxLabel: 'MwSt',
		taxRate: 0.19
	},

	// -----------------------------------------------------------------------
	// France
	// -----------------------------------------------------------------------
	{
		code: 'FR',
		name: 'France',
		nativeName: 'France',
		phoneCode: '+33',
		phonePlaceholder: '6 12 34 56 78',
		phoneMaxLength: 10,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['fr', 'en'],
		flag: '\uD83C\uDDEB\uD83C\uDDF7',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '75001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Carte Bancaire, Visa, Mastercard'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Virement SEPA'
			},
			{
				id: 'paypal',
				name: 'PayPal',
				icon: 'Wallet',
				isDefault: false,
				description: 'Payer avec PayPal'
			}
		],
		taxLabel: 'TVA',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// Spain
	// -----------------------------------------------------------------------
	{
		code: 'ES',
		name: 'Spain',
		nativeName: 'Espa\u00F1a',
		phoneCode: '+34',
		phonePlaceholder: '612 345 678',
		phoneMaxLength: 9,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['es', 'en'],
		flag: '\uD83C\uDDEA\uD83C\uDDF8',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '28001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bizum',
				name: 'Bizum',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pagar con Bizum'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Transferencia SEPA'
			}
		],
		taxLabel: 'IVA',
		taxRate: 0.21
	},

	// -----------------------------------------------------------------------
	// Italy
	// -----------------------------------------------------------------------
	{
		code: 'IT',
		name: 'Italy',
		nativeName: 'Italia',
		phoneCode: '+39',
		phonePlaceholder: '312 345 6789',
		phoneMaxLength: 10,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['it', 'en'],
		flag: '\uD83C\uDDEE\uD83C\uDDF9',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '00100',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, CartaSi'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Bonifico SEPA'
			},
			{
				id: 'postepay',
				name: 'PostePay',
				icon: 'Wallet',
				isDefault: false,
				description: 'Pagamento con PostePay'
			}
		],
		taxLabel: 'IVA',
		taxRate: 0.22
	},

	// -----------------------------------------------------------------------
	// Netherlands
	// -----------------------------------------------------------------------
	{
		code: 'NL',
		name: 'Netherlands',
		nativeName: 'Nederland',
		phoneCode: '+31',
		phonePlaceholder: '6 12345678',
		phoneMaxLength: 10,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['nl', 'en'],
		flag: '\uD83C\uDDF3\uD83C\uDDF1',
		postcodeFormat: '\\d{4}\\s?[A-Z]{2}',
		postcodePlaceholder: '1012 AB',
		paymentMethods: [
			{
				id: 'ideal',
				name: 'iDEAL',
				icon: 'Building2',
				isDefault: true,
				description: 'Betaal met iDEAL via uw bank'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'SEPA overboeking'
			}
		],
		taxLabel: 'BTW',
		taxRate: 0.21
	},

	// -----------------------------------------------------------------------
	// Portugal
	// -----------------------------------------------------------------------
	{
		code: 'PT',
		name: 'Portugal',
		nativeName: 'Portugal',
		phoneCode: '+351',
		phonePlaceholder: '912 345 678',
		phoneMaxLength: 9,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['pt', 'en'],
		flag: '\uD83C\uDDF5\uD83C\uDDF9',
		postcodeFormat: '\\d{4}-\\d{3}',
		postcodePlaceholder: '1000-001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Multibanco'
			},
			{
				id: 'mb_way',
				name: 'MB Way',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pagar com MB Way'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Transfer\u00EAncia SEPA'
			}
		],
		taxLabel: 'IVA',
		taxRate: 0.23
	},

	// -----------------------------------------------------------------------
	// Poland
	// -----------------------------------------------------------------------
	{
		code: 'PL',
		name: 'Poland',
		nativeName: 'Polska',
		phoneCode: '+48',
		phonePlaceholder: '512 345 678',
		phoneMaxLength: 9,
		currency: 'PLN',
		currencySymbol: 'z\u0142',
		languages: ['pl', 'en'],
		flag: '\uD83C\uDDF5\uD83C\uDDF1',
		postcodeFormat: '\\d{2}-\\d{3}',
		postcodePlaceholder: '00-001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'blik',
				name: 'BLIK',
				icon: 'Smartphone',
				isDefault: false,
				description: 'P\u0142a\u0107 za pomoc\u0105 BLIK'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Przelew bankowy'
			}
		],
		taxLabel: 'VAT',
		taxRate: 0.23
	},

	// -----------------------------------------------------------------------
	// Sweden
	// -----------------------------------------------------------------------
	{
		code: 'SE',
		name: 'Sweden',
		nativeName: 'Sverige',
		phoneCode: '+46',
		phonePlaceholder: '70 123 45 67',
		phoneMaxLength: 10,
		currency: 'SEK',
		currencySymbol: 'kr',
		languages: ['sv', 'en'],
		flag: '\uD83C\uDDF8\uD83C\uDDEA',
		postcodeFormat: '\\d{3}\\s?\\d{2}',
		postcodePlaceholder: '111 22',
		paymentMethods: [
			{
				id: 'swish',
				name: 'Swish',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Betala med Swish'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Banko\u00F6verf\u00F6ring'
			}
		],
		taxLabel: 'Moms',
		taxRate: 0.25
	},

	// -----------------------------------------------------------------------
	// Norway
	// -----------------------------------------------------------------------
	{
		code: 'NO',
		name: 'Norway',
		nativeName: 'Norge',
		phoneCode: '+47',
		phonePlaceholder: '412 34 567',
		phoneMaxLength: 8,
		currency: 'NOK',
		currencySymbol: 'kr',
		languages: ['no', 'en'],
		flag: '\uD83C\uDDF3\uD83C\uDDF4',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '0150',
		paymentMethods: [
			{
				id: 'vipps',
				name: 'Vipps',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Betal med Vipps'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard, BankAxept'
			}
		],
		taxLabel: 'MVA',
		taxRate: 0.25
	},

	// -----------------------------------------------------------------------
	// Denmark
	// -----------------------------------------------------------------------
	{
		code: 'DK',
		name: 'Denmark',
		nativeName: 'Danmark',
		phoneCode: '+45',
		phonePlaceholder: '20 12 34 56',
		phoneMaxLength: 8,
		currency: 'DKK',
		currencySymbol: 'kr',
		languages: ['da', 'en'],
		flag: '\uD83C\uDDE9\uD83C\uDDF0',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '1000',
		paymentMethods: [
			{
				id: 'mobilepay',
				name: 'MobilePay',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Betal med MobilePay'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard, Dankort'
			}
		],
		taxLabel: 'Moms',
		taxRate: 0.25
	},

	// -----------------------------------------------------------------------
	// Finland
	// -----------------------------------------------------------------------
	{
		code: 'FI',
		name: 'Finland',
		nativeName: 'Suomi',
		phoneCode: '+358',
		phonePlaceholder: '41 2345678',
		phoneMaxLength: 10,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['fi', 'en'],
		flag: '\uD83C\uDDEB\uD83C\uDDEE',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '00100',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Pankkisiirto'
			}
		],
		taxLabel: 'ALV',
		taxRate: 0.24
	},

	// -----------------------------------------------------------------------
	// Austria
	// -----------------------------------------------------------------------
	{
		code: 'AT',
		name: 'Austria',
		nativeName: '\u00D6sterreich',
		phoneCode: '+43',
		phonePlaceholder: '664 1234567',
		phoneMaxLength: 11,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['de', 'en'],
		flag: '\uD83C\uDDE6\uD83C\uDDF9',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '1010',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'eps',
				name: 'EPS',
				icon: 'Building2',
				isDefault: false,
				description: 'EPS Online-\u00DCberweisung'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'SEPA-\u00DCberweisung'
			}
		],
		taxLabel: 'USt',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// Belgium
	// -----------------------------------------------------------------------
	{
		code: 'BE',
		name: 'Belgium',
		nativeName: 'Belgi\u00EB / Belgique',
		phoneCode: '+32',
		phonePlaceholder: '470 12 34 56',
		phoneMaxLength: 9,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['nl', 'fr', 'de', 'en'],
		flag: '\uD83C\uDDE7\uD83C\uDDEA',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '1000',
		paymentMethods: [
			{
				id: 'bancontact',
				name: 'Bancontact',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Betaal met Bancontact / Payer avec Bancontact'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard'
			},
			{
				id: 'sepa',
				name: 'SEPA Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'SEPA-overschrijving / Virement SEPA'
			}
		],
		taxLabel: 'BTW/TVA',
		taxRate: 0.21
	},

	// -----------------------------------------------------------------------
	// Ireland
	// -----------------------------------------------------------------------
	{
		code: 'IE',
		name: 'Ireland',
		nativeName: '\u00C9ire',
		phoneCode: '+353',
		phonePlaceholder: '85 123 4567',
		phoneMaxLength: 9,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['en'],
		flag: '\uD83C\uDDEE\uD83C\uDDEA',
		postcodeFormat: '[A-Z]\\d{2}\\s?[A-Z\\d]{4}',
		postcodePlaceholder: 'D02 AF30',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Direct bank transfer'
			}
		],
		taxLabel: 'VAT',
		taxRate: 0.23
	},

	// -----------------------------------------------------------------------
	// Switzerland
	// -----------------------------------------------------------------------
	{
		code: 'CH',
		name: 'Switzerland',
		nativeName: 'Schweiz / Suisse / Svizzera',
		phoneCode: '+41',
		phonePlaceholder: '78 123 45 67',
		phoneMaxLength: 10,
		currency: 'CHF',
		currencySymbol: 'CHF',
		languages: ['de', 'fr', 'it', 'en'],
		flag: '\uD83C\uDDE8\uD83C\uDDED',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '8001',
		paymentMethods: [
			{
				id: 'twint',
				name: 'TWINT',
				icon: 'Smartphone',
				isDefault: true,
				description: 'Bezahlen mit TWINT'
			},
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: false,
				description: 'Visa, Mastercard, PostFinance'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Bank\u00FCberweisung'
			}
		],
		taxLabel: 'MwSt',
		taxRate: 0.077
	},

	// -----------------------------------------------------------------------
	// Czech Republic
	// -----------------------------------------------------------------------
	{
		code: 'CZ',
		name: 'Czech Republic',
		nativeName: '\u010Cesko',
		phoneCode: '+420',
		phonePlaceholder: '601 123 456',
		phoneMaxLength: 9,
		currency: 'CZK',
		currencySymbol: 'K\u010D',
		languages: ['cs', 'en'],
		flag: '\uD83C\uDDE8\uD83C\uDDFF',
		postcodeFormat: '\\d{3}\\s?\\d{2}',
		postcodePlaceholder: '110 00',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Bankovn\u00ED p\u0159evod'
			}
		],
		taxLabel: 'DPH',
		taxRate: 0.21
	},

	// -----------------------------------------------------------------------
	// Romania
	// -----------------------------------------------------------------------
	{
		code: 'RO',
		name: 'Romania',
		nativeName: 'Rom\u00E2nia',
		phoneCode: '+40',
		phonePlaceholder: '712 345 678',
		phoneMaxLength: 9,
		currency: 'RON',
		currencySymbol: 'lei',
		languages: ['ro', 'en'],
		flag: '\uD83C\uDDF7\uD83C\uDDF4',
		postcodeFormat: '\\d{6}',
		postcodePlaceholder: '010001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Transfer bancar'
			}
		],
		taxLabel: 'TVA',
		taxRate: 0.19
	},

	// -----------------------------------------------------------------------
	// Hungary
	// -----------------------------------------------------------------------
	{
		code: 'HU',
		name: 'Hungary',
		nativeName: 'Magyarorsz\u00E1g',
		phoneCode: '+36',
		phonePlaceholder: '20 123 4567',
		phoneMaxLength: 9,
		currency: 'HUF',
		currencySymbol: 'Ft',
		languages: ['hu', 'en'],
		flag: '\uD83C\uDDED\uD83C\uDDFA',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '1011',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Banki \u00E1tutal\u00E1s'
			}
		],
		taxLabel: '\u00C1FA',
		taxRate: 0.27
	},

	// -----------------------------------------------------------------------
	// Greece
	// -----------------------------------------------------------------------
	{
		code: 'GR',
		name: 'Greece',
		nativeName: '\u0395\u03BB\u03BB\u03AC\u03B4\u03B1',
		phoneCode: '+30',
		phonePlaceholder: '691 234 5678',
		phoneMaxLength: 10,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['el', 'en'],
		flag: '\uD83C\uDDEC\uD83C\uDDF7',
		postcodeFormat: '\\d{3}\\s?\\d{2}',
		postcodePlaceholder: '105 57',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: '\u03A4\u03C1\u03B1\u03C0\u03B5\u03B6\u03B9\u03BA\u03AE \u03BC\u03B5\u03C4\u03B1\u03C6\u03BF\u03C1\u03AC'
			}
		],
		taxLabel: '\u03A6\u03A0\u0391',
		taxRate: 0.24
	},

	// -----------------------------------------------------------------------
	// Bulgaria
	// -----------------------------------------------------------------------
	{
		code: 'BG',
		name: 'Bulgaria',
		nativeName: '\u0411\u044A\u043B\u0433\u0430\u0440\u0438\u044F',
		phoneCode: '+359',
		phonePlaceholder: '88 123 4567',
		phoneMaxLength: 9,
		currency: 'BGN',
		currencySymbol: '\u043B\u0432',
		languages: ['bg', 'en'],
		flag: '\uD83C\uDDE7\uD83C\uDDEC',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '1000',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: '\u0411\u0430\u043D\u043A\u043E\u0432 \u043F\u0440\u0435\u0432\u043E\u0434'
			}
		],
		taxLabel: '\u0414\u0414\u0421',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// Croatia
	// -----------------------------------------------------------------------
	{
		code: 'HR',
		name: 'Croatia',
		nativeName: 'Hrvatska',
		phoneCode: '+385',
		phonePlaceholder: '91 234 5678',
		phoneMaxLength: 9,
		currency: 'EUR',
		currencySymbol: '\u20AC',
		languages: ['hr', 'en'],
		flag: '\uD83C\uDDED\uD83C\uDDF7',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '10000',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Bankovni prijenos'
			}
		],
		taxLabel: 'PDV',
		taxRate: 0.25
	},

	// -----------------------------------------------------------------------
	// UAE
	// -----------------------------------------------------------------------
	{
		code: 'AE',
		name: 'United Arab Emirates',
		nativeName: '\u0627\u0644\u0625\u0645\u0627\u0631\u0627\u062A',
		phoneCode: '+971',
		phonePlaceholder: '50 123 4567',
		phoneMaxLength: 9,
		currency: 'AED',
		currencySymbol: '\u062F.\u0625',
		languages: ['en'],
		flag: '\uD83C\uDDE6\uD83C\uDDEA',
		postcodeFormat: '.*',
		postcodePlaceholder: '',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Amex'
			},
			{
				id: 'apple_pay',
				name: 'Apple Pay',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pay using Apple Pay'
			},
			{
				id: 'cash',
				name: 'Cash',
				icon: 'Banknote',
				isDefault: false,
				description: 'Pay in cash after service completion'
			}
		],
		taxLabel: 'VAT',
		taxRate: 0.05
	},

	// -----------------------------------------------------------------------
	// Australia
	// -----------------------------------------------------------------------
	{
		code: 'AU',
		name: 'Australia',
		nativeName: 'Australia',
		phoneCode: '+61',
		phonePlaceholder: '412 345 678',
		phoneMaxLength: 9,
		currency: 'AUD',
		currencySymbol: 'A$',
		languages: ['en'],
		flag: '\uD83C\uDDE6\uD83C\uDDFA',
		postcodeFormat: '\\d{4}',
		postcodePlaceholder: '2000',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Amex'
			},
			{
				id: 'payid',
				name: 'PayID',
				icon: 'Building2',
				isDefault: false,
				description: 'Pay using PayID'
			},
			{
				id: 'apple_pay',
				name: 'Apple Pay',
				icon: 'Smartphone',
				isDefault: false,
				description: 'Pay using Apple Pay'
			}
		],
		taxLabel: 'GST',
		taxRate: 0.1
	},

	// -----------------------------------------------------------------------
	// Turkey
	// -----------------------------------------------------------------------
	{
		code: 'TR',
		name: 'Turkey',
		nativeName: 'T\u00FCrkiye',
		phoneCode: '+90',
		phonePlaceholder: '532 123 4567',
		phoneMaxLength: 10,
		currency: 'TRY',
		currencySymbol: '\u20BA',
		languages: ['tr', 'en'],
		flag: '\uD83C\uDDF9\uD83C\uDDF7',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '34000',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Troy'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: 'Banka havalesi'
			}
		],
		taxLabel: 'KDV',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// Russia
	// -----------------------------------------------------------------------
	{
		code: 'RU',
		name: 'Russia',
		nativeName: '\u0420\u043E\u0441\u0441\u0438\u044F',
		phoneCode: '+7',
		phonePlaceholder: '912 345 67 89',
		phoneMaxLength: 10,
		currency: 'RUB',
		currencySymbol: '\u20BD',
		languages: ['ru', 'en'],
		flag: '\uD83C\uDDF7\uD83C\uDDFA',
		postcodeFormat: '\\d{6}',
		postcodePlaceholder: '101000',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard, Mir'
			},
			{
				id: 'sbp',
				name: 'SBP',
				icon: 'Smartphone',
				isDefault: false,
				description: '\u0421\u0438\u0441\u0442\u0435\u043C\u0430 \u0431\u044B\u0441\u0442\u0440\u044B\u0445 \u043F\u043B\u0430\u0442\u0435\u0436\u0435\u0439'
			}
		],
		taxLabel: '\u041D\u0414\u0421',
		taxRate: 0.2
	},

	// -----------------------------------------------------------------------
	// Ukraine
	// -----------------------------------------------------------------------
	{
		code: 'UA',
		name: 'Ukraine',
		nativeName: '\u0423\u043A\u0440\u0430\u0457\u043D\u0430',
		phoneCode: '+380',
		phonePlaceholder: '50 123 4567',
		phoneMaxLength: 9,
		currency: 'UAH',
		currencySymbol: '\u20B4',
		languages: ['uk', 'en'],
		flag: '\uD83C\uDDFA\uD83C\uDDE6',
		postcodeFormat: '\\d{5}',
		postcodePlaceholder: '01001',
		paymentMethods: [
			{
				id: 'card',
				name: 'Credit/Debit Card',
				icon: 'CreditCard',
				isDefault: true,
				description: 'Visa, Mastercard'
			},
			{
				id: 'bank_transfer',
				name: 'Bank Transfer',
				icon: 'Building2',
				isDefault: false,
				description: '\u0411\u0430\u043D\u043A\u0456\u0432\u0441\u044C\u043A\u0438\u0439 \u043F\u0435\u0440\u0435\u043A\u0430\u0437'
			}
		],
		taxLabel: '\u041F\u0414\u0412',
		taxRate: 0.2
	}
];

// ---------------------------------------------------------------------------
// Derived collections
// ---------------------------------------------------------------------------

/** Map of jurisdiction by country code for O(1) lookups. */
export const jurisdictionMap: Record<string, Jurisdiction> = Object.fromEntries(
	jurisdictions.map((j) => [j.code, j])
);

/** Default jurisdiction (India). */
export const defaultJurisdiction: Jurisdiction = jurisdictionMap['IN'];

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

/**
 * Find a jurisdiction by its phone code (e.g., '+91' -> India).
 * If multiple jurisdictions share the same phone code (e.g., +7 for RU/KZ),
 * the first match is returned.
 */
export function getJurisdictionByPhoneCode(phoneCode: string): Jurisdiction | undefined {
	return jurisdictions.find((j) => j.phoneCode === phoneCode);
}

/**
 * Attempt to detect the user's jurisdiction from a country code.
 * Falls back to the default jurisdiction (IN) if not found.
 */
export function detectJurisdiction(countryCode?: string): Jurisdiction {
	if (countryCode && jurisdictionMap[countryCode.toUpperCase()]) {
		return jurisdictionMap[countryCode.toUpperCase()];
	}
	return defaultJurisdiction;
}
