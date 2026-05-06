<script lang="ts">
	import { onMount } from 'svelte';

	const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

	let overviewId = $state('');
	let localPath = $state('');
	let selectedProjectName = $state('');
	let features = $state<string[]>([]);
	let components = $state<string[]>([]);
	let candidates = $state<string[]>([]);
	let directoryStatus = $state('');
	let errorMessage = $state('');
	let isLoading = $state(false);

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		overviewId = params.get('overviewId') ?? '';
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
	<h1>プロジェクト名・構成要素確認</h1>
	<p>保存済みシステム概要から機能抽出・プロジェクト名候補生成を行い、初期ディレクトリを作成します。</p>

	<div class="card">
		<label for="overviewId">概要ID</label>
		<input id="overviewId" bind:value={overviewId} placeholder="UUID" />
		<button onclick={handleExtract} disabled={isLoading || !overviewId.trim()}>
			{isLoading ? '処理中...' : '機能抽出・候補生成'}
		</button>
	</div>

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
			<select bind:value={selectedProjectName}>
				{#each candidates as c}
					<option value={c}>{c}</option>
				{/each}
			</select>

			<label for="localPath">ローカル作成先パス（絶対パス）</label>
			<input id="localPath" bind:value={localPath} placeholder="/Users/yourname/gitspace" />

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

	input,
	select {
		padding: 0.55rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid #ccc;
		border-radius: 6px;
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
