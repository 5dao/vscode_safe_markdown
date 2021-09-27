# safe_markdown

markdown auto encryption and decryption 

- with .md suffix, eg. your-file-name.md
- with Right-click menu

## useage

1. make ras private key and pub key

2. config your vscode

export VSCODE_SAFE_MARKDOWN="/path-to-your/xx_rsa"

> pub key path = ${VSCODE_SAFE_MARKDOWN}+".pub"


3. Right-click your markdown file , encrypt or decrypt it.

## about encryption file

```json
{
    "key":"RSA(uuid)",
    "data":"AES-128-cbc(txt,uuid)"
}
```
