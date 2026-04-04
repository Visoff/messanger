<script lang="ts">
    import { API_URL, API_URL_WITH_PROTOCOL } from "$lib/api/env";
    import { onMount } from "svelte";

    let stream: MediaStream | undefined = $state(undefined);
    let remoteStream: MediaStream | undefined = $state();

    onMount(async () => {
        const url = new URL(window.location.href);
        let room_id = url.searchParams.get("room_id");
        if (!room_id) {
            const resp = await fetch(`${API_URL}/conference/room`, {
                method: "POST",
            });
            room_id = await resp.text();
            console.log("created room", room_id);
            url.searchParams.set("room_id", room_id);
            window.history.replaceState({}, "", url.toString());
        }
        console.log("room_id", room_id);
        const ws = new WebSocket(
            `${API_URL_WITH_PROTOCOL("ws://", "wss://")}/conference/room/${room_id}`,
        );
        await new Promise((resolve) => {
            ws.onopen = resolve;
        });
        const u_stream = await navigator.mediaDevices.getUserMedia({
            video: true,
        });
        stream = u_stream;

        const peer = new RTCPeerConnection({
            iceServers: [
                {
                    urls: ["turn:localhost:3478"],
                    username: "username",
                    credential: "password",
                },
            ],
            iceTransportPolicy: "all",
        });

        let IceCandidates: RTCIceCandidate[] = [];

        ws.onmessage = async (e) => {
            const data = JSON.parse(e.data);
            console.log(data);
            if (data.type == "offer") {
                peer.setRemoteDescription(
                    new RTCSessionDescription(data.offer),
                );
                peer.createAnswer().then((a) => {
                    peer.setLocalDescription(a);
                    ws.send(
                        JSON.stringify({
                            type: "answer",
                            answer: a,
                        }),
                    );
                });
                for (const c of IceCandidates) {
                    await peer.addIceCandidate(c);
                }
            } else if (data.type == "answer") {
                peer.setRemoteDescription(
                    new RTCSessionDescription(data.answer),
                );
                for (const c of IceCandidates) {
                    await peer.addIceCandidate(c);
                }
            } else if (data.type == "candidate") {
                if (peer.remoteDescription) {
                    await peer.addIceCandidate(
                        new RTCIceCandidate(data.candidate),
                    );
                } else {
                    IceCandidates.push(new RTCIceCandidate(data.candidate));
                }
            }
        };

        peer.onicecandidate = (e) => {
            if (e.candidate) {
                ws.send(
                    JSON.stringify({
                        type: "candidate",
                        candidate: e.candidate,
                    }),
                );
            }
        };

        peer.oniceconnectionstatechange = () => {
            console.log("ICE connection state:", peer.iceConnectionState);
            if (peer.iceConnectionState === "failed") {
                console.error("ICE failed, gathering stats...");
                peer.getStats().then((stats) => {
                    stats.forEach((report) => {
                        if (
                            report.type === "ice-candidate" &&
                            report.candidateType === "relay"
                        ) {
                            console.log("Relay candidate:", report);
                        }
                    });
                });
            }
        };

        peer.onnegotiationneeded = () => {
            peer.createOffer().then((o) => {
                peer.setLocalDescription(o);
                ws.send(
                    JSON.stringify({
                        type: "offer",
                        offer: o,
                    }),
                );
            });
        };

        peer.ontrack = (e) => {
            if (!remoteStream) {
                remoteStream = new MediaStream();
            }
            remoteStream.addTrack(e.track);
            console.log(remoteStream);
        };

        ws.onclose = () => {
            peer.close();
        };

        u_stream.getTracks().forEach((track) => {
            peer.addTrack(track, u_stream);
        });
    });
</script>

{#if stream}
    <video autoplay srcObject={stream}></video>
{/if}
{#if remoteStream}
    <video autoplay srcObject={remoteStream}></video>
{/if}
