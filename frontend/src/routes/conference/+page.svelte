<script lang="ts">
    import { API_URL, WEBSOCKET_URL } from "$lib/api/env";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";

    let localStream: MediaStream | undefined = $state(undefined);
    let remoteStreams: MediaStream[] = $state([]);
    let roomId: string = $state("");
    let participants = $state(1);
    let audioMuted = $state(false);
    let videoOff = $state(false);
    let iceFailed = $state(false);
    let ws: WebSocket | undefined;
    let peer: RTCPeerConnection | undefined;
    let callEnded = false;

    function cleanup() {
        if (callEnded) return;
        callEnded = true;
        localStream?.getTracks().forEach(t => t.stop());
        peer?.close();
        ws?.close();
    }

    function removeRemoteStream(streamId: string) {
        const s = remoteStreams.find(s => s.id === streamId);
        if (s) {
            s.getTracks().forEach(t => t.stop());
            remoteStreams = remoteStreams.filter(s => s.id !== streamId);
        }
    }

    function removeRemoteStreamsByPeerId(peerId: string) {
        const prefix = peerId + "-";
        const removed = remoteStreams.filter(s =>
            s.getTracks().some(t => t.id.startsWith(prefix))
        );
        removed.forEach(s => s.getTracks().forEach(t => t.stop()));
        remoteStreams = remoteStreams.filter(s => !removed.includes(s));
    }

    function toggleAudio() {
        if (!localStream) return;
        audioMuted = !audioMuted;
        localStream.getAudioTracks().forEach(t => (t.enabled = !audioMuted));
    }

    function toggleVideo() {
        if (!localStream) return;
        videoOff = !videoOff;
        localStream.getVideoTracks().forEach(t => (t.enabled = !videoOff));
    }

    function leaveCall() {
        ws?.close();
        localStream?.getTracks().forEach(t => t.stop());
        setTimeout(() => {
            if (window.opener) {
                window.close();
            } else {
                goto("/");
            }
        }, 100);
    }

    onMount(async () => {
        try {
            const url = new URL(window.location.href);
            let id = url.searchParams.get("room_id");
            if (id) {
                roomId = id;
            } else {
                const resp = await fetch(`${API_URL}/conference/room`, { method: "POST" });
                if (!resp.ok) throw new Error("Failed to create room: " + resp.status);
                id = await resp.text();
                if (!id) throw new Error("Empty room id");
                roomId = id;
                url.searchParams.set("room_id", id);
                window.history.replaceState({}, "", url.toString());
            }

            const ip = import.meta.env.VITE_WEBRTC_SERVER_IP;

            peer = new RTCPeerConnection({
                iceServers: [
                    { urls: `stun:${ip}:3478` },
                    {
                        urls: [`turn:${ip}:3478`],
                        username: "username",
                        credential: "password",
                    },
                ],
                iceTransportPolicy: "all",
            });

            const pendingIceCandidates: RTCIceCandidateInit[] = [];

            ws = new WebSocket(`${WEBSOCKET_URL}/room/${roomId}`);

            const msgQueue: string[] = [];
            let processing = false;

            async function processMessage(raw: string) {
                const data = JSON.parse(raw);
                try {
                    if (data.type === "offer") {
                        await peer!.setRemoteDescription(new RTCSessionDescription(data.offer));
                        const answer = await peer!.createAnswer();
                        await peer!.setLocalDescription(answer);
                        ws?.send(JSON.stringify({ type: "answer", answer }));
                        for (const c of pendingIceCandidates) {
                            await peer!.addIceCandidate(new RTCIceCandidate(c));
                        }
                        pendingIceCandidates.length = 0;
                    } else if (data.type === "answer") {
                        await peer!.setRemoteDescription(new RTCSessionDescription(data.answer));
                        for (const c of pendingIceCandidates) {
                            await peer!.addIceCandidate(new RTCIceCandidate(c));
                        }
                        pendingIceCandidates.length = 0;
                    } else if (data.type === "candidate") {
                        if (peer!.remoteDescription) {
                            await peer!.addIceCandidate(new RTCIceCandidate(data.candidate));
                        } else {
                            pendingIceCandidates.push(data.candidate);
                        }
                    } else if (data.type === "peer_left" && data.peer_id) {
                        removeRemoteStreamsByPeerId(data.peer_id);
                    }
                } catch (err) {
                    console.error("ws message error:", err, data);
                }
            }

            ws.onmessage = (e) => {
                msgQueue.push(e.data);
                if (!processing) {
                    processing = true;
                    (async () => {
                        while (msgQueue.length) {
                            await processMessage(msgQueue.shift()!);
                        }
                        processing = false;
                    })();
                }
            };

            ws.onerror = () => {
                console.error("WebSocket error");
            };

            ws.onclose = () => {
                if (!callEnded) {
                    cleanup();
                    setTimeout(() => {
                        if (window.opener) {
                            window.close();
                        } else {
                            goto("/");
                        }
                    }, 100);
                }
            };

            await new Promise<void>((resolve, reject) => {
                if (!ws) return reject(new Error("ws not created"));
                const onopen = () => { ws!.onopen = null; resolve(); };
                const onerror = () => { ws!.onerror = null; reject(new Error("WebSocket connection failed")); };
                ws.onopen = onopen;
                ws.onerror = onerror;
            });

            let u_stream: MediaStream;
            try {
                u_stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            } catch (e) {
                console.error("media", e);
                cleanup();
                return;
            }
            localStream = u_stream;

            peer.onicecandidate = (e) => {
                if (e.candidate) {
                    ws?.send(JSON.stringify({ type: "candidate", candidate: e.candidate.toJSON() }));
                }
            };

            peer.oniceconnectionstatechange = () => {
                if (peer!.iceConnectionState === "connected" || peer!.iceConnectionState === "completed") {
                    participants = Math.max(participants, remoteStreams.length + 1);
                } else if (peer!.iceConnectionState === "failed") {
                    iceFailed = true;
                }
            };

            peer.ontrack = (e) => {
                const s = e.streams?.[0];
                if (!s) return;
                if (!remoteStreams.find(existing => existing.id === s.id)) {
                    remoteStreams = [...remoteStreams, s];
                    participants = Math.max(participants, remoteStreams.length + 1);
                }
                s.getTracks().forEach(t => {
                    t.onended = () => {
                        if (!s.getVideoTracks().length && !s.getAudioTracks().length) {
                            removeRemoteStream(s.id);
                        }
                    };
                });
            };

            u_stream.getTracks().forEach((track) => {
                peer!.addTrack(track, u_stream);
            });

            const offer = await peer.createOffer();
            await peer.setLocalDescription(offer);
            ws?.send(JSON.stringify({ type: "offer", offer }));
        } catch (err) {
            console.error("conference setup error:", err);
            cleanup();
        }
    });
</script>

<div class="conference">
    <div class="grid">
        {#if remoteStreams.length > 0}
            {#each remoteStreams as stream (stream.id)}
                <div class="remote-container">
                    <video autoplay playsinline class="remote-video" srcObject={stream}></video>
                </div>
            {/each}
        {:else}
            <div class="remote-container">
                <div class="waiting">
                    {#if iceFailed}
                        <p class="error-text">Connection failed. Please try again.</p>
                        <button onclick={leaveCall} class="retry-btn">Leave</button>
                    {:else}
                        <div class="spinner"></div>
                        <p>Waiting for others to join...</p>
                    {/if}
                </div>
            </div>
        {/if}
        <div class="local-container">
            {#if localStream}
                <video autoplay muted class="local-video" srcObject={localStream}></video>
            {/if}
            <div class="self-label">You</div>
        </div>
    </div>

    <div class="top-bar">
        <div class="room-badge">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
            </svg>
            {roomId}
        </div>
        <div class="participant-count">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
            {participants}
        </div>
    </div>

    <div class="controls">
        <button onclick={toggleAudio} class="control-btn" class:active={!audioMuted} class:inactive={audioMuted} title={audioMuted ? "Unmute mic" : "Mute mic"}>
            {#if audioMuted}
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="1" y1="1" x2="23" y2="23"/>
                    <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
                    <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
                    <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
                </svg>
            {:else}
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
                    <path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
                </svg>
            {/if}
        </button>
        <button onclick={toggleVideo} class="control-btn" class:active={!videoOff} class:inactive={videoOff} title={videoOff ? "Turn on camera" : "Turn off camera"}>
            {#if videoOff}
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="1" y1="1" x2="23" y2="23"/>
                    <path d="M15 10l4.55-2.27A1 1 0 0 1 21 8.64v6.72a1 1 0 0 1-1.45.9L15 14"/>
                    <rect x="1" y="5" width="14" height="14" rx="2" ry="2"/>
                </svg>
            {:else}
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="1" y="5" width="14" height="14" rx="2" ry="2"/>
                    <path d="M15 10l4.55-2.27A1 1 0 0 1 21 8.64v6.72a1 1 0 0 1-1.45.9L15 14"/>
                </svg>
            {/if}
        </button>
        <button onclick={leaveCall} class="control-btn end-call" title="Leave call">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
            </svg>
        </button>
    </div>
</div>

<style>
    .conference {
        position: fixed;
        inset: 0;
        background: #1a1a2e;
        display: flex;
        flex-direction: column;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        color: #fff;
    }

    .grid {
        flex: 1;
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
        gap: 8px;
        padding: 24px;
        position: relative;
        align-content: center;
    }

    .remote-container {
        width: 100%;
        aspect-ratio: 16 / 9;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 12px;
        overflow: hidden;
        background: #16213e;
    }

    .remote-video {
        width: 100%;
        height: 100%;
        object-fit: contain;
    }

    .waiting {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 16px;
        color: #888;
    }

    .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid #333;
        border-top-color: #4f8cff;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    .error-text {
        color: #e74c3c;
        font-size: 14px;
        text-align: center;
    }

    .retry-btn {
        background: #e74c3c;
        color: #fff;
        border: none;
        padding: 8px 24px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 14px;
    }

    .retry-btn:hover {
        background: #c0392b;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .local-container {
        position: absolute;
        bottom: 100px;
        right: 24px;
        width: 200px;
        height: 140px;
        border-radius: 12px;
        overflow: hidden;
        background: #0f3460;
        box-shadow: 0 4px 20px rgba(0,0,0,0.4);
        border: 2px solid rgba(255,255,255,0.15);
    }

    @media (max-width: 767px) {
        .local-container {
            width: 120px;
            height: 90px;
            right: 12px;
            bottom: 80px;
        }
    }

    .local-video {
        width: 100%;
        height: 100%;
        object-fit: cover;
        transform: scaleX(-1);
    }

    .self-label {
        position: absolute;
        bottom: 6px;
        left: 8px;
        font-size: 12px;
        color: rgba(255,255,255,0.7);
        background: rgba(0,0,0,0.5);
        padding: 2px 8px;
        border-radius: 4px;
    }

    .top-bar {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        display: flex;
        align-items: center;
        gap: 16px;
        padding: 16px 24px;
        background: linear-gradient(to bottom, rgba(0,0,0,0.5), transparent);
        pointer-events: none;
        z-index: 10;
    }

    .room-badge, .participant-count {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: rgba(255,255,255,0.7);
        background: rgba(0,0,0,0.4);
        padding: 4px 12px;
        border-radius: 8px;
        pointer-events: auto;
    }

    .room-badge {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 200px;
    }

    @media (max-width: 767px) {
        .room-badge {
            max-width: 120px;
        }
    }

    .controls {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 12px;
        padding: 16px 24px;
        background: #111;
        border-top: 1px solid rgba(255,255,255,0.06);
    }

    .control-btn {
        width: 48px;
        height: 48px;
        border: none;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.15s ease;
        background: #2a2a3e;
        color: #fff;
        -webkit-tap-highlight-color: transparent;
        border-radius: 50%;
    }

    .control-btn:hover {
        background: #3a3a4e;
    }

    .control-btn.inactive {
        background: #e74c3c;
    }

    .control-btn.inactive:hover {
        background: #c0392b;
    }

    .control-btn.end-call {
        background: #e74c3c;
        width: 52px;
        height: 52px;
    }

    .control-btn.end-call:hover {
        background: #c0392b;
    }
</style>
