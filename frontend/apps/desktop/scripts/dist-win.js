const { spawn } = require("node:child_process");
const { cp, mkdir, readFile, rm, writeFile } = require("node:fs/promises");
const { dirname, resolve } = require("node:path");
const process = require("node:process");

const scriptDir = __dirname;
const appDir = resolve(scriptDir, "..");
const stageDir = resolve(appDir, ".win-pack");
const runtimeDependencies = ["@electron-toolkit/preload", "@electron-toolkit/utils"];
const builderVariable = (name) => `\${${name}}`;

main().catch((error) => {
	console.error(error);
	process.exit(1);
});

async function main() {
	await run("pnpm", ["run", "compile"]);
	await prepareStage();
	await run("electron-builder", [
		"--projectDir",
		stageDir,
		"--config",
		resolve(stageDir, "electron-builder.yml"),
		"--win",
		"--x64",
		"--publish",
		"never",
	]);
}

async function prepareStage() {
	await rm(stageDir, { recursive: true, force: true });
	await mkdir(stageDir, { recursive: true });

	await cp(resolve(appDir, "out"), resolve(stageDir, "out"), { recursive: true });
	await cp(resolve(appDir, "resources"), resolve(stageDir, "resources"), {
		recursive: true,
	});

	for (const dependency of runtimeDependencies) {
		await copyRuntimeDependency(dependency);
	}

	const appPackage = JSON.parse(await readFile(resolve(appDir, "package.json"), "utf8"));
	appPackage.dependencies = Object.fromEntries(
		await Promise.all(
			runtimeDependencies.map(async (dependency) => [
				dependency,
				getPackageVersion(await readPackage(dependency)),
			]),
		),
	);
	delete appPackage.devDependencies;

	await writeFile(resolve(stageDir, "package.json"), `${JSON.stringify(appPackage, null, 2)}\n`);

	const electronPackage = await readPackage("electron");
	await writeFile(
		resolve(stageDir, "electron-builder.yml"),
		[
			"appId: com.leros.desktop",
			"productName: Leros",
			`electronVersion: ${getPackageVersion(electronPackage)}`,
			"icon: resources/icon.png",
			"directories:",
			"  buildResources: build",
			"  output: ../dist",
			"files:",
			"  - out/**/*",
			"  - resources/**/*",
			"  - package.json",
			"asar: true",
			"asarUnpack:",
			"  - resources/**",
			"win:",
			"  signAndEditExecutable: false",
			"  target:",
			"    - target: nsis",
			"      arch:",
			"        - x64",
			`  artifactName: ${builderVariable("productName")}-${builderVariable("version")}-win-${builderVariable("arch")}.${builderVariable("ext")}`,
			"nsis:",
			"  oneClick: false",
			"  perMachine: false",
			"  allowToChangeInstallationDirectory: true",
			"  createDesktopShortcut: true",
			"  createStartMenuShortcut: true",
			"npmRebuild: false",
			"",
		].join("\n"),
	);
}

async function copyRuntimeDependency(packageName) {
	const source = resolve(appDir, "node_modules", ...packageName.split("/"));
	const target = resolve(stageDir, "node_modules", ...packageName.split("/"));

	await mkdir(dirname(target), { recursive: true });
	// 复制真实包内容而不是 pnpm junction，避免打包产物依赖本机 node_modules。
	await cp(source, target, { recursive: true, dereference: true });
}

async function readPackage(packageName) {
	return JSON.parse(
		await readFile(
			resolve(appDir, "node_modules", ...packageName.split("/"), "package.json"),
			"utf8",
		),
	);
}

function getPackageVersion(packageJson) {
	return packageJson.version;
}

function run(command, args) {
	return new Promise((resolvePromise, reject) => {
		const child = spawn(command, args, {
			cwd: appDir,
			stdio: "inherit",
			shell: process.platform === "win32",
		});

		child.on("error", reject);
		child.on("close", (code) => {
			if (code === 0) {
				resolvePromise();
				return;
			}

			reject(new Error(`${command} ${args.join(" ")} exited with code ${code}`));
		});
	});
}
