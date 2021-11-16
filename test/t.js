

function b64() {

    var s1 = "we are floating in space"
    var s1Data = Buffer.from(s1, "utf-8")
    var s1b64 = s1Data.toString("base64")
    console.log(s1b64 === "d2UgYXJlIGZsb2F0aW5nIGluIHNwYWNl")

    var b64Data = Buffer.from("d2UgYXJlIGZsb2F0aW5nIGluIHNwYWNl", 'base64')
    console.log(s1 == b64Data.toString("utf-8"))


    var bbdata = "eyJrZXkiOiJrayIsImJvZHkiOiJiYiIsIm1ha2VfdHMiOiIxMjMxMjMxMjMxMTExMjIyMjM0NTU1NTU1NSIsImVkaXRfY291bnQiOjEsImxhc3RfdHMiOiIxMTExIn0"
    var bbdataData = Buffer.from(bbdata, 'base64')
    var bb = JSON.parse(bbdataData.toString("utf-8"));
    console.log(bb)

    const BB = {
        key: 'kk',
        body: 'bb',
        make_ts: '12312312311112222345555555',
        edit_count: 1,
        last_ts: '1111'
    }
    var bbData = JSON.stringify(BB)
    var bbDataData = Buffer.from(bbData, 'utf-8')
    console.log(bbdataData.toString('base64url'))
    console.log(bbdata)
}
b64();