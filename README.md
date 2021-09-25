# safe_markdown

markdown auto encryption and decryption 

- with .md suffix, eg. your-file-name.md
- with Right-click menu

## useage

1. make ras private key and pub key

2. config your vscode

```js
"safe_markdown.pri": "/path-to-your/rsa_pri.pem",
"safe_markdown.pub": "/path-to-your/rsa_pub.pem",
```

3. Right-click your markdown file , encrypt or decrypt it.

## about encryption file

```json
{
    "key":"RSA(uuid)",
    "data":"AES-128-cbc(txt,uuid)"
}
```
