const express = require('express');
const fs = require('fs');
const ivm = require('isolated-vm');
const app = express();
app.use(express.json());

const isolate = new ivm.Isolate({
    memoryLimit: 128,
    onCatastrophicError: () => {
        process.abort();
    }
});

app.post('/execute', async (req, res) => {
    console.log("Got new request");

    const { jsFileName, wasmFileName, tests } = req.body;
    console.log(req.body);

    try {
        const code = fs.readFileSync(jsFileName, 'utf8');
        const wasm = fs.readFileSync(wasmFileName).toString('base64');

        console.log("Created new isolate!");
        const context = await isolate.createContext();
        console.log("Created new context!");
        const jail = context.global;
        await jail.set('global', jail.derefInto());
        console.log("Created new jail!");

        const wasmCopy = new ivm.ExternalCopy(wasm).copyInto();
        const testsCopy = new ivm.ExternalCopy(tests).copyInto();

        const instantiateScript = `
        var Module = {};
        let wasmBase64 = "${wasm}";
        let tests = "${tests}";

        function atob(input) {
            const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
            let str = input.replace(/=+$/, '');
            let output = '';

            if (str.length % 4 === 1) {
                throw new Error("'atob' failed: The string to be decoded is not correctly encoded.");
            }
            for (let bc = 0, bs = 0, buffer, i = 0; buffer = str.charAt(i++); ~buffer && (bs = bc % 4 ? bs * 64 + buffer : buffer, bc++ % 4) ? output += String.fromCharCode(255 & bs >> (-2 * bc & 6)) : 0) {
                buffer = chars.indexOf(buffer);
            }
            return output;
        }

        function base64ToUint8Array(base64) {
            const binaryString = atob(base64);
            const length = binaryString.length;
            const bytes = new Uint8Array(length);
            for (let i = 0; i < length; i++) {
                bytes[i] = binaryString.charCodeAt(i);
            }
            return bytes;
        }

        logInfo("Inside isolate, setting up WASM!");
        Module.instantiateWasm = (info, receiveInstance) => {
            const wasmBinary = base64ToUint8Array(wasmBase64);
            WebAssembly.instantiate(wasmBinary, info)
                .then(obj => receiveInstance(obj.instance))
                .catch(err => { throw err; });
            logInfo("WASM ready!");
            return {};
        };
        `;

        let outputs = "";
        let errors = "";

        jail.setSync('logInfo', function (...args) {
            console.log("ISOLATE-INFO: ", ...args);
        });
        jail.setSync('logOutput', function (arg) {
            console.log("Output!", arg);
            outputs += arg + '\n';
        });
        jail.setSync('logError', function (arg) {
            console.log("Error!", arg);
            errors += arg + '\n';
        });

        const instantiatedScript = await isolate.compileScript(instantiateScript);
        console.log("Instantiated script!");
        await instantiatedScript.run(context, [wasmCopy, testsCopy]);
        console.log("Ran instantiated script!");

        const script = await isolate.compileScript(code);
        console.log("Compiled script!");
        await script.run(context);
        console.log("Ran main script!");

        res.status(200).json({ outputs, errors });
    } catch (err) {
        console.error(err.message);
        res.status(500).json({ error: err.message });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Task-runner service is running on port ${PORT}`);
});