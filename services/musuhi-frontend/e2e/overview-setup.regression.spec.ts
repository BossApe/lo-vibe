import { expect, test } from '@playwright/test';

test.describe('SCR-001/SCR-002 regression', () => {
	test('SCR-002 has a back link with overviewId', async ({ page, request }) => {
		const content = '- 会員管理\n- 在庫管理\n- 注文管理';
		const createRes = await request.post('http://localhost:8080/api/v1/system-overviews', {
			data: { content }
		});
		expect(createRes.ok()).toBeTruthy();
		const createBody = await createRes.json();
		const overviewId = createBody.data.id as string;

		await page.goto(`/projects/setup?overviewId=${overviewId}`);
		const backLink = page.getByRole('link', { name: '概要入力に戻る' });
		await expect(backLink).toBeVisible();
		await expect(backLink).toHaveAttribute('href', `/overview?overviewId=${overviewId}`);
	});

	test('SCR-002 shows AI reason for god-name candidate', async ({ page, request }) => {
		const content = '- 在庫管理\n- 商品管理';
		const createRes = await request.post('http://localhost:8080/api/v1/system-overviews', {
			data: { content }
		});
		expect(createRes.ok()).toBeTruthy();
		const createBody = await createRes.json();
		const overviewId = createBody.data.id as string;

		await page.goto(`/projects/setup?overviewId=${overviewId}`);
		await expect(page.getByRole('heading', { name: 'プロジェクト名候補' })).toBeVisible();
		const reason = page.locator('.candidate-reason');
		await expect(reason).toBeVisible();
		await expect(reason).not.toContainText('システム概要のキーワードから生成した候補です。');
		await expect(reason).not.toBeEmpty();
	});

	test('SCR-001 updates existing overview after returning from SCR-002', async ({ page, request }) => {
		const originalContent = '- 会員管理\n- 在庫管理\n- 注文管理';
		const updatedContent = '- 会員管理\n- 在庫管理\n- 注文管理\n- 通知機能';
		const createRes = await request.post('http://localhost:8080/api/v1/system-overviews', {
			data: { content: originalContent }
		});
		expect(createRes.ok()).toBeTruthy();
		const createBody = await createRes.json();
		const overviewId = createBody.data.id as string;

		await page.goto(`/projects/setup?overviewId=${overviewId}`);
		await page.getByRole('link', { name: '概要入力に戻る' }).click();
		await expect(page).toHaveURL(new RegExp(`/overview\\?overviewId=${overviewId}$`));

		const textarea = page.getByLabel('システム概要テキスト *');
		await expect(textarea).toBeVisible();
		await textarea.fill(updatedContent);
		await page.getByRole('button', { name: '確定して次へ' }).click();
		await expect(page).toHaveURL(new RegExp(`/projects/setup\\?overviewId=${overviewId}$`));

		const getRes = await request.get(`http://localhost:8080/api/v1/system-overviews/${overviewId}`);
		expect(getRes.ok()).toBeTruthy();
		const getBody = await getRes.json();
		expect(getBody.data.content).toBe(updatedContent);
	});
});
