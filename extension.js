const fs = require('fs');
const crypto = require("crypto");
const vscode = require('vscode');

function activate(context) {
	let disposable = vscode.commands.registerCommand('safe_markdown.encrypt', function () {
		safe_markdown_encrypt();
	});

	let disposable2 = vscode.commands.registerCommand('safe_markdown.decrypt', function () {
		safe_markdown_decrypt();
	});
	context.subscriptions.push(disposable);
	context.subscriptions.push(disposable2);
}

function deactivate() { }

module.exports = {
	activate,
	deactivate
}

let template = {
	"key": "rsa(uuid)",
	"body": "aes(txt,uuid)",
	"make_ts": "2006-01-02 15:04:05",
	"edit_count": 1,
	"last_ts": "2006-01-02 15:04:05"
}

async function safe_markdown_encrypt() {
	let activeDoc = vscode.window.activeTextEditor.document;
	// if txt is encrypt return
	try {
		let data = JSON.parse(activeDoc.getText());
		if (data.key) {
			vscode.window.showErrorMessage(`file is encrypted txt`);
			return;
		}
	} catch (e) {
		//
	}

	let cfg = vscode.workspace.getConfiguration('safe_markdown');

	let pem = "";
	try {
		pem = fs.readFileSync(cfg.get("pub"), { encoding: "ascii" });
	} catch (e) {
		vscode.window.showErrorMessage(`pub file error`);
		console.log(e);
		return
	}

	let pubKey;
	try {
		pubKey = crypto.createPublicKey({
			key: pem,
			format: "pem",
		});
	} catch (e) {
		vscode.window.showErrorMessage(`pub key error`);
		console.log(e);
		return
	}

	let uuid;
	try {
		uuid = Buffer.from(make_uuid(), 'hex')
		if (uuid.length != 16) {
			vscode.window.showErrorMessage(`uuid make error`);
			return;
		}
	} catch (e) {
		vscode.window.showErrorMessage(`uuid make error`);
		console.log(e)
		return
	}

	// console.log("===uuid", uuid.toString('hex'));

	const ecryptedKeyData = crypto.publicEncrypt(
		{
			key: pubKey,
			padding: crypto.constants.RSA_PKCS1_PADDING,
		},
		uuid,//Buffer.from(activeDoc.getText()),
	);

	let encrypt_body = aes_encrypt(Buffer.from(activeDoc.getText(), 'utf8'), uuid);

	vscode.window.activeTextEditor.edit(editBuilder => {
		var lastLine = activeDoc.lineAt(activeDoc.lineCount - 1);
		var textRange = new vscode.Range(0, 0, activeDoc.lineCount - 1, lastLine.range.end.character);

		editBuilder.replace(textRange, JSON.stringify({
			"key": ecryptedKeyData.toString('hex'),
			"body": Buffer.from(encrypt_body, 'hex').toString('hex'),
		}));
	});
}

async function safe_markdown_decrypt() {
	let activeDoc = vscode.window.activeTextEditor.document;

	let data;
	try {
		data = JSON.parse(activeDoc.getText());
	} catch (e) {
		vscode.window.showErrorMessage(`file not json`);
		console.log(e);
		return;
	}
	// console.log(data.key);

	let cfg = vscode.workspace.getConfiguration('safe_markdown');

	let pem = "";
	try {
		pem = fs.readFileSync(cfg.get("pri"), { encoding: "ascii" });
	} catch (e) {
		vscode.window.showErrorMessage(`pem file error`);
		console.log(e);
		return
	}

	let pass = await vscode.window.showInputBox(
		{
			value: "",
			prompt: "your pri pem pass",
			placeHolder: "",
			password: true,
		});

	let priKey;
	try {
		priKey = crypto.createPrivateKey({
			key: pem,
			format: "pem",
			type: "pkcs1",
			passphrase: pass,
		});
	} catch (e) {
		vscode.window.showErrorMessage(`your pem pass error`);
		console.log(e);
		return
	}

	let decryptedKeyData;
	try {
		decryptedKeyData = crypto.privateDecrypt({
			key: priKey,
			padding: crypto.constants.RSA_PKCS1_PADDING,
		},
			Buffer.from(data.key, 'hex'),
		);
	} catch (e) {
		return;
	}

	// console.log("uuid", decryptedKeyData.toString('hex'));


	let body;
	try {
		body = Buffer.from(aes_decrypt(data.body, decryptedKeyData), 'hex').toString('utf8');
	}
	catch (e) {
		vscode.window.showErrorMessage(`aes decrypt error`);
		console.log(e);
		return
	}

	vscode.window.activeTextEditor.edit(editBuilder => {
		var lastLine = activeDoc.lineAt(activeDoc.lineCount - 1);
		var textRange = new vscode.Range(0, 0, activeDoc.lineCount - 1, lastLine.range.end.character);
		editBuilder.replace(textRange, body);
	});
}

// key len([]byte)=16
function aes_encrypt(plaintext, key) {
	var cip, encrypted; encrypted = '';
	cip = crypto.createCipheriv('aes-128-cbc', Buffer.from(key, 'hex'), Buffer.from(key, 'hex'));
	encrypted += cip.update(plaintext, 'hex', 'hex');
	encrypted += cip.final('hex');
	return encrypted;
}

function aes_decrypt(encrypted, key) {
	var _decipher, decrypted, err;
	decrypted = '';
	_decipher = crypto.createDecipheriv('aes-128-cbc', Buffer.from(key, 'hex'), Buffer.from(key, 'hex'));
	decrypted += _decipher.update(encrypted, 'hex', 'hex');
	decrypted += _decipher.final('hex');
	return decrypted;
}

function make_uuid() {
	var s = [];
	var hexDigits = "0123456789abcdef";
	for (var i = 0; i < 32; i++) {
		s[i] = hexDigits.substr(Math.floor(Math.random() * 0x10), 1);
	}
	var uuid = s.join("");
	return uuid;
}