<script lang="ts">
	import { onMount, tick } from "svelte";

	interface Message {
		msg: string;
		date: string;
		name: string;
		own: boolean;
		id: string;
	}
	let msgs: Message[] = [];
	let msg: string;
	let socket: WebSocket;
	let name: string = "ななし";
	let msgspace: Element;
	let wsUri: string;

	function sendEnterMessage(e) {
		if (e.key === "Enter") {
			sendMessage();
		}
	}

	function sendMessage() {
		if (msg == "") return;
		socket.send(
			JSON.stringify({
				msg: msg,
				date: new Date().toLocaleString(),
				name: name,
			})
		);
		msg = "";
	}

	// WebSocketの接続先判定
	if (location.protocol == "https:") {
		wsUri = "wss://" + location.host + "/ws";
	} else {
		wsUri = "ws://" + location.host + "/ws";
	}

	const scroll = async () => {
		await tick();
		msgspace.scrollTop = msgspace.scrollHeight;
	};

	function connect() {
		socket = new WebSocket(wsUri);
		console.log(socket);
		socket.onopen = () => {
			console.log("Connected to server");
		};
	}

	onMount(() => {
		console.log("onMount");
		connect()
		socket.onmessage = (e) => {
			const data = JSON.parse(e.data);
			msgs.push({
				msg: data.msg,
				date: data.date,
				name: data.name,
				own: data.own,
				id: data.id,
			});
			msgs = msgs
			scroll()
		};
		socket.onclose = () => {
			console.log("接続が切れました。3秒後に再接続します。");
			msgs.push({
				msg: "WebSocket Disconnected",
				date: new Date().toLocaleString(),
				name: "System",
				own: false,
				id: ""
			});
			msgs = msgs
			setTimeout(() => {
				connect();
			}, 3000);
		};
	});

	// webscoketのkeepalive
	setInterval(() => {
		socket.send("ping");
	}, 10000);
</script>

<section class="flex flex-col justify-center w-screen min-h-screen bg-gray-100 text-gray-800 p-10">
	<div class="flex flex-col flex-grow w-full max-w-xl bg-white shadow-xl rounded-lg overflow-hidden">
		<div class="flex flex-col flex-grow h-0 p-4 overflow-auto" bind:this={msgspace}>
			{#each msgs as msg}
				{#if msg.own}
					<div
						class="flex w-full mt-2 space-x-3 max-w-xs ml-auto justify-end">
						<div>
							{msg.name}
							<div
								class="bg-blue-600 text-white p-3 rounded-l-lg rounded-br-lg">
								<p class="text-sm">{msg.msg}</p>
							</div>
							<span class="text-xs text-gray-500 leading-none">{msg.date}</span><br />
							<span class="text-xs text-gray-500 leading-none">{msg.id}</span>
						</div>
					</div>
				{:else}
					<div class="flex w-full mt-2 space-x-3 max-w-xs">
						<div>
							{msg.name}
							<div
								class="bg-gray-300 p-3 rounded-r-lg rounded-bl-lg">
								<p class="text-sm">{msg.msg}</p>
							</div>
							<span class="text-xs text-gray-500 leading-none"
								>{msg.date}</span><br />
							<span class="text-xs text-gray-500 leading-none">{msg.id}</span>
						</div>
					</div>
				{/if}
			{/each}
		</div>

		<div class="bg-gray-300 p-4">
			Name<input class="flex items-center h-10 w-100 rounded px-3 text-sm" type="text" bind:value={name} />
		</div>
		<div class="bg-gray-300 p-4">
			<input class="flex items-center h-10 w-full rounded px-3 text-sm" type="text" placeholder="Type your message…" bind:value={msg} on:keypress={sendEnterMessage}/>
			<button class="flex items-center h-10 w-full rounded px-3 text-sm" on:click={sendMessage}>Send</button>
		</div>
	</div>
</section>

<style>
</style>
