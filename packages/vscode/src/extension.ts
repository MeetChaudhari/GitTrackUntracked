import * as path from 'node:path';
import { execFile } from 'node:child_process';
import * as vscode from 'vscode';

const output = vscode.window.createOutputChannel('GitTrackUntracked');

function binary(): string {
  return vscode.workspace.getConfiguration('gittrackuntracked').get<string>('binaryPath', 'gitu');
}

function workspacePath(): string | undefined {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}

function run(args: string[], cwd = workspacePath()): Promise<void> {
  output.show(true);
  output.appendLine(`$ ${binary()} ${args.join(' ')}`);
  return new Promise((resolve, reject) => {
    execFile(binary(), args, { cwd }, (error, stdout, stderr) => {
      if (stdout) output.append(stdout);
      if (stderr) output.append(stderr);
      if (error) {
        vscode.window.showErrorMessage(`GitTrackUntracked failed: ${error.message}`);
        reject(error);
        return;
      }
      resolve();
    });
  });
}

function requireWorkspace(): string | undefined {
  const root = workspacePath();
  if (!root) vscode.window.showWarningMessage('Open a project folder first.');
  return root;
}

export function activate(context: vscode.ExtensionContext): void {
  context.subscriptions.push(output);
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.initializeVault', async () => {
    const remote = await vscode.window.showInputBox({ prompt: 'Private vault Git remote URL', ignoreFocusOut: true });
    if (remote) await run(['vault', 'init', '--remote', remote], undefined);
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.initializeProject', async () => {
    if (requireWorkspace()) await run(['init']);
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.addActiveFile', async () => {
    const root = requireWorkspace();
    const editor = vscode.window.activeTextEditor;
    if (!root || !editor) return;
    const relative = path.relative(root, editor.document.uri.fsPath);
    if (relative.startsWith('..') || path.isAbsolute(relative)) {
      vscode.window.showWarningMessage('The active file is outside the workspace.');
      return;
    }
    await run(['add', relative]);
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.status', async () => {
    if (requireWorkspace()) await run(['status']);
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.sync', async () => {
    if (requireWorkspace()) await run(['sync']);
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.restore', async () => {
    if (requireWorkspace()) await run(['restore']);
  }));
}

export function deactivate(): void {}
