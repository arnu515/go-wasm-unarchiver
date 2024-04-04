const go = new Go();

let mod, inst;

WebAssembly.instantiateStreaming(fetch("../out/main.wasm"), go.importObject).then((result) => {
	mod = result.module;
	inst = result.instance;
	return run();
}).then(() => console.log("WASM: Go module exited.")).catch(console.error)

async function run() {
	await go.run(inst);
	// run resolves after the go program exits
	inst = await WebAssembly.instantiate(mod, go.importObject);
}

document.getElementById("form").addEventListener("submit", e => {
	e.preventDefault();
	const fd = new FormData(e.target);
	const files = fd.getAll("files");
	const arg = []
	const promises = Array.from({length: files.length})
	files.forEach((f, i) => {
		let contents;
		const reader = new FileReader();
		promises[i] = new Promise((resolve, reject) => {
			reader.onload = async () => {
				contents = new Uint8Array(reader.result);
				arg.push({name: f.name, contents});
				resolve();
			}
		})
		reader.readAsArrayBuffer(f);
	})
	Promise.all(promises).then(() => {
		if (go.exited){
			run()
		}
		res = zip.createZip(arg);
		if (typeof res === "string") {
			document.getElementById("error").textContent = res;
			return
		}
		const a = document.createElement("a");
		a.href = URL.createObjectURL(new Blob([res], {type: "application/zip"}));
		a.download = "archive.zip";
		a.click();
		URL.revokeObjectURL(a.href);
	});
})
