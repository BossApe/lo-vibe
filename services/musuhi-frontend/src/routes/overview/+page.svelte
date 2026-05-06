<script lang="ts">
	const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
	const MAX_LENGTH = 4096;

	let content = $state('');
	let errorMessage = $state('');
	let successMessage = $state('');
	let createdOverviewId = $state('');
	let isSubmitting = $state(false);

	function validate(): string {
		if (!content.trim()) {
			return 'システム概要を入力してください。';
		}
		if (content.length > MAX_LENGTH) {
			return `${MAX_LENGTH}文字以内で入力してください（現在: ${content.length}文字）。`;
		}
		return '';
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMessage = '';
		successMessage = '';
		createdOverviewId = '';

		const validationError = validate();
		if (validationError) {
			errorMessage = validationError;
			return;
		}

		isSubmitting = true;
		try {
			const res = await fetch(`${API_BASE}/api/v1/system-overviews`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ content })
			});

			if (res.status === 201) {
				const data = await res.json();
				createdOverviewId = data.data.id;
				successMessage = `保存しました（ID: ${createdOverviewId}）`;
				content = '';
			} else if (res.status === 422) {
				const data = await res.json();
				errorMessage = data.error?.message ?? 'バリデーションエラーが発生しました。';
			} else {
				errorMessage = 'サーバーエラーが発生しました。しばらく待ってから再試行してください。';
			}
		} catch {
			errorMessage = 'ネットワークエラーが発生しました。接続を確認してください。';
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>システム概要入力 | Musuhi</title>
</svelte:head>

<main class="container">
	<h1>システム概要入力</h1>
	<p class="description">
		開発するシステムの概要を箇条書き・メモ形式で入力してください。<br />
		入力内容はAIによる機能抽出・プロジェクト名生成に利用されます。
	</p>

	<form onsubmit={handleSubmit}>
		<div class="field">
			<label for="content">システム概要テキスト <span class="required">*</span></label>
			<textarea
				id="content"
				bind:value={content}
				placeholder="例：&#10;- ユーザ管理機能&#10;- 商品カタログ表示&#10;- カート・注文機能"
				rows={12}
				maxlength={MAX_LENGTH}
				disabled={isSubmitting}
				aria-describedby={errorMessage ? 'error-message' : undefined}
				aria-invalid={!!errorMessage}
			></textarea>
			<div class="char-count" aria-live="polite">
				{content.length} / {MAX_LENGTH} 文字
			</div>
		</div>

		{#if errorMessage}
			<div id="error-message" class="alert alert-error" role="alert">
				{errorMessage}
			</div>
		{/if}

		{#if successMessage}
			<div class="alert alert-success" role="status">
				{successMessage}
				{#if createdOverviewId}
					<div class="next-link">
						<a href={`/projects/setup?overviewId=${createdOverviewId}`}>次へ: プロジェクト名・構成要素確認へ進む</a>
					</div>
				{/if}
			</div>
		{/if}

		<div class="actions">
			<button type="submit" disabled={isSubmitting}>
				{isSubmitting ? '保存中...' : '保存する'}
			</button>
		</div>
	</form>
</main>

<style>
	.container {
		max-width: 720px;
		margin: 2rem auto;
		padding: 0 1rem;
		font-family: system-ui, sans-serif;
	}

	h1 {
		font-size: 1.5rem;
		margin-bottom: 0.5rem;
	}

	.description {
		color: #555;
		margin-bottom: 1.5rem;
		line-height: 1.6;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		margin-bottom: 1rem;
	}

	label {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.required {
		color: #c00;
	}

	textarea {
		width: 100%;
		padding: 0.6rem 0.8rem;
		font-size: 0.95rem;
		border: 1px solid #ccc;
		border-radius: 4px;
		resize: vertical;
		box-sizing: border-box;
		font-family: inherit;
		line-height: 1.5;
	}

	textarea:focus {
		outline: 2px solid #0066cc;
		border-color: transparent;
	}

	textarea[aria-invalid='true'] {
		border-color: #c00;
	}

	.char-count {
		font-size: 0.8rem;
		color: #666;
		text-align: right;
	}

	.alert {
		padding: 0.75rem 1rem;
		border-radius: 4px;
		margin-bottom: 1rem;
		font-size: 0.9rem;
	}

	.alert-error {
		background: #fff0f0;
		border: 1px solid #c00;
		color: #c00;
	}

	.alert-success {
		background: #f0fff4;
		border: 1px solid #0a0;
		color: #050;
	}

	.next-link {
		margin-top: 0.4rem;
	}

	.next-link a {
		color: #0645ad;
		text-decoration: underline;
	}

	.actions {
		display: flex;
		justify-content: flex-end;
	}

	button[type='submit'] {
		padding: 0.6rem 1.5rem;
		background: #0066cc;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 0.95rem;
		cursor: pointer;
	}

	button[type='submit']:hover:not(:disabled) {
		background: #0055aa;
	}

	button[type='submit']:disabled {
		background: #99b;
		cursor: not-allowed;
	}
</style>
