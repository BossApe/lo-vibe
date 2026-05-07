<script lang="ts">
	import { onMount } from 'svelte';

	type ProjectNameCandidate = {
		name: string;
		reason?: string;
		aiSuggested: boolean;
	};

	const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
	const BASE_DIR = '/Users/m.nohara/gitspace/';

	let overviewId = $state('');
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
	let hasAutoExtracted = $state(false);

	let selectedCandidate = $derived(
		candidateItems.find((item) => item.name === selectedProjectName)
	);

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		overviewId = params.get('overviewId') ?? '';
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
			candidateItems = nameResult.data?.items ?? [];
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
				projectName: selectedProjectName,
				localPath,
				template: 'default'
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

			{#if selectedCandidate?.aiSuggested && selectedCandidate.reason}
				<p class="candidate-reason">{selectedCandidate.reason}</p>
			{/if}

			<label for="localPath">ローカル作成先パス（絶対パス）</label>
			<input id="localPath" bind:value={localPath} oninput={() => localPathEdited = true} placeholder="/Users/yourname/gitspace/プロジェクト名" />

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
