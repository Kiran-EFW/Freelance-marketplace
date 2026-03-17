const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8000/api/v1';

export const load = async ({ params, fetch }: { params: { slug: string }; fetch: typeof globalThis.fetch }) => {
	const { slug } = params;

	let article = null;
	let relatedArticles: any[] = [];
	let error = '';

	try {
		const res = await fetch(`${API_BASE}/content/${slug}`);
		if (res.ok) {
			const data = await res.json();
			article = data.data;
		} else if (res.status === 404) {
			error = 'Article not found';
		} else {
			error = 'Failed to load article';
		}
	} catch {
		error = 'Failed to connect to the server';
	}

	// Fetch related articles if we got the main article
	if (article?.id) {
		try {
			const relRes = await fetch(`${API_BASE}/content/${article.id}/related?limit=5`);
			if (relRes.ok) {
				const relData = await relRes.json();
				relatedArticles = relData.data || [];
			}
		} catch {
			// Silent fail for related articles
		}
	}

	return {
		article,
		relatedArticles,
		error
	};
};
