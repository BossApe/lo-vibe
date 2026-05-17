<script lang="ts">
	import { onMount } from 'svelte';

	type ProjectNameCandidate = {
		name: string;
		reason?: string;
		aiSuggested: boolean;
	};

	const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
	const BASE_DIR = '/Users/m.nohara/gitspace/';
	const initialOverviewId =
		typeof window !== 'undefined'
			? (new URLSearchParams(window.location.search).get('overviewId') ?? '')
			: '';

	let overviewId = $state(initialOverviewId);
	let localPath = $state('');
	let localPathEdited = $state(false);
	let selectedProjectName = $state('');

	$effect(() => {
		if (!localPathEdited && selectedProjectName) {
			localPath = BASE_DIR + selectedProjectName;
		}
	});
	let features = $state<string[]>([]);
	let components = $state<string[]>([]);
	let candidates = $state<string[]>([]);
	let candidateItems = $state<ProjectNameCandidate[]>([]);
	let directoryStatus = $state('');
	let errorMessage = $state('');
	let isLoading = $state(false);
	let isSwitchingProfile = $state(false);
	let hasAutoExtracted = $state(false);
	let profileEnabled = $state(false);
	let modelProfile = $state('balanced');
	let availableProfiles = $state<string[]>(['fast', 'balanced', 'quality']);

	let selectedCandidate = $derived(
		candidateItems.find((item) => item.name === selectedProjectName)
	);
	let selectedCandidateReason = $derived.by(() => {
		if (!selectedCandidate) {
			return '';
		}
		if (selectedCandidate.reason && selectedCandidate.reason.trim()) {
			return selectedCandidate.reason;
		}
		return 'システム概要のキーワードから生成した候補です。必要に応じて自由入力で調整してください。';
	});

	onMount(() => {
		if (!overviewId) {
			const params = new URLSearchParams(window.location.search);
			overviewId = params.get('overviewId') ?? '';
		}
		void loadNameSuggestionProfile();
	});

	$effect(() => {
		if (!hasAutoExtracted && overviewId.trim()) {
			hasAutoExtracted = true;
			void handleExtract();
		}
	});

	async function postJson(path: string, body: unknown): Promise<any> {
		const res = await fetch(`${API_BASE}${path}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		const payload = await res.json().catch(() => ({}));
		if (!res.ok) {
			throw new Error(payload?.error?.message ?? 'APIエラーが発生しました。');
		}
		return payload;
	}

	async function getJson(path: string): Promise<any> {
		const res = await fetch(`${API_BASE}${path}`);
		const payload = await res.json().catch(() => ({}));
		if (!res.ok) {
			throw new Error(payload?.error?.message ?? 'APIエラーが発生しました。');
		}
		return payload;
	}

	async function putJson(path: string, body: unknown): Promise<any> {
		const res = await fetch(`${API_BASE}${path}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		const payload = await res.json().catch(() => ({}));
		if (!res.ok) {
			throw new Error(payload?.error?.message ?? 'APIエラーが発生しました。');
		}
		return payload;
	}

	async function loadNameSuggestionProfile() {
		try {
			const result = await getJson('/api/v1/projects/name-suggestion-profile');
			profileEnabled = Boolean(result.data?.enabled);
			availableProfiles = result.data?.availableProfiles ?? ['fast', 'balanced', 'quality'];
			modelProfile = result.data?.profile ?? 'balanced';
		} catch {
			profileEnabled = false;
		}
	}

	async function handleProfileChange(event: Event) {
		const next = (event.currentTarget as HTMLSelectElement).value;
		if (!profileEnabled || !next || next === modelProfile) {
			return;
		}

		isSwitchingProfile = true;
		errorMessage = '';
		try {
			const result = await putJson('/api/v1/projects/name-suggestion-profile', { profile: next });
			modelProfile = result.data?.profile ?? next;
			await handleExtract();
		} catch (e) {
			errorMessage = e instanceof Error ? e.message : 'モデル運用モードの切替に失敗しました。';
		} finally {
			isSwitchingProfile = false;
		}
	}

	async function handleExtract() {
		errorMessage = '';
		directoryStatus = '';
		isLoading = true;
		try {
			const extraction = await postJson('/api/v1/projects/extract-features', { overviewId });
			features = extraction.data?.features ?? [];
			components = extraction.data?.components ?? [];

			const nameResult = await postJson('/api/v1/projects/suggest-name', { overviewId });
			candidates = nameResult.data?.candidates ?? [];
			const items = nameResult.data?.items ?? [];
			candidateItems =
				items.length > 0
					? items
					: candidates.map((name: string) => ({ name, aiSuggested: false }));
			if (!selectedProjectName && candidates.length > 0) {
				selectedProjectName = candidates[0];
			}
		} catch (e) {
			errorMessage = e instanceof Error ? e.message : '処理に失敗しました。';
		} finally {
			isLoading = false;
		}
	}

	async function handleInitDirectory() {
		errorMessage = '';
		directoryStatus = '';
		isLoading = true;
		try {
			const result = await postJson('/api/v1/projects/init-directory', {
				projectName: selectedProjectName
			});
			directoryStatus = result.data?.directoryStatus ?? 'success';
		} catch (e) {
			errorMessage = e instanceof Error ? e.message : '初期ディレクトリ作成に失敗しました。';
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>プロジェクト名・構成要素確認 | Musuhi</title>
</svelte:head>

<main class="container">
	<p class="back-link"><a href={`/overview?overviewId=${overviewId}`}>概要入力に戻る</a></p>
	<h1>プロジェクト名・構成要素確認</h1>
	<p>保存済みシステム概要から機能抽出・プロジェクト名候補生成を行い、初期ディレクトリを作成します。</p>

	{#if isLoading && features.length === 0 && candidates.length === 0 && !errorMessage}
		<div class="card">
			<p>システム概要をもとに機能抽出とプロジェクト名候補生成を実行しています...</p>
		</div>
	{/if}

	{#if features.length > 0}
		<div class="card">
			<h2>抽出された機能</h2>
			<ul>
				{#each features as f}
					<li>{f}</li>
				{/each}
			</ul>
			<h2>構成要素</h2>
			<ul>
				{#each components as c}
					<li>{c}</li>
				{/each}
			</ul>
		</div>
	{/if}

	{#if candidates.length > 0}
		<div class="card">
			<h2>プロジェクト名候補</h2>
			{#if profileEnabled}
				<label for="modelProfile">モデル運用モード</label>
				<select id="modelProfile" bind:value={modelProfile} onchange={handleProfileChange} disabled={isSwitchingProfile || isLoading}>
					{#each availableProfiles as profile}
						<option value={profile}>{profile}</option>
					{/each}
				</select>
			{/if}
			<label for="projectName">プロジェクト名（候補から選択またはキーボード入力）</label>
			<input
				id="projectName"
				list="project-name-candidates"
				bind:value={selectedProjectName}
				placeholder="プロジェクト名を入力または候補から選択"
			/>
			<datalist id="project-name-candidates">
				{#each candidates as c}
					<option value={c}>{c}</option>
				{/each}
			</datalist>

			{#if selectedCandidateReason}
				<p class="candidate-reason">{selectedCandidateReason}</p>
			{/if}

			<label for="localPath">ローカル作成先パス（絶対パス）</label>

			<button onclick={handleInitDirectory} disabled={isLoading || !selectedProjectName || !localPath.trim()}>
				{isLoading ? '作成中...' : '初期ディレクトリ作成'}
			</button>
		</div>
	{/if}

	{#if directoryStatus}
		<p class="success">ディレクトリ作成結果: {directoryStatus}</p>
	{/if}

	{#if errorMessage}
		<p class="error" role="alert">{errorMessage}</p>
	{/if}
</main>

<style>
	.container {
		max-width: 840px;
		margin: 2rem auto;
		padding: 0 1rem;
		font-family: system-ui, sans-serif;
	}

	.card {
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 1rem;
		margin-top: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.back-link {
		margin: 0 0 0.75rem;
	}

	.back-link a {
		color: #1663c7;
		text-decoration: none;
		font-size: 0.95rem;
	}

	.back-link a:hover {
		text-decoration: underline;
	}

	input {
		padding: 0.55rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid #ccc;
		border-radius: 6px;
	}

	select {
		padding: 0.55rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid #ccc;
		border-radius: 6px;
		background: #fff;
	}

	.candidate-reason {
		margin: -0.1rem 0 0.2rem;
		color: #555;
		font-size: 0.9rem;
		line-height: 1.6;
		white-space: pre-line;
	}

	button {
		width: fit-content;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 6px;
		background: #1663c7;
		color: #fff;
		cursor: pointer;
	}

	button:disabled {
		background: #99a;
		cursor: not-allowed;
	}

	.success {
		margin-top: 1rem;
		color: #0a7a2f;
		font-weight: 600;
	}

	.error {
		margin-top: 1rem;
		color: #b00020;
		font-weight: 600;
	}
</style>
