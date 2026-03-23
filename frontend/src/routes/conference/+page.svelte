<script lang="ts">
    import { API_URL } from "$lib/api/env";
    import { onMount } from "svelte";

    let stream: MediaStream | undefined = $state(undefined);
    let remoteStream: MediaStream | undefined = $state();

    onMount(async () => {
        remoteStream = new MediaStream();
        const url = new URL(window.location.href);
        let room_id = url.searchParams.get("room_id")
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
        let ws_url = new URL(`${API_URL}/conference/room/${room_id}`);
        if (ws_url.protocol == "https:") {
            ws_url.protocol = "wss:";
        } else if (ws_url.protocol == "http:") {
            ws_url.protocol = "ws:";
        }
        const ws = new WebSocket(ws_url.toString());
        await new Promise((resolve) => {
            ws.onopen = resolve;
        })
        const u_stream = await navigator.mediaDevices.getUserMedia({
            video: true,
        });
        stream = u_stream;

        const peer = new RTCPeerConnection({
            iceServers: [
                {
                    urls: "stun:stun.l.google.com:19302",
                },
            ],
        });

        ws.onmessage = (e) => {
            const data = JSON.parse(e.data);
            if (data.type == "offer") {
                peer.setRemoteDescription(new RTCSessionDescription(data));
                u_stream.getTracks().forEach((track) => {
                    peer.addTrack(track, u_stream);
                });
                peer.createAnswer().then((a) => {
                    peer.setLocalDescription(a);
                    ws.send(JSON.stringify({
                        type: "answer",
                        answer: a,
                    }));
                });
            } else if (data.type == "answer") {
                peer.setRemoteDescription(new RTCSessionDescription(data));
            } else if (data.type == "candidate") {
                peer.addIceCandidate(new RTCIceCandidate(data.candidate));
            }
        }

        peer.onicecandidate = (e) => {
            if (e.candidate) {
                ws.send(JSON.stringify({
                    type: "candidate",
                    candidate: e.candidate,
                }));
            }
        };

        peer.onnegotiationneeded = () => {
            peer.createOffer().then((o) => {
                peer.setLocalDescription(o);
                ws.send(JSON.stringify({
                    type: "offer",
                    offer: o,
                }));
            });
        }

        peer.ontrack = (e) => {
            if (!remoteStream) {
                remoteStream = new MediaStream();
            }
            remoteStream.addTrack(e.track);
        };
    })
</script>

{#if stream}
    <video autoplay srcobject={stream}></video>
    <video autoplay srcobject={remoteStream}></video>
{/if}
