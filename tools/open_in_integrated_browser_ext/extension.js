const vscode = require('vscode');

function activate(context) {
  const disposable = vscode.commands.registerCommand('openInIntegratedBrowser.open', async (uri) => {
    const target = uri instanceof vscode.Uri ? uri : getActiveFileUri();

    if (!target) {
      vscode.window.showWarningMessage('開く対象のファイルが見つかりません。');
      return;
    }

    const isHtml = /\.(html|htm)$/i.test(target.fsPath);
    if (!isHtml) {
      vscode.window.showWarningMessage('HTML ファイルのみ開けます。');
      return;
    }

    const fileUrl = target.toString(true);
    await vscode.commands.executeCommand('simpleBrowser.show', fileUrl);
  });

  context.subscriptions.push(disposable);
}

function getActiveFileUri() {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    return undefined;
  }
  return editor.document.uri;
}

function deactivate() {}

module.exports = {
  activate,
  deactivate,
};
