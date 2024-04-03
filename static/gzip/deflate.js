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
	const file = fd.get("file");
	const reader = new FileReader();
	reader.onload = async () => {
		const data = new Uint8Array(reader.result);
		if (go.exited) {
			run()
		}
		const res = gzip.deflate(data);
		if (typeof res === "string") {
			document.getElementById("error").textContent = "GZip Error: " + res;
		}

		const a = document.createElement("a");
		a.href = URL.createObjectURL(new Blob([res]));
		a.download = file.name.replace(/\.gz$/, "");
		a.click();
		URL.revokeObjectURL(a.href);
	}
	reader.readAsArrayBuffer(file);
})

document.getElementById("targz-form").addEventListener("submit", e => {
	e.preventDefault();
	document.getElementById("targz-output").innerHTML = "";
	const fd = new FormData(e.target);
	const file = fd.get("file");
	const reader = new FileReader();
	reader.onload = async () => {
		const data = new Uint8Array(reader.result);
		if (go.exited) {
			run()
		}
		const res = gzip.deflateTarGz(data, appendToOutput);
		if (typeof res === "string") {
			appendToOutput("Error: " + res, undefined);
		}
	}
	reader.readAsArrayBuffer(file);
})

/**
	* @param {string} name
	* @param {Uint8Array | undefined} contents
	*/
function appendToOutput(name, contents) {
	if (typeof contents === "undefined") {
		const p = document.createElement("p");
		p.style.padding = "0.5rem 1rem"
		p.style.color = "red"
		p.textContent = name;
		document.getElementById("targz-output").appendChild(p);
		return
	}

	const div = document.createElement("div");
	div.style.display = "flex"
	div.style.alignItems = "center"
	div.style.justifyContent = "space-between"
	div.style.padding = "0.5rem 1rem"

	const span = document.createElement("span");
	span.textContent = name;
	div.appendChild(span);

	const a = document.createElement("a");
	a.href = URL.createObjectURL(new Blob([contents]));
	a.textContent = "Download";
	a.download = name;
	div.appendChild(a);

	document.getElementById("targz-output").appendChild(div);
}
