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
        await new Promise((resolve, reject) => {
            ws.onopen = resolve;
            ws.onerror = reject;
        });
        let u_stream;
        try {
            u_stream = await navigator.mediaDevices.getUserMedia({
                video: true,
            });
        } catch (e) {
            console.error("media", e);
            return;
        }
        stream = u_stream;

        //const remote_ip = "192.168.100.6";
        const remote_ip = "158.160.209.129";

        const peer = new RTCPeerConnection({
            iceServers: [
                { urls: `stun:${remote_ip}:3478` },
                {
                    urls: [`turn:${remote_ip}:3478`],
                    username: "username",
                    credential: "password",
                },
            ],
            iceTransportPolicy: "relay",
        });

        const iceCandidates: RTCIceCandidate[] = [];

        ws.onmessage = async (e) => {
            const data = JSON.parse(e.data);
            console.log(data);
            if (data.type == "offer") {
                console.error("got offer");
            } else if (data.type == "answer") {
                peer.setRemoteDescription(
                    new RTCSessionDescription(data.answer),
                );
                for (const candidate of iceCandidates) {
                    await peer.addIceCandidate(candidate);
                }
                iceCandidates.length = 0;
            } else if (data.type == "candidate") {
                if (peer.remoteDescription) {
                    await peer.addIceCandidate(
                        new RTCIceCandidate(data.candidate),
                    );
                } else {
                    iceCandidates.push(
                        new RTCIceCandidate(data.candidate),
                    );
                }
            }
        };

        peer.onicecandidate = (e) => {
            if (e.candidate && e.candidate.address) {
                console.log("send candidate", e.candidate);
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
                            report.type === "ice-candidate"
                        ) {
                            console.log("candidate:", report);
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
            console.log("ws closed");
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
