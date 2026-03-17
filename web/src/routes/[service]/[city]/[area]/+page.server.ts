export const load = async ({ params }: { params: { service: string; city: string; area: string } }) => {
	const { service, city, area } = params;

	const capitalize = (w: string) => w.charAt(0).toUpperCase() + w.slice(1);
	const serviceLabel = service.split('-').map(capitalize).join(' ');
	const cityLabel = city.split('-').map(capitalize).join(' ');
	const areaLabel = area.split('-').map(capitalize).join(' ');

	// In production, this would call the API
	// const data = await api.seo.getLandingData(service, city, area);

	const providers = [
		{
			id: '1',
			name: 'Suresh Nair',
			rating: 4.8,
			reviewCount: 156,
			skills: [serviceLabel, 'Maintenance', 'Repairs'],
			hourlyRate: 500,
			distance: '1.2 km',
			isVerified: true,
			completedJobs: 156
		},
		{
			id: '2',
			name: 'Lakshmi Bai',
			rating: 4.9,
			reviewCount: 210,
			skills: [serviceLabel, 'Installation', 'Inspection'],
			hourlyRate: 600,
			distance: '2.1 km',
			isVerified: true,
			completedJobs: 210
		},
		{
			id: '3',
			name: 'Deepak Kumar',
			rating: 4.5,
			reviewCount: 78,
			skills: [serviceLabel, 'Emergency', 'Renovation'],
			hourlyRate: 450,
			distance: '3.5 km',
			isVerified: true,
			completedJobs: 78
		},
		{
			id: '4',
			name: 'Mohan Rao',
			rating: 4.7,
			reviewCount: 142,
			skills: [serviceLabel, 'Commercial', 'Residential'],
			hourlyRate: 550,
			distance: '1.8 km',
			isVerified: true,
			completedJobs: 142
		},
		{
			id: '5',
			name: 'Priya Sharma',
			rating: 4.6,
			reviewCount: 88,
			skills: [serviceLabel, 'Consultation', 'Planning'],
			hourlyRate: 400,
			distance: '4.2 km',
			isVerified: false,
			completedJobs: 88
		},
		{
			id: '6',
			name: 'Kiran Rao',
			rating: 4.4,
			reviewCount: 65,
			skills: [serviceLabel, 'Weekend', 'Express'],
			hourlyRate: 350,
			distance: '2.8 km',
			isVerified: true,
			completedJobs: 65
		}
	];

	const faqs = [
		{
			question: `How much does ${serviceLabel.toLowerCase()} cost in ${areaLabel}?`,
			answer: `The average cost for ${serviceLabel.toLowerCase()} services in ${areaLabel}, ${cityLabel} ranges from Rs. 300 to Rs. 800 per hour, depending on the complexity of the job and the provider's experience.`
		},
		{
			question: `How do I find trusted ${serviceLabel.toLowerCase()} providers near me?`,
			answer: `Seva verifies all providers through a comprehensive KYC process. You can browse verified providers, check their ratings, read reviews, and compare quotes before hiring.`
		},
		{
			question: `Can I get emergency ${serviceLabel.toLowerCase()} services?`,
			answer: `Yes, many providers on Seva offer emergency services. Filter for "Express" or "Emergency" providers, or post an urgent job to get quick responses from available providers.`
		},
		{
			question: `What if I am not satisfied with the service?`,
			answer: `Seva offers a dispute resolution process. If you are not satisfied, you can file a dispute within 48 hours of job completion. Our mediation team will review and resolve the issue.`
		},
		{
			question: `Are the providers insured?`,
			answer: `Many providers on Seva carry their own insurance. You can check each provider's profile for insurance status. Seva also offers buyer protection for jobs booked through the platform.`
		}
	];

	const stats = {
		totalProviders: providers.length + 12,
		avgRating: 4.6,
		completedJobs: 2450,
		avgResponseTime: '2 hours'
	};

	const relatedServices = [
		{ slug: 'plumbing', label: 'Plumbing' },
		{ slug: 'electrical', label: 'Electrical' },
		{ slug: 'cleaning', label: 'Cleaning' },
		{ slug: 'painting', label: 'Painting' },
		{ slug: 'gardening', label: 'Gardening' },
		{ slug: 'carpentry', label: 'Carpentry' }
	].filter((s) => s.slug !== service);

	const nearbyAreas = [
		{ slug: 'koramangala', label: 'Koramangala' },
		{ slug: 'indiranagar', label: 'Indiranagar' },
		{ slug: 'hsr-layout', label: 'HSR Layout' },
		{ slug: 'whitefield', label: 'Whitefield' },
		{ slug: 'jp-nagar', label: 'JP Nagar' },
		{ slug: 'marathahalli', label: 'Marathahalli' }
	].filter((a) => a.slug !== area);

	return {
		service: serviceLabel,
		serviceSlug: service,
		city: cityLabel,
		citySlug: city,
		area: areaLabel,
		areaSlug: area,
		providers,
		faqs,
		stats,
		relatedServices,
		nearbyAreas
	};
};
