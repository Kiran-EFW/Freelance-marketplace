// ---------------------------------------------------------------------------
// Service Catalogue — flat category tree with helpers
// ---------------------------------------------------------------------------

export interface ServiceCategory {
	id: string;
	slug: string;
	icon: string;
	parentId: string | null;
	translationKey: string;
	isActive: boolean;
	color: string;
	order: number;
}

// ---------------------------------------------------------------------------
// Category data
// ---------------------------------------------------------------------------

export const categories: ServiceCategory[] = [
	// -----------------------------------------------------------------------
	// 1. Home Repair & Maintenance
	// -----------------------------------------------------------------------
	{
		id: 'home_repair',
		slug: 'home-repair',
		icon: 'Wrench',
		parentId: null,
		translationKey: 'categories.home_repair',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 1
	},
	{
		id: 'plumbing',
		slug: 'plumbing',
		icon: 'Droplets',
		parentId: 'home_repair',
		translationKey: 'categories.plumbing',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 1
	},
	{
		id: 'electrical',
		slug: 'electrical',
		icon: 'Zap',
		parentId: 'home_repair',
		translationKey: 'categories.electrical',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 2
	},
	{
		id: 'carpentry',
		slug: 'carpentry',
		icon: 'Hammer',
		parentId: 'home_repair',
		translationKey: 'categories.carpentry',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 3
	},
	{
		id: 'painting',
		slug: 'painting',
		icon: 'Paintbrush',
		parentId: 'home_repair',
		translationKey: 'categories.painting',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 4
	},
	{
		id: 'appliance_repair',
		slug: 'appliance-repair',
		icon: 'Refrigerator',
		parentId: 'home_repair',
		translationKey: 'categories.appliance_repair',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 5
	},
	{
		id: 'masonry_civil',
		slug: 'masonry-civil',
		icon: 'Construction',
		parentId: 'home_repair',
		translationKey: 'categories.masonry',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 6
	},
	{
		id: 'roofing',
		slug: 'roofing',
		icon: 'Home',
		parentId: 'home_repair',
		translationKey: 'categories.roofing',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 7
	},
	{
		id: 'welding_fabrication',
		slug: 'welding-fabrication',
		icon: 'Flame',
		parentId: 'home_repair',
		translationKey: 'categories.welding',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 8
	},
	{
		id: 'glass_work',
		slug: 'glass-work',
		icon: 'Square',
		parentId: 'home_repair',
		translationKey: 'categories.glass_work',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 9
	},
	{
		id: 'locksmith',
		slug: 'locksmith',
		icon: 'Lock',
		parentId: 'home_repair',
		translationKey: 'categories.locksmith',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 10
	},
	{
		id: 'pest_control',
		slug: 'pest-control',
		icon: 'Bug',
		parentId: 'home_repair',
		translationKey: 'categories.pest_control',
		isActive: true,
		color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
		order: 11
	},

	// -----------------------------------------------------------------------
	// 2. Cleaning
	// -----------------------------------------------------------------------
	{
		id: 'cleaning',
		slug: 'cleaning',
		icon: 'Sparkles',
		parentId: null,
		translationKey: 'categories.cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 2
	},
	{
		id: 'home_cleaning',
		slug: 'home-cleaning',
		icon: 'Home',
		parentId: 'cleaning',
		translationKey: 'categories.home_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 1
	},
	{
		id: 'kitchen_deep_cleaning',
		slug: 'kitchen-deep-cleaning',
		icon: 'CookingPot',
		parentId: 'cleaning',
		translationKey: 'categories.kitchen_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 2
	},
	{
		id: 'bathroom_deep_cleaning',
		slug: 'bathroom-deep-cleaning',
		icon: 'Bath',
		parentId: 'cleaning',
		translationKey: 'categories.bathroom_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 3
	},
	{
		id: 'sofa_upholstery',
		slug: 'sofa-upholstery',
		icon: 'Sofa',
		parentId: 'cleaning',
		translationKey: 'categories.sofa_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 4
	},
	{
		id: 'carpet_cleaning',
		slug: 'carpet-cleaning',
		icon: 'Layers',
		parentId: 'cleaning',
		translationKey: 'categories.carpet_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 5
	},
	{
		id: 'water_tank_cleaning',
		slug: 'water-tank-cleaning',
		icon: 'Container',
		parentId: 'cleaning',
		translationKey: 'categories.water_tank',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 6
	},
	{
		id: 'office_cleaning',
		slug: 'office-cleaning',
		icon: 'Building2',
		parentId: 'cleaning',
		translationKey: 'categories.office_cleaning',
		isActive: true,
		color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
		order: 7
	},

	// -----------------------------------------------------------------------
	// 3. Beauty & Wellness
	// -----------------------------------------------------------------------
	{
		id: 'beauty_wellness',
		slug: 'beauty-wellness',
		icon: 'Scissors',
		parentId: null,
		translationKey: 'categories.beauty_wellness',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 3
	},
	{
		id: 'salon_women',
		slug: 'salon-women',
		icon: 'Scissors',
		parentId: 'beauty_wellness',
		translationKey: 'categories.salon_women',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 1
	},
	{
		id: 'salon_men',
		slug: 'salon-men',
		icon: 'Scissors',
		parentId: 'beauty_wellness',
		translationKey: 'categories.salon_men',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 2
	},
	{
		id: 'massage_spa',
		slug: 'massage-spa',
		icon: 'Heart',
		parentId: 'beauty_wellness',
		translationKey: 'categories.massage_spa',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 3
	},
	{
		id: 'makeup_artist',
		slug: 'makeup-artist',
		icon: 'Palette',
		parentId: 'beauty_wellness',
		translationKey: 'categories.makeup_artist',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 4
	},
	{
		id: 'personal_trainer',
		slug: 'personal-trainer',
		icon: 'Dumbbell',
		parentId: 'beauty_wellness',
		translationKey: 'categories.personal_trainer',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 5
	},
	{
		id: 'yoga_instructor',
		slug: 'yoga-instructor',
		icon: 'Activity',
		parentId: 'beauty_wellness',
		translationKey: 'categories.yoga',
		isActive: true,
		color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
		order: 6
	},

	// -----------------------------------------------------------------------
	// 4. Professional Services
	// -----------------------------------------------------------------------
	{
		id: 'professional_services',
		slug: 'professional-services',
		icon: 'Briefcase',
		parentId: null,
		translationKey: 'categories.professional_services',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 4
	},
	{
		id: 'legal_services',
		slug: 'legal-services',
		icon: 'Scale',
		parentId: 'professional_services',
		translationKey: 'categories.legal',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 1
	},
	{
		id: 'finance_ca',
		slug: 'finance-ca',
		icon: 'Calculator',
		parentId: 'professional_services',
		translationKey: 'categories.finance_ca',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 2
	},
	{
		id: 'architecture_design',
		slug: 'architecture-design',
		icon: 'Ruler',
		parentId: 'professional_services',
		translationKey: 'categories.architecture',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 3
	},
	{
		id: 'documentation',
		slug: 'documentation',
		icon: 'FileText',
		parentId: 'professional_services',
		translationKey: 'categories.documentation',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 4
	},
	{
		id: 'business_consulting',
		slug: 'business-consulting',
		icon: 'TrendingUp',
		parentId: 'professional_services',
		translationKey: 'categories.consulting',
		isActive: true,
		color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
		order: 5
	},

	// -----------------------------------------------------------------------
	// 5. Vehicle Services
	// -----------------------------------------------------------------------
	{
		id: 'vehicle_services',
		slug: 'vehicle-services',
		icon: 'Car',
		parentId: null,
		translationKey: 'categories.vehicle_services',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 5
	},
	{
		id: 'car_mechanic',
		slug: 'car-mechanic',
		icon: 'Wrench',
		parentId: 'vehicle_services',
		translationKey: 'categories.car_mechanic',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 1
	},
	{
		id: 'two_wheeler_mechanic',
		slug: 'two-wheeler-mechanic',
		icon: 'Bike',
		parentId: 'vehicle_services',
		translationKey: 'categories.two_wheeler',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 2
	},
	{
		id: 'car_wash_detailing',
		slug: 'car-wash-detailing',
		icon: 'Droplets',
		parentId: 'vehicle_services',
		translationKey: 'categories.car_wash',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 3
	},
	{
		id: 'tyre_service',
		slug: 'tyre-service',
		icon: 'Circle',
		parentId: 'vehicle_services',
		translationKey: 'categories.tyre_service',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 4
	},
	{
		id: 'towing',
		slug: 'towing',
		icon: 'Truck',
		parentId: 'vehicle_services',
		translationKey: 'categories.towing',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 5
	},
	{
		id: 'driving_instructor',
		slug: 'driving-instructor',
		icon: 'GraduationCap',
		parentId: 'vehicle_services',
		translationKey: 'categories.driving_instructor',
		isActive: true,
		color: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
		order: 6
	},

	// -----------------------------------------------------------------------
	// 6. Education & Tutoring
	// -----------------------------------------------------------------------
	{
		id: 'education_tutoring',
		slug: 'education-tutoring',
		icon: 'GraduationCap',
		parentId: null,
		translationKey: 'categories.education_tutoring',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 6
	},
	{
		id: 'academic_tutor',
		slug: 'academic-tutor',
		icon: 'BookOpen',
		parentId: 'education_tutoring',
		translationKey: 'categories.academic_tutor',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 1
	},
	{
		id: 'music_teacher',
		slug: 'music-teacher',
		icon: 'Music',
		parentId: 'education_tutoring',
		translationKey: 'categories.music_teacher',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 2
	},
	{
		id: 'dance_teacher',
		slug: 'dance-teacher',
		icon: 'Music2',
		parentId: 'education_tutoring',
		translationKey: 'categories.dance_teacher',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 3
	},
	{
		id: 'language_tutor',
		slug: 'language-tutor',
		icon: 'Languages',
		parentId: 'education_tutoring',
		translationKey: 'categories.language_tutor',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 4
	},
	{
		id: 'sports_coaching',
		slug: 'sports-coaching',
		icon: 'Medal',
		parentId: 'education_tutoring',
		translationKey: 'categories.sports_coaching',
		isActive: true,
		color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400',
		order: 5
	},

	// -----------------------------------------------------------------------
	// 7. Care Services
	// -----------------------------------------------------------------------
	{
		id: 'care_services',
		slug: 'care-services',
		icon: 'HeartHandshake',
		parentId: null,
		translationKey: 'categories.care_services',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 7
	},
	{
		id: 'elderly_care',
		slug: 'elderly-care',
		icon: 'Heart',
		parentId: 'care_services',
		translationKey: 'categories.elderly_care',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 1
	},
	{
		id: 'baby_sitter_nanny',
		slug: 'baby-sitter-nanny',
		icon: 'Baby',
		parentId: 'care_services',
		translationKey: 'categories.babysitter',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 2
	},
	{
		id: 'home_nurse',
		slug: 'home-nurse',
		icon: 'Stethoscope',
		parentId: 'care_services',
		translationKey: 'categories.home_nurse',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 3
	},
	{
		id: 'physiotherapist',
		slug: 'physiotherapist',
		icon: 'Activity',
		parentId: 'care_services',
		translationKey: 'categories.physiotherapist',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 4
	},
	{
		id: 'pet_grooming',
		slug: 'pet-grooming',
		icon: 'PawPrint',
		parentId: 'care_services',
		translationKey: 'categories.pet_grooming',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 5
	},
	{
		id: 'veterinary',
		slug: 'veterinary',
		icon: 'Stethoscope',
		parentId: 'care_services',
		translationKey: 'categories.veterinary',
		isActive: true,
		color: 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
		order: 6
	},

	// -----------------------------------------------------------------------
	// 8. Events & Occasions
	// -----------------------------------------------------------------------
	{
		id: 'events_occasions',
		slug: 'events-occasions',
		icon: 'PartyPopper',
		parentId: null,
		translationKey: 'categories.events_occasions',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 8
	},
	{
		id: 'catering',
		slug: 'catering',
		icon: 'UtensilsCrossed',
		parentId: 'events_occasions',
		translationKey: 'categories.catering',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 1
	},
	{
		id: 'decoration',
		slug: 'decoration',
		icon: 'Flower2',
		parentId: 'events_occasions',
		translationKey: 'categories.decoration',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 2
	},
	{
		id: 'photography',
		slug: 'photography',
		icon: 'Camera',
		parentId: 'events_occasions',
		translationKey: 'categories.photography',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 3
	},
	{
		id: 'videography',
		slug: 'videography',
		icon: 'Video',
		parentId: 'events_occasions',
		translationKey: 'categories.videography',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 4
	},
	{
		id: 'wedding_planning',
		slug: 'wedding-planning',
		icon: 'Heart',
		parentId: 'events_occasions',
		translationKey: 'categories.wedding_planning',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 5
	},
	{
		id: 'sound_dj',
		slug: 'sound-dj',
		icon: 'Speaker',
		parentId: 'events_occasions',
		translationKey: 'categories.dj_music',
		isActive: true,
		color: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
		order: 6
	},

	// -----------------------------------------------------------------------
	// 9. Moving & Logistics
	// -----------------------------------------------------------------------
	{
		id: 'moving_logistics',
		slug: 'moving-logistics',
		icon: 'Truck',
		parentId: null,
		translationKey: 'categories.moving_logistics',
		isActive: true,
		color: 'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
		order: 9
	},
	{
		id: 'packers_movers',
		slug: 'packers-movers',
		icon: 'Package',
		parentId: 'moving_logistics',
		translationKey: 'categories.packers_movers',
		isActive: true,
		color: 'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
		order: 1
	},
	{
		id: 'furniture_assembly',
		slug: 'furniture-assembly',
		icon: 'Armchair',
		parentId: 'moving_logistics',
		translationKey: 'categories.furniture_assembly',
		isActive: true,
		color: 'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
		order: 2
	},
	{
		id: 'local_courier',
		slug: 'local-courier',
		icon: 'Send',
		parentId: 'moving_logistics',
		translationKey: 'categories.courier',
		isActive: true,
		color: 'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
		order: 3
	},
	{
		id: 'junk_removal',
		slug: 'junk-removal',
		icon: 'Trash2',
		parentId: 'moving_logistics',
		translationKey: 'categories.junk_removal',
		isActive: true,
		color: 'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
		order: 4
	},

	// -----------------------------------------------------------------------
	// 10. Tech & Digital
	// -----------------------------------------------------------------------
	{
		id: 'tech_digital',
		slug: 'tech-digital',
		icon: 'Monitor',
		parentId: null,
		translationKey: 'categories.tech_digital',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 10
	},
	{
		id: 'computer_laptop_repair',
		slug: 'computer-laptop-repair',
		icon: 'Laptop',
		parentId: 'tech_digital',
		translationKey: 'categories.computer_repair',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 1
	},
	{
		id: 'mobile_phone_repair',
		slug: 'mobile-phone-repair',
		icon: 'Smartphone',
		parentId: 'tech_digital',
		translationKey: 'categories.mobile_repair',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 2
	},
	{
		id: 'cctv_installation',
		slug: 'cctv-installation',
		icon: 'Eye',
		parentId: 'tech_digital',
		translationKey: 'categories.cctv_install',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 3
	},
	{
		id: 'wifi_network_setup',
		slug: 'wifi-network-setup',
		icon: 'Wifi',
		parentId: 'tech_digital',
		translationKey: 'categories.network_setup',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 4
	},
	{
		id: 'smart_home_setup',
		slug: 'smart-home-setup',
		icon: 'Home',
		parentId: 'tech_digital',
		translationKey: 'categories.smart_home',
		isActive: true,
		color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400',
		order: 5
	},

	// -----------------------------------------------------------------------
	// 11. Crop & Land Services
	// -----------------------------------------------------------------------
	{
		id: 'crop_land',
		slug: 'crop-land',
		icon: 'TreePine',
		parentId: null,
		translationKey: 'categories.crop_land',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 11
	},
	{
		id: 'gardening_landscaping',
		slug: 'gardening-landscaping',
		icon: 'Flower2',
		parentId: 'crop_land',
		translationKey: 'categories.gardening_landscaping',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 1
	},
	{
		id: 'tree_cutting_pruning',
		slug: 'tree-cutting-pruning',
		icon: 'Axe',
		parentId: 'crop_land',
		translationKey: 'categories.tree_cutting_pruning',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 2
	},
	{
		id: 'pest_spraying',
		slug: 'pest-spraying',
		icon: 'Bug',
		parentId: 'crop_land',
		translationKey: 'categories.pest_spraying',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 3
	},
	{
		id: 'irrigation_well_services',
		slug: 'irrigation-well-services',
		icon: 'Droplets',
		parentId: 'crop_land',
		translationKey: 'categories.irrigation_well_services',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 4
	},
	{
		id: 'coconut_tree_climbing',
		slug: 'coconut-tree-climbing',
		icon: 'TreePine',
		parentId: 'crop_land',
		translationKey: 'categories.coconut_tree_climbing',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 5
	},
	{
		id: 'tractor_farm_equipment',
		slug: 'tractor-farm-equipment',
		icon: 'Truck',
		parentId: 'crop_land',
		translationKey: 'categories.tractor_farm_equipment',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 6
	},
	{
		id: 'ploughing_land_prep',
		slug: 'ploughing-land-prep',
		icon: 'Layers',
		parentId: 'crop_land',
		translationKey: 'categories.ploughing_land_prep',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 7
	},
	{
		id: 'crop_harvesting',
		slug: 'crop-harvesting',
		icon: 'Wheat',
		parentId: 'crop_land',
		translationKey: 'categories.crop_harvesting',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 8
	},
	{
		id: 'seed_sapling_supply',
		slug: 'seed-sapling-supply',
		icon: 'Sprout',
		parentId: 'crop_land',
		translationKey: 'categories.seed_sapling_supply',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 9
	},
	{
		id: 'fertilizer_manure',
		slug: 'fertilizer-manure',
		icon: 'FlaskConical',
		parentId: 'crop_land',
		translationKey: 'categories.fertilizer_manure',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 10
	},
	{
		id: 'soil_testing',
		slug: 'soil-testing',
		icon: 'TestTube',
		parentId: 'crop_land',
		translationKey: 'categories.soil_testing',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 11
	},
	{
		id: 'borewell_drilling',
		slug: 'borewell-drilling',
		icon: 'Drill',
		parentId: 'crop_land',
		translationKey: 'categories.borewell_drilling',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 12
	},
	{
		id: 'farm_fencing',
		slug: 'farm-fencing',
		icon: 'Fence',
		parentId: 'crop_land',
		translationKey: 'categories.farm_fencing',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 13
	},
	{
		id: 'organic_farming',
		slug: 'organic-farming',
		icon: 'Leaf',
		parentId: 'crop_land',
		translationKey: 'categories.organic_farming',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 14
	},
	{
		id: 'dairy_animal_husbandry',
		slug: 'dairy-animal-husbandry',
		icon: 'Beef',
		parentId: 'crop_land',
		translationKey: 'categories.dairy_animal_husbandry',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 15
	},
	{
		id: 'aquaculture_fisheries',
		slug: 'aquaculture-fisheries',
		icon: 'Fish',
		parentId: 'crop_land',
		translationKey: 'categories.aquaculture_fisheries',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 16
	},
	{
		id: 'farm_labor',
		slug: 'farm-labor',
		icon: 'Users',
		parentId: 'crop_land',
		translationKey: 'categories.farm_labor',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 17
	},
	{
		id: 'crop_transport',
		slug: 'crop-transport',
		icon: 'Truck',
		parentId: 'crop_land',
		translationKey: 'categories.crop_transport',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 18
	},
	{
		id: 'agri_consulting',
		slug: 'agri-consulting',
		icon: 'BookOpen',
		parentId: 'crop_land',
		translationKey: 'categories.agri_consulting',
		isActive: true,
		color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
		order: 19
	},

	// -----------------------------------------------------------------------
	// 12. Construction & Civil
	// -----------------------------------------------------------------------
	{
		id: 'construction',
		slug: 'construction',
		icon: 'HardHat',
		parentId: null,
		translationKey: 'categories.construction',
		isActive: true,
		color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
		order: 12
	},
	{
		id: 'general_contractor',
		slug: 'general-contractor',
		icon: 'Building',
		parentId: 'construction',
		translationKey: 'categories.general_contractor',
		isActive: true,
		color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
		order: 1
	},
	{
		id: 'civil_engineer',
		slug: 'civil-engineer',
		icon: 'Ruler',
		parentId: 'construction',
		translationKey: 'categories.civil_engineer',
		isActive: true,
		color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
		order: 2
	},
	{
		id: 'solar_panel_installation',
		slug: 'solar-panel-installation',
		icon: 'Sun',
		parentId: 'construction',
		translationKey: 'categories.solar_panel',
		isActive: true,
		color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
		order: 3
	},
	{
		id: 'surveyor',
		slug: 'surveyor',
		icon: 'Map',
		parentId: 'construction',
		translationKey: 'categories.surveyor',
		isActive: true,
		color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
		order: 4
	}
];

// ---------------------------------------------------------------------------
// Derived collections
// ---------------------------------------------------------------------------

/** All top-level categories (parentId === null), sorted by display order. */
export const topLevelCategories: ServiceCategory[] = categories
	.filter((c) => c.parentId === null)
	.sort((a, b) => a.order - b.order);

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

/**
 * Return every direct subcategory of the given parent, sorted by display order.
 */
export function getSubcategories(parentId: string): ServiceCategory[] {
	return categories
		.filter((c) => c.parentId === parentId)
		.sort((a, b) => a.order - b.order);
}

/**
 * Find a single category by its URL slug.
 */
export function getCategoryBySlug(slug: string): ServiceCategory | undefined {
	return categories.find((c) => c.slug === slug);
}

/**
 * Find a single category by its unique id.
 */
export function getCategoryById(id: string): ServiceCategory | undefined {
	return categories.find((c) => c.id === id);
}
