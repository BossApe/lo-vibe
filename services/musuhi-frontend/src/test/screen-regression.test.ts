import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

describe('SCR-001/002 回帰防止テスト', () => {
	it('SCR-001はAPIベースURLの既定値としてlocalhost:8080を持つ', () => {
		const file = resolve('src/routes/overview/+page.svelte');
		const source = readFileSync(file, 'utf-8');

		expect(source).toContain("import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'");
		expect(source).toContain("await goto(`/projects/setup?overviewId=${data.data.id}`)");
		expect(source).toContain('if (overviewId)');
		expect(source).toContain('/api/v1/system-overviews/${overviewId}');
		expect(source).toContain("method: 'PUT'");
	});

	it('SCR-002はoverviewId起点で抽出・候補生成APIを呼び出す', () => {
		const file = resolve('src/routes/projects/setup/+page.svelte');
		const source = readFileSync(file, 'utf-8');

		expect(source).toContain("const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';");
		expect(source).toContain("postJson('/api/v1/projects/extract-features', { overviewId })");
		expect(source).toContain("postJson('/api/v1/projects/suggest-name', { overviewId })");
		expect(source).toContain('href={`/overview?overviewId=${overviewId}`}');
		expect(source).toContain('selectedCandidateReason');
		expect(source).toContain('システム概要のキーワードから生成した候補です。');
	});
});
